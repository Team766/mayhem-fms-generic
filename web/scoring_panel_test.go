// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.Mayhem.LeaveStatuses)
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
	assert.Equal(t, [3]bool{true, false, true}, web.arena.RedRealtimeScore.CurrentScore.Mayhem.LeaveStatuses)
	redWs.Write("leave", leaveData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, [3]bool{true, false, false}, web.arena.RedRealtimeScore.CurrentScore.Mayhem.LeaveStatuses)

	// Send some counter scoring commands using the new GP1/GP2 protocol
	gp1Data := struct {
		Level      int
		Autonomous bool
		Adjustment int
	}{}
	gp2Data := struct {
		Autonomous bool
		Adjustment int
	}{}

	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)

	// Test GP1 Level 1 for Blue alliance (teleop)
	gp1Data.Level = 1
	gp1Data.Autonomous = false
	gp1Data.Adjustment = 1
	blueWs.Write("GP1", gp1Data)
	blueWs.Write("GP1", gp1Data)
	blueWs.Write("GP1", gp1Data)
	gp1Data.Adjustment = -1
	blueWs.Write("GP1", gp1Data)
	blueWs.Write("GP1", gp1Data)
	gp1Data.Adjustment = 1
	blueWs.Write("GP1", gp1Data)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}

	// Test GP2 for Red alliance (teleop)
	gp2Data.Autonomous = false
	gp2Data.Adjustment = -1
	redWs.Write("GP2", gp2Data)
	redWs.Write("GP2", gp2Data)
	gp2Data.Adjustment = 1
	redWs.Write("GP2", gp2Data)
	redWs.Write("GP2", gp2Data)
	redWs.Write("GP2", gp2Data)
	gp2Data.Adjustment = -1
	redWs.Write("GP2", gp2Data)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}

	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 2, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count)

	// Send some gamepiece scoring commands using GP1 protocol
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count)

	// Test GP1 Level 1 for Red alliance (auto)
	gp1Data.Level = 1
	gp1Data.Autonomous = true
	gp1Data.Adjustment = 1
	redWs.Write("GP1", gp1Data)
	redWs.Write("GP1", gp1Data)
	redWs.Write("GP1", gp1Data)
	gp1Data.Adjustment = -1
	redWs.Write("GP1", gp1Data)
	for i := 0; i < 4; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}

	// Test GP1 Level 1 for Red alliance (teleop)
	gp1Data.Autonomous = false
	gp1Data.Adjustment = 1
	redWs.Write("GP1", gp1Data)
	redWs.Write("GP1", gp1Data)

	// Test GP1 Level 2 for Red alliance (auto)
	gp1Data.Level = 2
	gp1Data.Autonomous = true
	redWs.Write("GP1", gp1Data)
	redWs.Write("GP1", gp1Data)

	// Test GP1 Level 1 for Red alliance (teleop) - decrement
	gp1Data.Level = 1
	gp1Data.Autonomous = false
	gp1Data.Adjustment = -1
	redWs.Write("GP1", gp1Data)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}

	assert.Equal(t, 1, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count)
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count)
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count)
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count)

	// Send more gamepiece scoring commands using GP1 and GP2 protocol

	// Auto phase - GP1 Level 1
	gp1Data.Level = 1
	gp1Data.Autonomous = true
	gp1Data.Adjustment = 1
	redWs.Write("GP1", gp1Data)

	// Auto phase - GP2
	gp2Data.Autonomous = true
	gp2Data.Adjustment = 1
	redWs.Write("GP2", gp2Data)

	// Auto phase - GP1 Level 2
	gp1Data.Level = 2
	redWs.Write("GP1", gp1Data)
	redWs.Write("GP1", gp1Data)

	// Teleop phase - GP1 Level 2
	gp1Data.Autonomous = false
	redWs.Write("GP1", gp1Data)

	// Blue alliance - GP1 Level 1
	gp1Data.Level = 1
	gp1Data.Autonomous = false
	blueWs.Write("GP1", gp1Data)

	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	// Red Auto
	assert.Equal(
		t,
		4, // 2 from lines 198-200 + 2 from line 181
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		3, // 1 from line 190 + 2 from line 181
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		1, // 1 from line 195
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.AutoGamepiece2Count,
	)
	// Red Current
	assert.Equal(
		t,
		1, // 1 from line 204
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		1, // From line 179
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		2, // From line 134
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count,
	)
	// Blue Auto
	assert.Equal(
		t,
		0, // No commands sent for Blue Auto GP1 Level 2
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		0, // No commands sent for Blue Auto GP1 Level 1
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0, // No commands sent for Blue Auto GP2
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.AutoGamepiece2Count,
	)
	// Blue Current
	assert.Equal(
		t,
		0, // No commands sent for Blue Teleop GP1 Level 2
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level2Count,
	)
	assert.Equal(
		t,
		3, // 1 from line 209 + 2 from line 133
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece1Level1Count,
	)
	assert.Equal(
		t,
		0, // No commands sent for Blue Teleop GP2
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.TeleopGamepiece2Count,
	)

	// Send some park status commands
	parkData := struct {
		TeamPosition int
	}{} // Note: ParkStatus field is not used in the implementation, it toggles based on position only
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.Mayhem.ParkStatuses)
	assert.Equal(t, [3]bool{false, false, false}, web.arena.BlueRealtimeScore.CurrentScore.Mayhem.ParkStatuses)
	parkData.TeamPosition = 1
	redWs.Write("park", parkData)  // true
	blueWs.Write("park", parkData) // true
	parkData.TeamPosition = 2
	blueWs.Write("park", parkData) // true
	parkData.TeamPosition = 3
	blueWs.Write("park", parkData) // true
	redWs.Write("park", parkData)  // true
	redWs.Write("park", parkData)  // false
	parkData.TeamPosition = 2
	redWs.Write("park", parkData) // true
	for i := 0; i < 7; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(
		t,
		[3]bool{true, true, false},
		web.arena.RedRealtimeScore.CurrentScore.Mayhem.ParkStatuses,
	)
	assert.Equal(
		t,
		[3]bool{true, true, true},
		web.arena.BlueRealtimeScore.CurrentScore.Mayhem.ParkStatuses,
	)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)
	leaveData.TeamPosition = 0
	redWs.Write("leave", leaveData)
	// Send invalid GP1 command
	gp1Data.Level = 0 // Invalid level
	redWs.Write("GP1", gp1Data)

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
