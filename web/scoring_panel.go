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
	"strings"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
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
		ScoresEndgame:    false,
		ScoresStructure1: true,
		ScoresStructure2: false,
	},
	"red_far": {
		Title:            "Red Far",
		Alliance:         "red",
		ScoresAuto:       false,
		ScoresEndgame:    true,
		ScoresStructure1: false,
		ScoresStructure2: true,
	},
	"blue_near": {
		Title:            "Blue Near",
		Alliance:         "blue",
		ScoresAuto:       true,
		ScoresEndgame:    false,
		ScoresStructure1: true,
		ScoresStructure2: false,
	},
	"blue_far": {
		Title:            "Blue Far",
		Alliance:         "blue",
		ScoresAuto:       false,
		ScoresEndgame:    true,
		ScoresStructure1: false,
		ScoresStructure2: true,
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
		handleWebErr(w, fmt.Errorf("Invalid position '%s'", position))
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
				score.Mayhem.LeaveStatuses[args.TeamPosition-1] = !score.Mayhem.LeaveStatuses[args.TeamPosition-1]
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
				score.Mayhem.ParkStatuses[args.TeamPosition-1] = !score.Mayhem.ParkStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else if command == "GP1" {
			args := struct {
				Level      int
				Autonomous bool
				Adjustment int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.Level >= 1 && args.Level <= 3 {
				if args.Autonomous {
					switch args.Level {
					case 1:
						// Add the adjustment and ensure we don't go below zero
						score.Mayhem.AutoGamepiece1Level1Count += args.Adjustment
						if score.Mayhem.AutoGamepiece1Level1Count < 0 {
							score.Mayhem.AutoGamepiece1Level1Count = 0
						}
						scoreChanged = true
					case 2:
						// Add the adjustment and ensure we don't go below zero
						score.Mayhem.AutoGamepiece1Level2Count += args.Adjustment
						if score.Mayhem.AutoGamepiece1Level2Count < 0 {
							score.Mayhem.AutoGamepiece1Level2Count = 0
						}
						scoreChanged = true
					}
				} else {
					switch args.Level {
					case 1:
						// Add the adjustment and ensure we don't go below zero
						score.Mayhem.TeleopGamepiece1Level1Count += args.Adjustment
						if score.Mayhem.TeleopGamepiece1Level1Count < 0 {
							score.Mayhem.TeleopGamepiece1Level1Count = 0
						}
						scoreChanged = true
					case 2:
						// Add the adjustment and ensure we don't go below zero
						score.Mayhem.TeleopGamepiece1Level2Count += args.Adjustment
						if score.Mayhem.TeleopGamepiece1Level2Count < 0 {
							score.Mayhem.TeleopGamepiece1Level2Count = 0
						}
						scoreChanged = true
					}
				}
			}
		} else if command == "GP2" {
			args := struct {
				Autonomous bool
				Adjustment int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if args.Autonomous {
				// Add the adjustment and ensure we don't go below zero
				score.Mayhem.AutoGamepiece2Count += args.Adjustment
				if score.Mayhem.AutoGamepiece2Count < 0 {
					score.Mayhem.AutoGamepiece2Count = 0
				}
				scoreChanged = true
			} else {
				// Add the adjustment and ensure we don't go below zero
				score.Mayhem.TeleopGamepiece2Count += args.Adjustment
				if score.Mayhem.TeleopGamepiece2Count < 0 {
					score.Mayhem.TeleopGamepiece2Count = 0
				}
				scoreChanged = true
			}
		} else if command == "addFoul" {
			args := struct {
				Alliance string
				IsMajor  bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := game.Foul{IsMajor: args.IsMajor}
			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			web.arena.RealtimeScoreNotifier.Notify()
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}
