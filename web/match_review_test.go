// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
)

func TestMatchReview(t *testing.T) {
	web := setupTestWeb(t)

	match1 := model.Match{Type: model.Practice, ShortName: "P1", Status: game.RedWonMatch}
	match2 := model.Match{Type: model.Practice, ShortName: "P2"}
	match3 := model.Match{Type: model.Qualification, ShortName: "Q1", Status: game.BlueWonMatch}
	match4 := model.Match{Type: model.Playoff, ShortName: "SF1-1", Status: game.TieMatch}
	match5 := model.Match{Type: model.Playoff, ShortName: "SF1-2"}
	web.arena.Database.CreateMatch(&match1)
	web.arena.Database.CreateMatch(&match2)
	web.arena.Database.CreateMatch(&match3)
	web.arena.Database.CreateMatch(&match4)
	web.arena.Database.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">P1<")
	assert.Contains(t, recorder.Body.String(), ">P2<")
	assert.Contains(t, recorder.Body.String(), ">Q1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-2<")
}

func TestMatchReviewEditExistingResult(t *testing.T) {
	web := setupTestWeb(t)

	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 8
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	match, _ := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	match.Status = game.RedWonMatch
	web.arena.Database.UpdateMatch(match)
	matchResult := model.BuildTestMatchResult(match.Id, 1)
	matchResult.MatchType = match.Type
	assert.Nil(t, web.arena.Database.CreateMatchResult(matchResult))

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Contains(t, recorder.Body.String(), ">94<")  // The red score
	assert.Contains(t, recorder.Body.String(), ">186<") // The blue score

	// Check response for non-existent match.
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Quarterfinal 4-3 ")

	// Update the score to something else.
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"GamePiece1Teleop\":21,\"Fouls\":[{\"TeamId\":973,\"RuleId\":4}]}"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Contains(t, recorder.Body.String(), ">12<") // The red score (7 endgame + 5 fouls)
	assert.Contains(t, recorder.Body.String(), ">42<") // The blue score
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	web := setupTestWeb(t)

	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 8
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	// Get the match and verify initial state
	match, err := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	assert.Nil(t, err)
	match.Status = game.RedWonMatch
	err = web.arena.Database.UpdateMatch(match)
	assert.Nil(t, err)

	t.Logf("Testing with match ID: %d, Type: %s, DisplayName: %s", match.Id, match.Type, match.LongName)

	// Verify initial match review page
	recorder := web.getHttpResponse("/match_review")
	body := recorder.Body.String()
	assert.Equal(t, 200, recorder.Code, "Expected status 200, got %d. Body: %s", recorder.Code, body)
	assert.Contains(t, body, ">QF4-3<", "Expected to find '>QF4-3<' in response body: %s", body)
	assert.NotContains(t, body, ">13<", "Unexpectedly found '>13<' (red score) in initial response: %s", body)
	assert.NotContains(t, body, ">10<", "Unexpectedly found '>10<' (blue score) in initial response: %s", body)

	// Verify edit page loads
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	editBody := recorder.Body.String()
	assert.Equal(t, 200, recorder.Code, "Expected status 200, got %d. Body: %s", recorder.Code, editBody)
	assert.Contains(t, editBody, " Quarterfinal 4-3 ", "Expected to find 'Quarterfinal 4-3' in edit page: %s", editBody)

	// Prepare test data with predictable scores
	// Expected calculations:
	// Red: LeaveStatuses=[1,1,1] (3x3=9), GamePiece1Teleop=2 (2x2=4), GamePiece2Teleop=3 (3x3=9) = 22 total
	// Blue: GamePiece1Teleop=21 (21x2=42), plus a foul worth 5 points = 47 total
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"LeaveStatuses\":[1,1,1],\"GamePiece1Teleop\":2,\"GamePiece2Teleop\":3},\"BlueScore\":{"+
			"\"GamePiece1Teleop\":21,\"Fouls\":[{\"TeamId\":973,\"RuleId\":4,\"IsMajor\":true,\"TimeSec\":0}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)

	// Submit the match result
	t.Logf("Submitting match result: %s", postBody)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	postBodyResponse := recorder.Body.String()
	assert.Equal(t, 303, recorder.Code, "Expected status 303, got %d. Body: %s", recorder.Code, postBodyResponse)

	// Verify the match result was saved to the database
	savedResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err, "Error getting saved match result: %v", err)
	assert.NotNil(t, savedResult, "Expected match result to be saved")

	if savedResult != nil {
		t.Logf("Saved match result: %+v", savedResult)
		t.Logf("Red score: %+v", savedResult.RedScore)
		t.Logf("Blue score: %+v", savedResult.BlueScore)

		// Verify the scores were saved correctly
		assert.Equal(t, [3]game.LeaveStatus{1, 1, 1}, savedResult.RedScore.LeaveStatuses)
		assert.Equal(t, 2, savedResult.RedScore.GamePiece1Teleop)
		assert.Equal(t, 3, savedResult.RedScore.GamePiece2Teleop)
		assert.Equal(t, 21, savedResult.BlueScore.GamePiece1Teleop)
		assert.Equal(t, 1, len(savedResult.BlueScore.Fouls))

		// Calculate and verify score summaries
		redSummary := savedResult.RedScoreSummary()
		blueSummary := savedResult.BlueScoreSummary()
		t.Logf("Red summary: %+v", redSummary)
		t.Logf("Blue summary: %+v", blueSummary)

		// Verify the match status was updated
		updatedMatch, err := web.arena.Database.GetMatchById(match.Id)
		assert.Nil(t, err)
		t.Logf("Updated match status: %v", updatedMatch.Status)
	}

	// Verify the match review page shows the updated scores
	recorder = web.getHttpResponse("/match_review")
	finalBody := recorder.Body.String()
	
	// Log the full response for debugging
	t.Logf("=== FULL RESPONSE BODY ===\n%s\n=========================", finalBody)
	t.Logf("Final response body: %s", finalBody)
	assert.Equal(t, 200, recorder.Code, "Expected status 200, got %d. Body: %s", recorder.Code, finalBody)
	
	// Check for match identifier and scores in the response
	assert.Contains(t, finalBody, ">QF4-3<", "Expected to find '>QF4-3<' in final response")
	
	expectedRedScore := "22" // 9 (3x3 Leave) + 4 (2x2 GamePiece1) + 9 (3x3 GamePiece2)
	expectedBlueScore := "47" // 42 (21x2 GamePiece1) + 5 (foul)
	
	// Check for scores in the response
	if !assert.Contains(t, finalBody, ">"+expectedRedScore+"<", "Expected to find '>%s<' (red score) in final response", expectedRedScore) {
		t.Logf("Full response body when looking for red score %s: %s", expectedRedScore, finalBody)
	}
	
	if !assert.Contains(t, finalBody, ">"+expectedBlueScore+"<", "Expected to find '>%s<' (blue score) in final response", expectedBlueScore) {
		t.Logf("Full response body when looking for blue score %s: %s", expectedBlueScore, finalBody)
	}

	// Additional verification: Check if the match appears in recent results
	t.Logf("Match LongName: '%s' (length: %d)", match.LongName, len(match.LongName))
	
	// Check if Recent Results section exists
	recentResultsSection := "Recent Results"
	hasRecentResults := strings.Contains(finalBody, recentResultsSection)
	t.Logf("Page contains '%s' section: %v", recentResultsSection, hasRecentResults)
	
	// Check if there are any matches in the recent results
	hasRecentMatches := strings.Contains(finalBody, "<div class=\"match-review-item\"")
	t.Logf("Page contains any match items: %v", hasRecentMatches)
	
	// Find the position of Recent Results section
	recentResultsIndex := strings.Index(finalBody, recentResultsSection)
	if recentResultsIndex >= 0 {
		sectionEnd := recentResultsIndex + 200  // Look at next 200 chars after section header
		if sectionEnd > len(finalBody) {
			sectionEnd = len(finalBody)
		}
		recentResultsContent := finalBody[recentResultsIndex:sectionEnd]
		t.Logf("Recent Results section content (first 200 chars):\n%s", recentResultsContent)
	} else {
		t.Log("Could not find Recent Results section in the page")
	}

	// Check if match name appears anywhere in the page
	matchNamePos := strings.Index(finalBody, match.LongName)
	if matchNamePos >= 0 {
		// Show context around where the match name appears
		start := matchNamePos - 50
		if start < 0 {
			start = 0
		}
		end := matchNamePos + len(match.LongName) + 50
		if end > len(finalBody) {
			end = len(finalBody)
		}
		context := finalBody[start:end]
		t.Logf("Found match name at position %d. Context:\n...%s...", matchNamePos, context)
	} else {
		t.Logf("Match name '%s' not found anywhere in the response", match.LongName)
	}

	// Original assertions (will fail if not found)
	assert.Contains(t, finalBody, recentResultsSection, "Expected to find 'Recent Results' section")
	assert.Contains(t, finalBody, match.LongName, "Expected to find match name in recent results")
}

func TestMatchReviewEditCurrentMatch(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 352",
		ShortName: "Q352",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)
	web.arena.LoadMatch(&match)
	assert.Equal(t, match, *web.arena.CurrentMatch)

	recorder := web.getHttpResponse("/match_review/current/edit")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Qualification 352 ")

	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"GamePiece1Teleop\":21,\"Fouls\":[{\"TeamId\":973,\"RuleId\":1}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse("/match_review/current/edit", postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())
	assert.Equal(t, "/match_play", recorder.Header().Get("Location"))

	// Check that the persisted match is still unedited and that the realtime scores have been updated instead.
	match2, _ := web.arena.Database.GetMatchById(match.Id)
	assert.Equal(t, game.MatchScheduled, match2.Status)
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameNone, game.EndgameFull, game.EndgamePartial},
		web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses,
	)
	assert.Equal(t, 21, web.arena.BlueRealtimeScore.CurrentScore.GamePiece1Teleop)
	assert.Equal(t, 0, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.Cards))
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.Cards))
}
