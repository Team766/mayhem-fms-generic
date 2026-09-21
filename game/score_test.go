// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 0, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.PostMatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 0, redSummary.Score)
	assert.Equal(t, false, redSummary.PlayoffDq)
	assert.Equal(t, 0, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 0, blueSummary.MatchPoints)
	assert.Equal(t, 0, blueSummary.PostMatchPoints)
	assert.Equal(t, 85, blueSummary.FoulPoints)
	assert.Equal(t, 85, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PlayoffDq)
	assert.Equal(t, 0, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 85, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	redSummary = redScore.Summarize(blueScore)
	assert.Equal(t, 0, redSummary.Score)
	assert.Equal(t, true, redSummary.PlayoffDq)
	// Red's disqualification should not affect blue's own summary.
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 85, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PlayoffDq)
	blueScore.PlayoffDq = true
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 0, blueSummary.Score)
	assert.Equal(t, true, blueSummary.PlayoffDq)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsMajor = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId++
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleId = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.PlayoffDq = !score2.PlayoffDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
