// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestFieldMonitorDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/field_monitor?displayId=1&ds=false&fta=true&reversed=false")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Monitor - Untitled Event - Cheesy Arena")
}

func TestFmsFieldMonitorDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/fms_field_monitor?displayId=1&ds=false&fta=true&reversed=false")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Monitor - Untitled Event - Cheesy Arena")
}

func TestFieldMonitorDisplayTwoVsTwo(t *testing.T) {
	// The rows keep upstream's mirrored pairing (the right side runs 3, 2, 1 from the top); 2v2 only drops station 3.
	testCases := []struct {
		twoVsTwoMode   bool
		teamIdsInOrder []string
	}{
		{false, []string{"leftTeam1Id", "rightTeam3Id", "leftTeam2Id", "rightTeam2Id", "leftTeam3Id", "rightTeam1Id"}},
		{true, []string{"leftTeam1Id", "rightTeam2Id", "leftTeam2Id", "rightTeam1Id"}},
	}
	teamIdRegex := regexp.MustCompile(`id="((?:left|right)Team\dId)"`)
	for _, testCase := range testCases {
		web := setupTestWeb(t)
		web.arena.EventSettings.TwoVsTwoMode = testCase.twoVsTwoMode
		recorder := web.getHttpResponse("/displays/field_monitor?displayId=1&ds=false&fta=true&reversed=false")
		assert.Equal(t, 200, recorder.Code)
		body := recorder.Body.String()
		var teamIds []string
		for _, match := range teamIdRegex.FindAllStringSubmatch(body, -1) {
			teamIds = append(teamIds, match[1])
		}
		assert.Equal(t, testCase.teamIdsInOrder, teamIds)
	}
}

func TestFmsFieldMonitorDisplayTwoVsTwo(t *testing.T) {
	web := setupTestWeb(t)

	url := "/displays/fms_field_monitor?displayId=1&ds=false&fta=true&reversed=false"
	recorder := web.getHttpResponse(url)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "leftTeam3Id")
	assert.Contains(t, recorder.Body.String(), "rightTeam3Id")

	web.arena.EventSettings.TwoVsTwoMode = true
	recorder = web.getHttpResponse(url)
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "leftTeam3Id")
	assert.NotContains(t, recorder.Body.String(), "rightTeam3Id")
}

func TestFieldMonitorDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.Database.CreateTeam(&model.Team{Id: 254})
	assert.Nil(t, web.arena.SubstituteTeams(0, 0, 0, 254, 0, 0))

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(
		wsUrl+"/displays/field_monitor/websocket?displayId=1&ds=false&fta=false", nil,
	)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "displayConfiguration")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "matchLoad")

	// Should not be able to update team notes.
	ws.Write("updateTeamNotes", map[string]any{"station": "B1", "notes": "Bypassed in M1"})
	assert.Contains(t, readWebsocketError(t, ws), "Must be in FTA mode to update team notes")
	assert.Equal(t, "", web.arena.AllianceStations["B1"].Team.FtaNotes)
}

func TestFieldMonitorFtaDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.Database.CreateTeam(&model.Team{Id: 254})
	assert.Nil(t, web.arena.SubstituteTeams(0, 0, 0, 254, 0, 0))

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(
		wsUrl+"/displays/field_monitor/websocket?displayId=1&ds=false&fta=true", nil,
	)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "displayConfiguration")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "matchLoad")

	// Should not be able to update team notes.
	ws.Write("updateTeamNotes", map[string]any{"station": "B1", "notes": "Bypassed in M1"})
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, "Bypassed in M1", web.arena.AllianceStations["B1"].Team.FtaNotes)

	// Check error scenarios.
	ws.Write("updateTeamNotes", map[string]any{"station": "N", "notes": "Bypassed in M2"})
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("updateTeamNotes", map[string]any{"station": "R3", "notes": "Bypassed in M3"})
	assert.Contains(t, readWebsocketError(t, ws), "No team present")
}
