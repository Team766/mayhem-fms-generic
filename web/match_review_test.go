// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	assert.Contains(t, recorder.Body.String(), "match-review-rps")
	assert.Contains(t, recorder.Body.String(), "RP")
}

func TestMatchReviewTwoVsTwo(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{Type: model.Practice, ShortName: "P1", Red1: 101, Red2: 102, Red3: 103}
	web.arena.Database.CreateMatch(&match)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "101, 102, 103")

	web.arena.EventSettings.TwoVsTwoMode = true
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101, 102, 103")
	assert.Contains(t, recorder.Body.String(), "101, 102")
}

func TestMatchReviewEditExistingResultTwoVsTwo(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.Database.CreateTeam(&model.Team{Id: 101})
	web.arena.Database.CreateTeam(&model.Team{Id: 102})
	web.arena.Database.CreateTeam(&model.Team{Id: 103})
	web.arena.Database.CreateTeam(&model.Team{Id: 104})
	web.arena.Database.CreateTeam(&model.Team{Id: 105})
	web.arena.Database.CreateTeam(&model.Team{Id: 106})
	match := model.Match{
		Type: model.Practice, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106,
	}
	web.arena.Database.CreateMatch(&match)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = match.Id
	web.arena.Database.CreateMatchResult(matchResult)

	recorder := web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Team 103")

	// With the setting on, a match that still has a third team can't be edited via this form.
	web.arena.EventSettings.TwoVsTwoMode = true
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 500, recorder.Code)

	// A genuinely two-team match edits fine, with only two card rows.
	match2 := model.Match{Type: model.Practice, Red1: 101, Red2: 102, Blue1: 104, Blue2: 105}
	web.arena.Database.CreateMatch(&match2)
	matchResult2 := model.NewMatchResult()
	matchResult2.MatchId = match2.Id
	web.arena.Database.CreateMatchResult(matchResult2)
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match2.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Team 103")
	assert.Contains(t, recorder.Body.String(), "Team 102")
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
	// Red scores 132 match points and no foul points; blue scores 43 match points plus the 60 foul points from red's
	// five major and two minor fouls.
	assert.Regexp(t, `(?s)>\s*132\s*</td>`, recorder.Body.String()) // The red score
	assert.Regexp(t, `(?s)>\s*103\s*</td>`, recorder.Body.String()) // The blue score
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")

	// Check response for non-existent match.
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Quarterfinal 4-3 ")
	assert.Contains(t, recorder.Body.String(), `id="redScore"`)
	assert.Contains(t, recorder.Body.String(), `id="blueScore"`)
	assert.Contains(t, recorder.Body.String(), `id="redSummary"`)
	assert.Contains(t, recorder.Body.String(), `id="blueSummary"`)
	assert.Contains(t, recorder.Body.String(), "score-summary-table-red")
	assert.Contains(t, recorder.Body.String(), "score-summary-table-blue")
	assert.NotContains(t, recorder.Body.String(), "score-summary-rp")
	assert.NotContains(t, recorder.Body.String(), `data-summary-field="BonusRankingPoints"`)
	assert.NotContains(t, recorder.Body.String(), "scoreTemplate")
	assert.NotContains(t, recorder.Body.String(), "text/x-handlebars-template")

	// Update the score to something else.
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"Fouls\":[{\"TeamId\":1,\"RuleId\":4}]},\"BlueScore\":{"+
			"\"Fouls\":[{\"TeamId\":973,\"RuleId\":2,\"IsMajor\":true}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Regexp(t, `(?s)>\s*10\s*</td>`, recorder.Body.String()) // The red score, from blue's major foul
	assert.Regexp(t, `(?s)>\s*5\s*</td>`, recorder.Body.String())  // The blue score, from red's minor foul
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")
}

