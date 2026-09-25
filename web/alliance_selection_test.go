// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllianceSelection(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 15
	web.arena.EventSettings.SelectionRound3Order = "L"
	for i := 1; i <= 10; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Check that there are no alliance placeholders to start.
	recorder := web.getHttpResponse("/alliance_selection")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Start the alliance selection.
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 15, len(web.arena.AllianceSelectionAlliances)) {
		assert.Equal(t, 4, len(web.arena.AllianceSelectionAlliances[0].TeamIds))
	}
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "Captain")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Reset the alliance selection.
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")
	web.arena.EventSettings.NumPlayoffAlliances = 3
	web.arena.EventSettings.SelectionRound3Order = ""
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 3, len(web.arena.AllianceSelectionAlliances)) {
		assert.Equal(t, 3, len(web.arena.AllianceSelectionAlliances[0].TeamIds))
	}

	// Update one team at a time.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=110")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 110, web.arena.AllianceSelectionAlliances[0].TeamIds[0])
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "\"110\"")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Update multiple teams at a time.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection1_0=103")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 101, web.arena.AllianceSelectionAlliances[0].TeamIds[0])
	assert.Equal(t, 102, web.arena.AllianceSelectionAlliances[0].TeamIds[1])
	assert.Equal(t, 103, web.arena.AllianceSelectionAlliances[1].TeamIds[0])
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Update remainder of teams.
	recorder = web.postHttpResponse(
		"/alliance_selection",
		"selection0_0=101&selection0_1=102&selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106&"+
			"selection2_0=107&selection2_1=108&selection2_2=109",
	)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Finalize alliance selection.
	web.arena.Database.CreateTeam(&model.Team{Id: 254, YellowCard: true})
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, err := web.arena.Database.GetAllAlliances()
	assert.Nil(t, err)
	if assert.Equal(t, 3, len(alliances)) {
		assert.Equal(t, 101, alliances[0].TeamIds[0])
		assert.Equal(t, 105, alliances[1].TeamIds[1])
		assert.Equal(t, 109, alliances[2].TeamIds[2])

		// Check that the initial lineup is populated correctly.
		assert.Equal(t, 102, alliances[0].Lineup[0])
		assert.Equal(t, 101, alliances[0].Lineup[1])
		assert.Equal(t, 103, alliances[0].Lineup[2])
	}
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, false)
	assert.Nil(t, err)
	assert.Equal(t, 16, len(matches))
	team, _ := web.arena.Database.GetTeamById(254)
	assert.False(t, team.YellowCard)
}

func TestAllianceSelectionTwoVsTwo(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.TwoVsTwoMode = true
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	web.arena.EventSettings.SelectionRound3Order = "L" // Must not push 2v2 alliances to four teams.
	assert.Nil(t, web.arena.CreatePlayoffTournament()) // As saving the settings would.
	for i := 1; i <= 4; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Starting alliance selection in 2v2 creates alliances of two, with no third or backup round.
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 2, len(web.arena.AllianceSelectionAlliances)) {
		assert.Equal(t, 2, len(web.arena.AllianceSelectionAlliances[0].TeamIds))
	}

	// Autofocus should never advance past the second (last) column.
	nextRow, nextCol := web.determineNextCell()
	assert.Equal(t, 0, nextRow)
	assert.Equal(t, 0, nextCol)

	recorder = web.postHttpResponse(
		"/alliance_selection", "selection0_0=101&selection0_1=102&selection1_0=103&selection1_1=104",
	)
	assert.Equal(t, 303, recorder.Code)
	nextRow, nextCol = web.determineNextCell()
	assert.Equal(t, -1, nextRow)
	assert.Equal(t, -1, nextCol)

	// Finalizing must leave the third lineup slot at 0 and create playoff matches with no third team.
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, err := web.arena.Database.GetAllAlliances()
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(alliances)) {
		assert.Equal(t, []int{101, 102}, alliances[0].TeamIds)
		assert.Equal(t, [3]int{102, 101, 0}, alliances[0].Lineup)
	}
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, false)
	assert.Nil(t, err)
	if assert.NotEmpty(t, matches) {
		var sawRealTeam bool
		for _, match := range matches {
			assert.Equal(t, 0, match.Red3)
			assert.Equal(t, 0, match.Blue3)
			if match.Red1 != 0 || match.Red2 != 0 || match.Blue1 != 0 || match.Blue2 != 0 {
				sawRealTeam = true
			}
		}
		assert.True(t, sawRealTeam, "expected at least one playoff match to already have real teams assigned")
	}

	// The bracket shows both teams of each two-team alliance.
	recorder = web.getHttpResponse("/api/bracket/svg")
	assert.Equal(t, 200, recorder.Code)
	for _, teamNum := range []string{`"teamnum r">101<`, `"teamnum r">102<`, `"teamnum b">103<`, `"teamnum b">104<`} {
		assert.Contains(t, recorder.Body.String(), teamNum)
	}
}

