// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/scoring/invalidposition")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid position")

	// All six positions should render, each with the controls appropriate to it.
	for _, position := range []string{"red", "blue", "red_near", "red_far", "blue_near", "blue_far"} {
		recorder = web.getHttpResponse("/panels/scoring/" + position)
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
		body := recorder.Body.String()

		parameters := positionParameters[position]
		if parameters.ShowsNear() {
			assert.Contains(t, body, "Auto Floor")
			assert.Contains(t, body, "Stacked")
			assert.Contains(t, body, "crown-teleop_top")
		} else {
			assert.NotContains(t, body, "Auto Floor")
		}
		if parameters.ShowsFar() {
			assert.Contains(t, body, "The Toss")
			assert.Contains(t, body, "Balance")
		} else {
			assert.NotContains(t, body, "The Toss")
		}
	}
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.TwoVsTwoMode = false

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)

	// Connect the near and far panels for each alliance; both register under the alliance, not the position.
	redNearConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_near/websocket", nil)
	assert.Nil(t, err)
	defer redNearConn.Close()
	redNearWs := websocket.NewTestWebsocket(redNearConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))

	redFarConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_far/websocket", nil)
	assert.Nil(t, err)
	defer redFarConn.Close()
	redFarWs := websocket.NewTestWebsocket(redFarConn)
	assert.Equal(t, 2, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))

	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))

	// Should get a few status updates right after connection.
	for _, ws := range []*websocket.Websocket{redNearWs, redFarWs, blueWs} {
		readWebsocketType(t, ws, "resetLocalState")
		readWebsocketType(t, ws, "matchLoad")
		readWebsocketType(t, ws, "matchTime")
		readWebsocketType(t, ws, "realtimeScore")
	}

	// Exercise each websocket command and confirm it changes the realtime score.
	redNearWs.Write("treasure", struct {
		Counter    string
		Adjustment int
	}{"auto_top", 1})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, 1, web.arena.RedRealtimeScore.CurrentScore.AutoTop)

	// The counter should clamp at zero and not go negative.
	redNearWs.Write("treasure", struct {
		Counter    string
		Adjustment int
	}{"auto_top", -5})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.AutoTop)

	// An unrecognized counter name should be rejected without a notification.
	redNearWs.Write("treasure", struct {
		Counter    string
		Adjustment int
	}{"nonexistent", 1})

	redNearWs.Write("crown", struct{ Value string }{"teleop_top"})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.CrownTeleopTop, web.arena.RedRealtimeScore.CurrentScore.Crown)

	redFarWs.Write("leave", struct{ TeamPosition int }{1})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.True(t, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses[0])
	// Sending it again toggles it back off.
	redFarWs.Write("leave", struct{ TeamPosition int }{1})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.False(t, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses[0])

	redFarWs.Write("auto_balance", struct{ TeamPosition int }{2})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.True(t, web.arena.RedRealtimeScore.CurrentScore.AutoBalanceStatuses[1])

	redFarWs.Write("endgame", struct {
		TeamPosition int
		Value        int
	}{3, int(game.EndgameBalance)})
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.EndgameBalance, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[2])

	// Reject an out-of-range robot position (there is no station 4) and an out-of-range endgame value.
	redFarWs.Write("leave", struct{ TeamPosition int }{4})
	redFarWs.Write("endgame", struct {
		TeamPosition int
		Value        int
	}{1, 3})

	redFarWs.Write("toss", nil)
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.True(t, web.arena.RedRealtimeScore.CurrentScore.Toss)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redNearWs.Write("invalid", nil)

	// Test committing logic; the alliance isn't ready until both its near and far panels have committed.
	redNearWs.Write("commitMatch", nil)
	readWebsocketType(t, redNearWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
	web.arena.MatchState = field.PostMatch

	redNearWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.False(t, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red") >= web.arena.ScoringPanelRegistry.GetNumPanels("red"))

	redFarWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 2, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redNearWs, "matchLoad")
	readWebsocketType(t, redNearWs, "realtimeScore")
	readWebsocketType(t, redFarWs, "matchLoad")
	readWebsocketType(t, redFarWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
}

func TestScoringPanelWebsocketRejectsThirdRobotInTwoVsTwoMode(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.TwoVsTwoMode = true

	server, wsUrl := web.startTestServer()
	defer server.Close()
	redFarConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_far/websocket", nil)
	assert.Nil(t, err)
	defer redFarConn.Close()
	redFarWs := websocket.NewTestWebsocket(redFarConn)
	readWebsocketType(t, redFarWs, "resetLocalState")
	readWebsocketType(t, redFarWs, "matchLoad")
	readWebsocketType(t, redFarWs, "matchTime")
	readWebsocketType(t, redFarWs, "realtimeScore")

	// Station 3 is not in play in 2v2 mode, so the command should be silently rejected.
	redFarWs.Write("leave", struct{ TeamPosition int }{3})
	redFarWs.Write("leave", struct{ TeamPosition int }{1})
	readWebsocketType(t, redFarWs, "realtimeScore")
	assert.False(t, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses[2])
	assert.True(t, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses[0])
}
