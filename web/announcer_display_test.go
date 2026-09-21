// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnnouncerDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/announcer?displayId=1")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Announcer Display - Untitled Event - Cheesy Arena")
}

func TestAnnouncerDisplayMatchLoad(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{Type: model.Playoff, Red1: 254, Red2: 1114, Blue3: 2056}
	web.arena.LoadMatch(&match)

	recorder := web.getHttpResponse("/displays/announcer/match_load")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "254")
	assert.Contains(t, recorder.Body.String(), "1114")
	assert.Contains(t, recorder.Body.String(), "2056")
}

func TestAnnouncerDisplayMatchLoadTwoVsTwo(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.Database.CreateTeam(&model.Team{Id: 254})
	web.arena.Database.CreateTeam(&model.Team{Id: 1114})
	web.arena.Database.CreateTeam(&model.Team{Id: 118})
	web.arena.Database.CreateTeam(&model.Team{Id: 148})
	match := model.Match{Type: model.Practice, Red1: 254, Red2: 1114, Blue1: 118, Blue2: 148}
	web.arena.LoadMatch(&match)

	recorder := web.getHttpResponse("/displays/announcer/match_load")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No team present")

	web.arena.EventSettings.TwoVsTwoMode = true
	recorder = web.getHttpResponse("/displays/announcer/match_load")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "No team present")
}

func TestAnnouncerDisplayScorePosted(t *testing.T) {
	web := setupTestWeb(t)
	for _, test := range []struct {
		status game.MatchStatus
		winner string
		class  string
	}{
		{game.RedWonMatch, "Red", "bg-danger"},
		{game.BlueWonMatch, "Blue", "bg-primary"},
		{game.TieMatch, "Tie", "bg-tie"},
	} {
		match := model.Match{Type: model.Qualification, LongName: "Qual 17", Status: test.status}
		web.arena.SavedMatch = &match

		recorder := web.getHttpResponse("/displays/announcer/score_posted")
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Qual 17")
		assert.Contains(t, recorder.Body.String(), "Winner: "+test.winner)
		assert.Contains(t, recorder.Body.String(), test.class)
	}
}

// Verifies that the score-posted template renders the game's own vocabulary (not just the match name), including
// the crown location and the ranking points, for a realistic score.
func TestAnnouncerDisplayScorePostedGameFields(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{Type: model.Qualification, LongName: "Qual 17", Status: game.RedWonMatch}
	web.arena.SavedMatch = &match
	web.arena.SavedMatchResult = &model.MatchResult{
		RedScore: &game.Score{
			AutoFirst:           1,
			TeleopTop:           2,
			Crown:               game.CrownTeleopTop,
			LeaveStatuses:       [3]bool{true, true, false},
			AutoBalanceStatuses: [3]bool{true, false, false},
			EndgameStatuses:     [3]game.EndgameStatus{game.EndgameBalance, game.EndgamePark, game.EndgameNone},
			Toss:                true,
		},
		BlueScore: &game.Score{},
		RedCards:  map[string]string{},
		BlueCards: map[string]string{},
	}

	recorder := web.getHttpResponse("/displays/announcer/score_posted")
	assert.Equal(t, 200, recorder.Code)
	body := recorder.Body.String()
	assert.Contains(t, body, "Auto Treasure")
	assert.Contains(t, body, "Teleop Top +10") // The crown's location and bonus.
	assert.Contains(t, body, "Endgame RP")
	assert.Contains(t, body, "Yes") // Red balanced a robot, so it earns the Endgame RP.
}

func TestAnnouncerDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/announcer/websocket?displayId=1", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "displayConfiguration")
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "scorePosted")

	web.arena.MatchLoadNotifier.Notify()
	readWebsocketType(t, ws, "matchLoad")
	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	web.arena.StartMatch()
	web.arena.Update()
	messages := readWebsocketMultiple(t, ws, 3)
	_, ok := messages["audienceDisplayMode"]
	assert.True(t, ok)
	_, ok = messages["eventStatus"]
	assert.True(t, ok)
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	web.arena.RealtimeScoreNotifier.Notify()
	readWebsocketType(t, ws, "realtimeScore")
	web.arena.ScorePostedNotifier.Notify()
	readWebsocketType(t, ws, "scorePosted")
}
