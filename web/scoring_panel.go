// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

type ScoringPosition struct {
	Title            string
	Alliance         string
	ScoresAuto       bool
	ScoresEndgame    bool
	ScoresStructure1 bool
	ScoresStructure2 bool
}

var positionParameters = map[string]ScoringPosition{
	"red_near": {
		Title:            "Red Near",
		Alliance:         "red",
		ScoresAuto:       true,
		ScoresEndgame:    true,
		ScoresStructure1: true,
		ScoresStructure2: false,
	},
	"red_far": {
		Title:            "Red Far",
		Alliance:         "red",
		ScoresAuto:       false,
		ScoresEndgame:    false,
		ScoresStructure1: false,
		ScoresStructure2: true,
	},
	"blue_near": {
		Title:            "Blue Near",
		Alliance:         "blue",
		ScoresAuto:       false,
		ScoresEndgame:    false,
		ScoresStructure1: false,
		ScoresStructure2: true,
	},
	"blue_far": {
		Title:            "Blue Far",
		Alliance:         "blue",
		ScoresAuto:       true,
		ScoresEndgame:    true,
		ScoresStructure1: true,
		ScoresStructure2: false,
	},
}

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	parameters, ok := positionParameters[position]
	if !ok {
		handleWebErr(w, fmt.Errorf("invalid position '%s'", position))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PositionName string
		Position     ScoringPosition
	}{web.arena.EventSettings, position, parameters}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	if position != "red_near" && position != "red_far" && position != "blue_near" && position != "blue_far" {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
		return
	}
	alliance := strings.Split(position, "_")[0]

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(position, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(position, ws)

	// Instruct panel to clear any local state in case this is a reconnect
	ws.Write("resetLocalState", nil)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier,
	)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}
		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(position, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if command == "leave" {
			args := struct {
				TeamPosition int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
				score.LeaveStatuses[args.TeamPosition-1] = !score.LeaveStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else if command == "park" {
			args := struct {
				TeamPosition int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
				score.ParkStatuses[args.TeamPosition-1] = !score.ParkStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else if command == "updateScore" {
			args := struct {
				Field      string
				Adjustment int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Use reflection to update the given field in the score.
			field := reflect.ValueOf(score).Elem().FieldByName(args.Field)
			if field.IsValid() && field.CanSet() && field.Kind() == reflect.Int {
				newValue := field.Int() + int64(args.Adjustment)
				if newValue >= 0 {
					field.SetInt(newValue)
					scoreChanged = true
				}
			}
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}