// Posts a full JSON body covering every game.Score field via the edit-result form's endpoint and confirms that
// every field round-trips into the persisted match result.
func TestMatchReviewEditResultRoundTripsGameFields(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type: model.Qualification, ShortName: "Q1", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105,
		Blue3: 106,
	}
	web.arena.Database.CreateMatch(&match)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = match.Id
	assert.Nil(t, web.arena.Database.CreateMatchResult(matchResult))

	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,"+
			"\"RedScore\":{\"AutoFloor\":1,\"AutoFirst\":2,\"AutoTop\":3,\"TeleopFloor\":4,\"TeleopFirst\":5,"+
			"\"TeleopTop\":6,\"TeleopStacked\":7,\"Crown\":6,\"LeaveStatuses\":[true,false,true],"+
			"\"AutoBalanceStatuses\":[false,true,false],\"EndgameStatuses\":[0,1,2],\"Toss\":true,"+
			"\"Fouls\":[{\"TeamId\":1,\"RuleId\":1}]},"+
			"\"BlueScore\":{\"AutoFloor\":0,\"Crown\":0,\"Toss\":false,\"Fouls\":[]},"+
			"\"RedCards\":{},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	savedResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, savedResult.RedScore.AutoFloor)
	assert.Equal(t, 2, savedResult.RedScore.AutoFirst)
	assert.Equal(t, 3, savedResult.RedScore.AutoTop)
	assert.Equal(t, 4, savedResult.RedScore.TeleopFloor)
	assert.Equal(t, 5, savedResult.RedScore.TeleopFirst)
	assert.Equal(t, 6, savedResult.RedScore.TeleopTop)
	assert.Equal(t, 7, savedResult.RedScore.TeleopStacked)
	assert.Equal(t, game.CrownTeleopTop, savedResult.RedScore.Crown)
	assert.Equal(t, [3]bool{true, false, true}, savedResult.RedScore.LeaveStatuses)
	assert.Equal(t, [3]bool{false, true, false}, savedResult.RedScore.AutoBalanceStatuses)
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameNone, game.EndgamePark, game.EndgameBalance},
		savedResult.RedScore.EndgameStatuses,
	)
	assert.True(t, savedResult.RedScore.Toss)
	assert.Equal(t, 1, len(savedResult.RedScore.Fouls))

	// The recomputed summary should reflect the new fields, including the crown bonus (CrownTeleopTop = +10).
	summary := savedResult.RedScoreSummary()
	assert.Equal(t, 4+4, summary.LeavePoints) // Two robots left (stations 1 and 3).
	assert.Equal(t, 4*1+8*2+12*3, summary.AutoTreasurePoints)
	assert.Equal(t, 10, summary.CrownBonusPoints)
	assert.Equal(t, 2*4+5*5+10*6+8*7+10, summary.TeleopTreasurePoints)
	assert.Equal(t, 2+12, summary.EndgamePoints) // Station 2 parked (2), station 3 balanced (12).
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	web := setupTestWeb(t)

	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 8
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	match, _ := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	match.Status = game.RedWonMatch
	web.arena.Database.UpdateMatch(match)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.NotRegexp(t, `(?s)>\s*35\s*<span class="match-review-rps`, recorder.Body.String()) // The red score
	assert.NotRegexp(t, `(?s)>\s*15\s*<span class="match-review-rps`, recorder.Body.String()) // The blue score

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Quarterfinal 4-3 ")

	// Update the score to something else.
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"Fouls\":[{\"TeamId\":1,\"RuleId\":4}]},\"BlueScore\":{"+
			"\"Fouls\":[{\"TeamId\":973,\"RuleId\":2,\"IsMajor\":true}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Regexp(t, `(?s)>\s*10\s*</td>`, recorder.Body.String()) // The red score, from blue's major foul
	assert.Regexp(t, `(?s)>\s*5\s*</td>`, recorder.Body.String())  // The blue score, from red's minor foul
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")
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
	assert.Contains(t, recorder.Body.String(), "score-summary-rp")
	assert.Contains(t, recorder.Body.String(), "Ranking Points")

	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{},\"BlueScore\":{"+
			"\"Fouls\":[{\"TeamId\":973,\"RuleId\":1}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse("/match_review/current/edit", postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())
	assert.Equal(t, "/match_play", recorder.Header().Get("Location"))

	// Check that the persisted match is still unedited and that the realtime scores have been updated instead.
	match2, _ := web.arena.Database.GetMatchById(match.Id)
	assert.Equal(t, game.MatchScheduled, match2.Status)
	assert.Equal(t, 0, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.Cards))
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.Cards))
}

func TestMatchReviewSummary(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 1",
		ShortName: "Q1",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)

	postBody := fmt.Sprintf(
		"{\"MatchId\":%d,\"RedScore\":{},\"BlueScore\":{"+
			"\"Fouls\":[{\"TeamId\":1004,\"RuleId\":2,\"IsMajor\":true}]},"+
			"\"RedCards\":{},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse(fmt.Sprintf("/match_review/%d/summary", match.Id), postBody)
	assert.Equal(t, 200, recorder.Code, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header()["Content-Type"][0])

	var response MatchReviewSummaryResponse
	assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 10, response.RedSummary.Score)
	assert.Equal(t, 0, response.RedSummary.MatchPoints)
	assert.Equal(t, 10, response.RedSummary.FoulPoints)
	assert.Equal(t, 0, response.BlueSummary.Score)

	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)

	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/summary", match.Id), "{\"MatchId\":12345}")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "match ID 12345")
}

func TestMatchReviewSummaryCurrentMatch(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 1",
		ShortName: "Q1",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)
	web.arena.LoadMatch(&match)

	postBody := fmt.Sprintf(
		"{\"MatchId\":%d,\"RedScore\":{\"Fouls\":[{\"TeamId\":1001,\"RuleId\":2,\"IsMajor\":true}]},"+
			"\"BlueScore\":{},\"RedCards\":{},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse("/match_review/current/summary", postBody)
	assert.Equal(t, 200, recorder.Code, recorder.Body.String())

	var response MatchReviewSummaryResponse
	assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 0, response.RedSummary.Score)
	assert.Equal(t, 10, response.BlueSummary.Score)

	assert.Equal(t, 0, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)
}
