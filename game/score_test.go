// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 10, redSummary.LeavePoints)
	assert.Equal(t, 31, redSummary.AutoPoints)
	assert.Equal(t, 9, redSummary.NumGamepiece1)
	assert.Equal(t, 27, redSummary.Gamepiece1Points)
	assert.Equal(t, 6, redSummary.NumGamepiece2)
	assert.Equal(t, 16, redSummary.Gamepiece2Points)
	assert.Equal(t, 10, redSummary.ParkPoints)
	assert.Equal(t, 63, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 63, redSummary.Score)
	assert.Equal(t, true, redSummary.LeaveBonusRankingPoint)
	assert.Equal(t, true, redSummary.Gamepiece1BonusRankingPoint)
	assert.Equal(t, true, redSummary.ParkBonusRankingPoint)
	assert.Equal(t, 3, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 5, blueSummary.LeavePoints)
	assert.Equal(t, 30, blueSummary.AutoPoints)
	assert.Equal(t, 14, blueSummary.NumGamepiece1)
	assert.Equal(t, 40, blueSummary.Gamepiece1Points)
	assert.Equal(t, 5, blueSummary.NumGamepiece2)
	assert.Equal(t, 12, blueSummary.Gamepiece2Points)
	assert.Equal(t, 5, blueSummary.ParkPoints)
	assert.Equal(t, 62, blueSummary.MatchPoints)
	assert.Equal(t, 34, blueSummary.FoulPoints)
	assert.Equal(t, 96, blueSummary.Score)
	assert.Equal(t, false, blueSummary.LeaveBonusRankingPoint)
	assert.Equal(t, true, blueSummary.Gamepiece1BonusRankingPoint)
	assert.Equal(t, false, blueSummary.ParkBonusRankingPoint)
	assert.Equal(t, 1, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)
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
	score2.LeaveStatuses[0] = false
	assert.False(t, score1.Equals(score2))

	score2 = TestScore1()
	score2.Mayhem.AutoGamepiece1Level1Count++
	assert.False(t, score1.Equals(score2))

	score2 = TestScore1()
	score2.ParkStatuses[0] = false
	assert.False(t, score1.Equals(score2))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
}

func TestLeaveBonusRankingPoint(t *testing.T) {
	score := Score{RobotsBypassed: [3]bool{false, false, false}, LeaveStatuses: [3]bool{true, true, true}}
	summary := score.Summarize(&Score{})
	assert.True(t, summary.LeaveBonusRankingPoint)

	score.LeaveStatuses[1] = false
	summary = score.Summarize(&Score{})
	assert.False(t, summary.LeaveBonusRankingPoint)

	score.LeaveStatuses[1] = true
	score.RobotsBypassed[1] = true
	score.LeaveStatuses[1] = false
	summary = score.Summarize(&Score{})
	assert.True(t, summary.LeaveBonusRankingPoint)
}

func TestGamepiece1BonusRankingPoint(t *testing.T) {
	score := Score{Mayhem: Mayhem{AutoGamepiece1Level1Count: 4, TeleopGamepiece1Level2Count: 4}}
	summary := score.Summarize(&Score{})
	assert.True(t, summary.Gamepiece1BonusRankingPoint)

	score.Mayhem.TeleopGamepiece1Level2Count = 3
	summary = score.Summarize(&Score{})
	assert.False(t, summary.Gamepiece1BonusRankingPoint)
}

func TestParkBonusRankingPoint(t *testing.T) {
	score := Score{RobotsBypassed: [3]bool{false, false, false}, ParkStatuses: [3]bool{true, true, true}}
	summary := score.Summarize(&Score{})
	assert.True(t, summary.ParkBonusRankingPoint)

	score.ParkStatuses[1] = false
	summary = score.Summarize(&Score{})
	assert.False(t, summary.ParkBonusRankingPoint)

	score.ParkStatuses[1] = true
	score.RobotsBypassed[1] = true
	score.ParkStatuses[1] = false
	summary = score.Summarize(&Score{})
	assert.True(t, summary.ParkBonusRankingPoint)
}
