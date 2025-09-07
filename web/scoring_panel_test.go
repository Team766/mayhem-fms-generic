// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
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
	recorder = web.getHttpResponse("/panels/scoring/red_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/red_far")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_far")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_near/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))
	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue_near/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "resetLocalState")
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "resetLocalState")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// Send some autonomous period scoring commands.
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)
	leaveData := struct {
		TeamPosition int
	}{}
	web.arena.MatchState = field.AutoPeriod
	leaveData.TeamPosition = 1
	redWs.Write("leave", leaveData)
	leaveData.TeamPosition = 3
	redWs.Write("leave", leaveData)
	for i := 0; i < 2; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [3]bool{true, false, true}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)
	redWs.Write("leave", leaveData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, [3]bool{true, false, false}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)

	// Send some counter scoring commands
	counterData := struct {
		Adjustment int
		Current    bool
		Autonomous bool
		NearSide   bool
	}{}
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)
	counterData.Adjustment = 1
	blueWs.Write("barge", counterData)
	blueWs.Write("barge", counterData)
	blueWs.Write("barge", counterData)
	counterData.Adjustment = -1
	blueWs.Write("barge", counterData)
	blueWs.Write("barge", counterData)
	counterData.Adjustment = 1
	blueWs.Write("barge", counterData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	counterData.Adjustment = -1
	redWs.Write("processor", counterData)
	redWs.Write("processor", counterData)
	counterData.Adjustment = 1
	redWs.Write("processor", counterData)
	redWs.Write("processor", counterData)
	redWs.Write("processor", counterData)
	counterData.Adjustment = -1
	redWs.Write("processor", counterData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 2, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)

	// Send some gamepiece scoring commands
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count)
	counterData.Adjustment = 1
	counterData.Current = true
	counterData.Autonomous = true
	counterData.NearSide = true
	redWs.Write("trough", counterData)
	redWs.Write("trough", counterData)
	redWs.Write("trough", counterData)
	counterData.Adjustment = -1
	redWs.Write("trough", counterData)
	for i := 0; i < 4; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	counterData.Autonomous = false
	counterData.Adjustment = 1
	redWs.Write("trough", counterData)
	redWs.Write("trough", counterData)
	counterData.Current = false
	counterData.Autonomous = true
	redWs.Write("trough", counterData)
	counterData.NearSide = false
	redWs.Write("trough", counterData)
	counterData.Adjustment = -1
	counterData.Current = true
	counterData.Autonomous = false
	redWs.Write("trough", counterData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 4, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count)
	assert.Equal(t, 3, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count)
	assert.Equal(t, 1, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count)

	// Send some gamepiece scoring commands
	gamepieceData := struct {
		GamepieceType  int
		GamepieceLevel int
		Current        bool
		Autonomous     bool
	}{}
	// Auto phase
	gamepieceData.GamepieceType = 1  // Gamepiece1
	gamepieceData.GamepieceLevel = 1 // Level1
	gamepieceData.Current = false
	gamepieceData.Autonomous = true
	redWs.Write("gamepiece", gamepieceData)
	gamepieceData.GamepieceType = 2  // Gamepiece2
	redWs.Write("gamepiece", gamepieceData)
	gamepieceData.GamepieceLevel = 2 // Level2
	redWs.Write("gamepiece", gamepieceData)
	redWs.Write("gamepiece", gamepieceData)
	// Teleop phase
	gamepieceData.GamepieceType = 1  // Gamepiece1
	gamepieceData.GamepieceLevel = 2 // Level2
	redWs.Write("gamepiece", gamepieceData)
	// Blue alliance
	gamepieceData.Current = true
	gamepieceData.Autonomous = false
	gamepieceData.GamepieceType = 1  // Gamepiece1
	gamepieceData.GamepieceLevel = 1 // Level1
	blueWs.Write("gamepiece", gamepieceData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	// Red Auto
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece2Count,
	)
	// Red Current
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count,
	)
	// Blue Auto
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece2Count,
	)
	// Blue Current
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0,
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count,
	)

	// Send some park status commands
	parkData := struct {
		TeamPosition int
		ParkStatus   bool
	}{}
	assert.Equal(
		t,
		[3]bool{false, false, false},
		web.arena.RedRealtimeScore.CurrentScore.ParkStatuses,
	)
	assert.Equal(
		t,
		[3]bool{false, false, false},
		web.arena.BlueRealtimeScore.CurrentScore.ParkStatuses,
	)
	parkData.TeamPosition = 1
	parkData.ParkStatus = true
	redWs.Write("park", parkData)
	parkData.TeamPosition = 2
	parkData.ParkStatus = true
	blueWs.Write("park", parkData)
	parkData.TeamPosition = 3
	parkData.ParkStatus = true
	blueWs.Write("park", parkData)
	parkData.TeamPosition = 3
	parkData.ParkStatus = true
	redWs.Write("park", parkData)
	parkData.TeamPosition = 3
	parkData.ParkStatus = true
	redWs.Write("park", parkData)
	parkData.TeamPosition = 2
	parkData.ParkStatus = false
	redWs.Write("park", parkData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(
		t,
		[3]bool{true, false, true},
		web.arena.RedRealtimeScore.CurrentScore.ParkStatuses,
	)
	assert.Equal(
		t,
		[3]bool{false, true, true},
		web.arena.BlueRealtimeScore.CurrentScore.ParkStatuses,
	)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)
	leaveData.TeamPosition = 0
	redWs.Write("leave", leaveData)
	counterData.Current = false
	counterData.Autonomous = false
	redWs.Write("trough", counterData)
	parkData.TeamPosition = 1
	parkData.ParkStatus = true
	blueWs.Write("park", parkData)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
}