// Verifies that changing the 2v2 setting partway through an alliance selection neither crashes the page nor
// changes which cells the selection autofocuses: that follows the size the alliances were created with.
func TestAllianceSelectionTwoVsTwoSettingChangedMidSelection(t *testing.T) {
	testCases := []struct {
		name             string
		twoVsTwoAtStart  bool
		teamsPerAlliance int
		// Next cell after the first two columns are filled.
		expectedRow int
		expectedCol int
	}{
		{"2v2 turned off", true, 2, -1, -1},
		{"2v2 turned on", false, 3, 1, 2},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			web := setupTestWeb(t)
			web.arena.EventSettings.TwoVsTwoMode = testCase.twoVsTwoAtStart
			web.arena.EventSettings.NumPlayoffAlliances = 2
			web.arena.EventSettings.SelectionRound2Order = "L"
			for i := 1; i <= 6; i++ {
				web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
			}
			recorder := web.postHttpResponse("/alliance_selection/start", "")
			assert.Equal(t, 303, recorder.Code)
			if !assert.Equal(t, testCase.teamsPerAlliance, len(web.arena.AllianceSelectionAlliances[0].TeamIds)) {
				return
			}

			web.arena.EventSettings.TwoVsTwoMode = !testCase.twoVsTwoAtStart
			// A backup round turned on partway through must not be looked for in alliances created without one.
			web.arena.EventSettings.SelectionRound3Order = "F"

			recorder = web.getHttpResponse("/alliance_selection")
			assert.Equal(t, 200, recorder.Code)
			nextRow, nextCol := web.determineNextCell()
			assert.Equal(t, 0, nextRow)
			assert.Equal(t, 0, nextCol)

			recorder = web.postHttpResponse(
				"/alliance_selection", "selection0_0=101&selection0_1=102&selection1_0=103&selection1_1=104",
			)
			assert.Equal(t, 303, recorder.Code)
			recorder = web.getHttpResponse("/alliance_selection")
			assert.Equal(t, 200, recorder.Code)
			nextRow, nextCol = web.determineNextCell()
			assert.Equal(t, testCase.expectedRow, nextRow)
			assert.Equal(t, testCase.expectedCol, nextCol)
		})
	}
}

func TestAllianceSelectionErrors(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	for i := 1; i <= 6; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Start an alliance selection that is already underway.
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already in progress")

	// Select invalid teams.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid team number")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=100")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "ineligible for selection")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection1_1=101")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already part of an alliance")

	// Finalize early and without required parameters.
	recorder = web.postHttpResponse(
		"/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM&matchSpacingSec=360",
	)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "until all spots have been filled")
	recorder = web.postHttpResponse(
		"/alliance_selection",
		"selection0_0=101&selection0_1=102&selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106",
	)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "valid start time")

	// Finalize for real.
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)

	// Do other things after finalization.
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	web.arena.AllianceSelectionRankedTeams = []model.AllianceSelectionRankedTeam{}
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
}

func TestAllianceSelectionReset(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	for i := 1; i <= 6; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Start, populate, and finalize the alliance selection.
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse(
		"/alliance_selection",
		"selection0_0=101&selection0_1=102&selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106",
	)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ := web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ := web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)

	// Reset the alliance selection before any matches have been played.
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.Empty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.Empty(t, matches)

	// Start, populate, and finalize the alliance selection again.
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse(
		"/alliance_selection",
		"selection0_0=101&selection0_1=102&selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106",
	)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)

	// Mark a match as played and verify that the alliance selection can no longer be reset.
	matches[0].Status = game.RedWonMatch
	assert.Nil(t, web.arena.Database.UpdateMatch(&matches[0]))
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "matches have already started")
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)
}

func TestAllianceSelectionAutofocus(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2

	// Straight draft.
	web.arena.EventSettings.SelectionRound2Order = "F"
	web.arena.EventSettings.SelectionRound3Order = "F"
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j := web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[0] = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[1] = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[0] = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[1] = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[2] = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[2] = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[3] = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[3] = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)

	// Double-serpentine draft.
	web.arena.EventSettings.SelectionRound2Order = "L"
	web.arena.EventSettings.SelectionRound3Order = "L"
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[0] = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[1] = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[0] = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[1] = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[2] = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[2] = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[3] = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[3] = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)
}

func TestAllianceSelectionWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/alliance_selection/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "allianceSelection")
	readWebsocketType(t, ws, "audienceDisplayMode")

	// Test starting and stopping the timer.
	allianceSelectionMessage := struct {
		ShowTimer bool
	}{}
	ws.Write("startTimer", nil)
	assert.Nil(t, mapstructure.Decode(readWebsocketType(t, ws, "allianceSelection"), &allianceSelectionMessage))
	assert.Equal(t, true, allianceSelectionMessage.ShowTimer)
	ws.Write("stopTimer", nil)
	assert.Nil(t, mapstructure.Decode(readWebsocketType(t, ws, "allianceSelection"), &allianceSelectionMessage))
	assert.Equal(t, true, allianceSelectionMessage.ShowTimer)
	ws.Write("hideTimer", nil)
	assert.Nil(t, mapstructure.Decode(readWebsocketType(t, ws, "allianceSelection"), &allianceSelectionMessage))
	assert.Equal(t, false, allianceSelectionMessage.ShowTimer)
	ws.Write("restartTimer", nil)
	assert.Nil(t, mapstructure.Decode(readWebsocketType(t, ws, "allianceSelection"), &allianceSelectionMessage))
	assert.Equal(t, true, allianceSelectionMessage.ShowTimer)
}
