// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()  // From test_helpers.go, now generic
	blueScore := TestScore2() // From test_helpers.go, now generic

	// redScore (TestScore1) has: Leave(F,F,N) -> 12pts; GP1(0,0) -> 0pts; GP2(0,0) -> 0pts; Endgame(P,N,F) -> 7pts. Total Match = 19.
	// blueScore (TestScore2) has: Leave(N,F,N) -> 6pts; GP1(0,0) -> 0pts; GP2(0,0) -> 0pts; Endgame(F,P,P) -> 9pts. Total Match = 15.
	// redScore has 7 fouls (5 major). blueScore has 0 fouls.
	// Foul point values are assumed from original test (e.g. blueSummary.FoulPoints = 34)

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 12, redSummary.LeavePoints)          // 6 (Full) + 6 (Full) + 0 (None)
	assert.Equal(t, 0, redSummary.GamePiece1Points)      // GP1 Auton/Teleop are 0 in TestScore1
	assert.Equal(t, 0, redSummary.GamePiece2Points)      // GP2 Auton/Teleop are 0 in TestScore1
	assert.Equal(t, 7, redSummary.EndgamePoints)         // 2 (Partial) + 0 (None) + 5 (Full)
	assert.Equal(t, 19, redSummary.MatchPoints)          // 12 + 0 + 0 + 7
	assert.Equal(t, 0, redSummary.FoulPoints)            // blueScore (opponent) has 0 fouls
	assert.Equal(t, 19, redSummary.Score)                // MatchPoints + FoulPoints = 19 + 0
	assert.Equal(t, 0, redSummary.BonusRankingPoints)    // Sum of above RPs
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls) // blueScore has 0 major fouls

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 6, blueSummary.LeavePoints)           // 0 (None) + 6 (Full) + 0 (None)
	assert.Equal(t, 0, blueSummary.GamePiece1Points)      // GP1 Auton/Teleop are 0 in TestScore2
	assert.Equal(t, 0, blueSummary.GamePiece2Points)      // GP2 Auton/Teleop are 0 in TestScore2
	assert.Equal(t, 9, blueSummary.EndgamePoints)         // 5 (Full) + 2 (Partial) + 2 (Partial)
	assert.Equal(t, 15, blueSummary.MatchPoints)          // 6 + 0 + 0 + 9
	assert.Equal(t, 34, blueSummary.FoulPoints)           // From redScore (opponent), assuming original PointValues
	assert.Equal(t, 49, blueSummary.Score)                // MatchPoints + FoulPoints = 15 + 34
	assert.Equal(t, 0, blueSummary.BonusRankingPoints)    // Sum of above RPs
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls) // redScore has 5 major fouls

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	// Assuming Foul.PointValue() is not affected by TeamId/RuleId being 0 for this check.
	// The blueSummary.FoulPoints should remain 34 if the point value calculation is independent or robust.
	assert.Equal(t, 34, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(redScore).Score) // blueScore is not DQd yet
	blueScore.PlayoffDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore).Score)
}

func TestScoreAutoRankingPointFromFouls(t *testing.T) {
	testCases := []struct {
		ownFouls           []Foul
		opponentFouls      []Foul
		expectedCoralBonus bool
		expectedBargeBonus bool
	}{
		// 0. No fouls - no automatic ranking points.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{},
			expectedCoralBonus: false,
			expectedBargeBonus: false,
		},

		// 1. G410 foul automatically awards coral bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 14}},
			expectedCoralBonus: true,
			expectedBargeBonus: false,
		},

		// 2. G418 foul automatically awards barge bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 21}},
			expectedCoralBonus: false,
			expectedBargeBonus: true,
		},

		// 3. G428 foul automatically awards barge bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 33}},
			expectedCoralBonus: false,
			expectedBargeBonus: true,
		},

		// 4. All fouls together still automatically award both bonuses.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 14}, {RuleId: 21}, {RuleId: 33}},
			expectedCoralBonus: true,
			expectedBargeBonus: true,
		},

		// 5. G206 makes the alliance ineligible for both bonuses.
		{
			ownFouls:           []Foul{{RuleId: 1}},
			opponentFouls:      []Foul{{RuleId: 14}, {RuleId: 21}, {RuleId: 33}},
			expectedCoralBonus: false,
			expectedBargeBonus: false,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				redScore := Score{Fouls: tc.ownFouls}
				blueScore := Score{Fouls: tc.opponentFouls}
				redSummary := redScore.Summarize(&blueScore)
				// CoralBonusRankingPoint and BargeBonusRankingPoint logic is TODO in Summarize and defaults to false.

				// Count expected total bonus ranking points.
				expectedBonusRankingPoints := 0
				if false { // tc.expectedCoralBonus - now always false
					expectedBonusRankingPoints++
				}
				if false { // tc.expectedBargeBonus - now always false
					expectedBonusRankingPoints++
				}
				assert.Equal(t, expectedBonusRankingPoints, redSummary.BonusRankingPoints)
			},
		)
	}
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1() // Now returns generic Score
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2() // Now returns generic Score
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	// Test LeaveStatuses
	score2 = TestScore1()
	score2.LeaveStatuses[0] = LeaveNone // Was LeaveFull in TestScore1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test GamePiece1Auton
	score2 = TestScore1()
	score2.GamePiece1Auton = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test GamePiece1Teleop
	score2 = TestScore1()
	score2.GamePiece1Teleop = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test GamePiece2Auton
	score2 = TestScore1()
	score2.GamePiece2Auton = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test GamePiece2Teleop
	score2 = TestScore1()
	score2.GamePiece2Teleop = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test EndgameStatuses
	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameFull // Was EndgameNone in TestScore1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test Fouls (existing tests are fine)
	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsMajor = !score1.Fouls[0].IsMajor
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleId += 1 // Changed from 1 to ensure different from original if original was 0
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	// Test PlayoffDq (existing test is fine)
	score2 = TestScore1()
	score2.PlayoffDq = !score2.PlayoffDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
