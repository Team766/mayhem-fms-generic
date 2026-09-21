// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

type ScoringPosition struct {
	Title    string
	Alliance string

	// Side is "near" or "far" for a single-side panel, or "" for a combined panel that shows both sides.
	Side string
}

// ShowsNear returns true if this position's panel should show the NEAR (shelf scorer) controls.
func (position ScoringPosition) ShowsNear() bool {
	return position.Side == "" || position.Side == "near"
}

// ShowsFar returns true if this position's panel should show the FAR (robot scorer) controls.
func (position ScoringPosition) ShowsFar() bool {
	return position.Side == "" || position.Side == "far"
}

var positionParameters = map[string]ScoringPosition{
	"red": {
		Title:    "Red",
		Alliance: "red",
	},
	"blue": {
		Title:    "Blue",
		Alliance: "blue",
	},
	"red_near": {
		Title:    "Red Near",
		Alliance: "red",
		Side:     "near",
	},
	"red_far": {
		Title:    "Red Far",
		Alliance: "red",
		Side:     "far",
	},
	"blue_near": {
		Title:    "Blue Near",
		Alliance: "blue",
		Side:     "near",
	},
	"blue_far": {
		Title:    "Blue Far",
		Alliance: "blue",
		Side:     "far",
	},
}

// The dragon's crown placements, keyed by the id used in the "crown" websocket command.
var crownPlacementsByName = map[string]game.CrownPlacement{
	"none":           game.CrownNone,
	"auto_floor":     game.CrownAutoFloor,
	"auto_first":     game.CrownAutoFirst,
	"auto_top":       game.CrownAutoTop,
	"teleop_floor":   game.CrownTeleopFloor,
	"teleop_first":   game.CrownTeleopFirst,
	"teleop_top":     game.CrownTeleopTop,
	"teleop_stacked": game.CrownTeleopStacked,
}

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	parameters, ok := positionParameters[position]
	if !ok {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
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
	parameters, ok := positionParameters[position]
	if !ok {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
		return
	}
	alliance := parameters.Alliance

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
	defer closeWebsocket(ws)

	// Register under the alliance (rather than the near/far position) so that the commit-readiness count covers
	// both scoring positions for the alliance; the score isn't ready to post until both have committed.
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Instruct panel to clear any local state in case this is a reconnect
	writeWebsocketMessage(ws, "resetLocalState", nil)

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
				writeWebsocketError(ws, "Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if command == "treasure" {
			args := struct {
				Counter    string
				Adjustment int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}

			if counter := treasureCounter(score, args.Counter); counter != nil {
				*counter += args.Adjustment
				if *counter < 0 {
					*counter = 0
				}
				scoreChanged = true
			}
		} else if command == "crown" {
			args := struct {
				Value string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}

			if placement, ok := crownPlacementsByName[args.Value]; ok {
				score.Crown = placement
				scoreChanged = true
			}
		} else if command == "leave" {
			args := struct {
				TeamPosition int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}

			if isValidRobotPosition(args.TeamPosition, web.arena.EventSettings.TwoVsTwoMode) {
				score.LeaveStatuses[args.TeamPosition-1] = !score.LeaveStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else if command == "auto_balance" {
			args := struct {
				TeamPosition int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}

			if isValidRobotPosition(args.TeamPosition, web.arena.EventSettings.TwoVsTwoMode) {
				score.AutoBalanceStatuses[args.TeamPosition-1] = !score.AutoBalanceStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else if command == "endgame" {
			args := struct {
				TeamPosition int
				Value        int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}

			if isValidRobotPosition(args.TeamPosition, web.arena.EventSettings.TwoVsTwoMode) &&
				args.Value >= int(game.EndgameNone) && args.Value <= int(game.EndgameBalance) {
				score.EndgameStatuses[args.TeamPosition-1] = game.EndgameStatus(args.Value)
				scoreChanged = true
			}
		} else if command == "toss" {
			score.Toss = !score.Toss
			scoreChanged = true
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Returns a pointer to the Score counter field named by the given id, or nil if the name is not recognized.
func treasureCounter(score *game.Score, name string) *int {
	switch name {
	case "auto_floor":
		return &score.AutoFloor
	case "auto_first":
		return &score.AutoFirst
	case "auto_top":
		return &score.AutoTop
	case "teleop_floor":
		return &score.TeleopFloor
	case "teleop_first":
		return &score.TeleopFirst
	case "teleop_top":
		return &score.TeleopTop
	case "teleop_stacked":
		return &score.TeleopStacked
	default:
		return nil
	}
}

// Returns true if the given 1-indexed robot position is valid for the current alliance size.
func isValidRobotPosition(teamPosition int, twoVsTwoMode bool) bool {
	if teamPosition < 1 || teamPosition > 3 {
		return false
	}
	if twoVsTwoMode && teamPosition == 3 {
		return false
	}
	return true
}
