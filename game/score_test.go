// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreAutoTreasurePoints(t *testing.T) {
	testCases := []struct {
		name     string
		score    Score
		expected int
	}{
		{"nothing placed", Score{}, 0},
		{"one on the floor", Score{AutoFloor: 1}, 4},
		{"three on the floor", Score{AutoFloor: 3}, 12},
		{"one on the first shelf", Score{AutoFirst: 1}, 8},
		{"two on the first shelf", Score{AutoFirst: 2}, 16},
		{"one on the top shelf", Score{AutoTop: 1}, 12},
		{"two on the top shelf", Score{AutoTop: 2}, 24},
		{"one of each", Score{AutoFloor: 1, AutoFirst: 1, AutoTop: 1}, 24},
		{"a full auto", Score{AutoFloor: 2, AutoFirst: 3, AutoTop: 4}, 8 + 24 + 48},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				summary := testCase.score.Summarize(&Score{})
				assert.Equal(t, testCase.expected, summary.AutoTreasurePoints)
				// Auto treasures are part of the auton points and of the treasure count, but never of the shelf count.
				assert.Equal(t, testCase.expected, summary.AutonPoints)
				assert.Equal(t, 0, summary.ShelfTreasureCount)
			},
		)
	}
}

func TestScoreTeleopTreasurePoints(t *testing.T) {
	testCases := []struct {
		name                       string
		score                      Score
		expectedPoints             int
		expectedShelfTreasureCount int
	}{
		{"nothing placed", Score{}, 0, 0},
		{"one on the floor", Score{TeleopFloor: 1}, 2, 0},
		{"five on the floor", Score{TeleopFloor: 5}, 10, 0},
		{"one on the first shelf", Score{TeleopFirst: 1}, 5, 1},
		{"three on the first shelf", Score{TeleopFirst: 3}, 15, 3},
		{"one on the top shelf", Score{TeleopTop: 1}, 10, 1},
		{"four on the top shelf", Score{TeleopTop: 4}, 40, 4},
		{"one stacked", Score{TeleopStacked: 1}, 8, 1},
		{"two stacked", Score{TeleopStacked: 2}, 16, 2},
		{
			"one of each",
			Score{TeleopFloor: 1, TeleopFirst: 1, TeleopTop: 1, TeleopStacked: 1},
			2 + 5 + 10 + 8,
			3,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				summary := testCase.score.Summarize(&Score{})
				assert.Equal(t, testCase.expectedPoints, summary.TeleopTreasurePoints)
				assert.Equal(t, testCase.expectedShelfTreasureCount, summary.ShelfTreasureCount)
				// Teleop treasures never count toward the auton points.
				assert.Equal(t, 0, summary.AutonPoints)
			},
		)
	}
}

func TestScoreTreasureCount(t *testing.T) {
	score := Score{
		AutoFloor:     1,
		AutoFirst:     2,
		AutoTop:       3,
		TeleopFloor:   4,
		TeleopFirst:   5,
		TeleopTop:     6,
		TeleopStacked: 7,
		Crown:         CrownTeleopTop,
		Toss:          true,
	}
	summary := score.Summarize(&Score{})

	// Every counter counts once; the crown is already one of the counted treasures and the toss cube is not a treasure.
	assert.Equal(t, 1+2+3+4+5+6+7, summary.TreasureCount)
	assert.Equal(t, 5+6+7, summary.ShelfTreasureCount)
}

func TestScoreCrownBonus(t *testing.T) {
	testCases := []struct {
		name                         string
		crown                        CrownPlacement
		expectedCrownBonusPoints     int
		expectedAutoTreasurePoints   int
		expectedTeleopTreasurePoints int
	}{
		{"no crown", CrownNone, 0, 0, 0},
		{"auto floor", CrownAutoFloor, 4, 4, 0},
		{"auto first", CrownAutoFirst, 8, 8, 0},
		{"auto top", CrownAutoTop, 12, 12, 0},
		{"teleop floor", CrownTeleopFloor, 2, 0, 2},
		{"teleop first", CrownTeleopFirst, 5, 0, 5},
		{"teleop top", CrownTeleopTop, 10, 0, 10},
		{"teleop stacked", CrownTeleopStacked, 8, 0, 8},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				// The bonus is added whenever the crown is set, even if the matching counter is zero.
				score := Score{Crown: testCase.crown}
				summary := score.Summarize(&Score{})
				assert.Equal(t, testCase.expectedCrownBonusPoints, summary.CrownBonusPoints)
				assert.Equal(t, testCase.crown, summary.Crown)
				assert.Equal(t, testCase.expectedAutoTreasurePoints, summary.AutoTreasurePoints)
				assert.Equal(t, testCase.expectedTeleopTreasurePoints, summary.TeleopTreasurePoints)

				// The bonus is counted once, inside the period's treasure points, and never again in the match points.
				assert.Equal(t, testCase.expectedAutoTreasurePoints, summary.AutonPoints)
				assert.Equal(
					t,
					testCase.expectedAutoTreasurePoints+testCase.expectedTeleopTreasurePoints,
					summary.MatchPoints,
				)

				// The crown selector never changes any treasure count.
				assert.Equal(t, 0, summary.TreasureCount)
				assert.Equal(t, 0, summary.ShelfTreasureCount)
			},
		)
	}
}

func TestScoreCrownDoublesItsOwnPlacementOnly(t *testing.T) {
	// A crown on the top shelf in teleop scores 10 (counter) + 10 (bonus) = 20; other treasures are unaffected.
	score := Score{TeleopTop: 3, Crown: CrownTeleopTop}
	summary := score.Summarize(&Score{})
	assert.Equal(t, 30+10, summary.TeleopTreasurePoints)
	assert.Equal(t, 3, summary.ShelfTreasureCount)

	// On the first shelf in auto: 8 + 8 = 16.
	score = Score{AutoFirst: 1, Crown: CrownAutoFirst}
	summary = score.Summarize(&Score{})
	assert.Equal(t, 16, summary.AutoTreasurePoints)

	// Stacked in teleop: 8 + 8 = 16.
	score = Score{TeleopStacked: 1, Crown: CrownTeleopStacked}
	summary = score.Summarize(&Score{})
	assert.Equal(t, 16, summary.TeleopTreasurePoints)

	// Nothing else is doubled: not the robot points, not the toss, not the foul points.
	score = Score{
		TeleopTop:       1,
		Crown:           CrownTeleopTop,
		LeaveStatuses:   [3]bool{true, true, false},
		EndgameStatuses: [3]EndgameStatus{EndgameBalance, EndgameNone, EndgameNone},
		Toss:            true,
	}
	summary = score.Summarize(&Score{Fouls: []Foul{{1, false, 0, ruleIdMa2616}}})
	assert.Equal(t, 8, summary.LeavePoints)
	assert.Equal(t, 12, summary.EndgamePoints)
	assert.Equal(t, 2, summary.TossPoints)
	assert.Equal(t, 5, summary.FoulPoints)
	assert.Equal(t, 8+20+12+2, summary.MatchPoints)
}

func TestScoreLeaveAndAutoBalance(t *testing.T) {
	testCases := []struct {
		name                      string
		leaveStatuses             [3]bool
		autoBalanceStatuses       [3]bool
		expectedLeavePoints       int
		expectedAutoBalancePoints int
	}{
		{"no robots", [3]bool{}, [3]bool{}, 0, 0},
		{"one robot left", [3]bool{true, false, false}, [3]bool{}, 4, 0},
		{"two robots left", [3]bool{true, true, false}, [3]bool{}, 8, 0},
		{"three robots left", [3]bool{true, true, true}, [3]bool{}, 12, 0},
		{"one robot balanced", [3]bool{}, [3]bool{true, false, false}, 0, 12},
		{"two robots balanced", [3]bool{}, [3]bool{true, true, false}, 0, 24},
		{"three robots balanced", [3]bool{}, [3]bool{true, true, true}, 0, 36},
		// Leave and auto balance are independent toggles and stack: 4 + 12 = 16 for one robot.
		{"one robot left and balanced", [3]bool{true, false, false}, [3]bool{true, false, false}, 4, 12},
		{"balanced without a leave", [3]bool{}, [3]bool{false, true, false}, 0, 12},
		// The third station is unused in a 2v2 match, so its statuses simply stay false.
		{"2v2 with both robots doing both", [3]bool{true, true, false}, [3]bool{true, true, false}, 8, 24},
		// A bypassed robot in the third station has no effect either way.
		{"3v3 with the third robot bypassed", [3]bool{true, true, false}, [3]bool{false, false, false}, 8, 0},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				score := Score{LeaveStatuses: testCase.leaveStatuses, AutoBalanceStatuses: testCase.autoBalanceStatuses}
				summary := score.Summarize(&Score{})
				assert.Equal(t, testCase.expectedLeavePoints, summary.LeavePoints)
				assert.Equal(t, testCase.expectedAutoBalancePoints, summary.AutoBalancePoints)
				assert.Equal(t, testCase.expectedLeavePoints+testCase.expectedAutoBalancePoints, summary.AutonPoints)
			},
		)
	}
}

func TestScoreEndgamePoints(t *testing.T) {
	testCases := []struct {
		name                        string
		endgameStatuses             [3]EndgameStatus
		expectedEndgamePoints       int
		expectedEndgameRankingPoint bool
	}{
		{"nothing", [3]EndgameStatus{}, 0, false},
		{"one parked", [3]EndgameStatus{EndgamePark, EndgameNone, EndgameNone}, 2, false},
		{"two parked", [3]EndgameStatus{EndgamePark, EndgamePark, EndgameNone}, 4, false},
		{"three parked", [3]EndgameStatus{EndgamePark, EndgamePark, EndgamePark}, 6, false},
		{"one balanced", [3]EndgameStatus{EndgameBalance, EndgameNone, EndgameNone}, 12, true},
		{"two balanced", [3]EndgameStatus{EndgameBalance, EndgameBalance, EndgameNone}, 24, true},
		{"three balanced", [3]EndgameStatus{EndgameBalance, EndgameBalance, EndgameBalance}, 36, true},
		{"one of each", [3]EndgameStatus{EndgameBalance, EndgamePark, EndgameNone}, 14, true},
		{"balanced in the third station", [3]EndgameStatus{EndgameNone, EndgameNone, EndgameBalance}, 12, true},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				score := Score{EndgameStatuses: testCase.endgameStatuses}
				summary := score.Summarize(&Score{})
				assert.Equal(t, testCase.expectedEndgamePoints, summary.EndgamePoints)
				assert.Equal(t, testCase.expectedEndgamePoints, summary.MatchPoints)
				assert.Equal(t, testCase.expectedEndgameRankingPoint, summary.EndgameRankingPoint)
				// The endgame is not part of the autonomous period.
				assert.Equal(t, 0, summary.AutonPoints)
			},
		)
	}
}

func TestScoreEndgameStatusIsExclusive(t *testing.T) {
	// Each robot has exactly one endgame status, so a robot can never score both park and balance.
	for status, expectedPoints := range map[EndgameStatus]int{EndgameNone: 0, EndgamePark: 2, EndgameBalance: 12} {
		score := Score{EndgameStatuses: [3]EndgameStatus{status, EndgameNone, EndgameNone}}
		assert.Equal(t, expectedPoints, score.Summarize(&Score{}).EndgamePoints)
	}

	// An auto balance is not an endgame balance and scores no endgame points and no Endgame RP.
	score := Score{AutoBalanceStatuses: [3]bool{true, true, true}}
	summary := score.Summarize(&Score{})
	assert.Equal(t, 36, summary.AutoBalancePoints)
	assert.Equal(t, 0, summary.EndgamePoints)
	assert.False(t, summary.EndgameRankingPoint)
}

func TestScoreToss(t *testing.T) {
	summary := (&Score{}).Summarize(&Score{})
	assert.Equal(t, 0, summary.TossPoints)

	// The Toss is worth 2 points once per alliance, regardless of alliance size.
	summary = (&Score{Toss: true}).Summarize(&Score{})
	assert.Equal(t, 2, summary.TossPoints)
	assert.Equal(t, 2, summary.MatchPoints)
	assert.Equal(t, 2, summary.Score)

	// The toss cube is not a treasure and counts toward nothing else.
	assert.Equal(t, 0, summary.AutonPoints)
	assert.Equal(t, 0, summary.TreasureCount)
	assert.Equal(t, 0, summary.ShelfTreasureCount)
	assert.Equal(t, 0, summary.BonusRankingPoints)
}

func TestScoreFoulPoints(t *testing.T) {
	testCases := []struct {
		name                          string
		opponentFouls                 []Foul
		expectedFoulPoints            int
		expectedNumOpponentMajorFouls int
	}{
		{"no fouls", nil, 0, 0},
		{"one minor", []Foul{{1, false, 0, ruleIdMa2616}}, 5, 0},
		{"one major", []Foul{{1, true, 0, ruleIdMa2604}}, 10, 1},
		{"two majors", []Foul{{1, true, 0, ruleIdMa2604}, {2, true, 0, ruleIdG415}}, 20, 2},
		{
			"a mix",
			[]Foul{{1, false, 0, ruleIdMa2606}, {2, true, 0, ruleIdMa2604}, {3, true, 0, ruleIdMa2604}},
			25,
			2,
		},
		// A foul with no rule selected scores normally.
		{"a minor with no rule", []Foul{{1, false, 0, 0}}, 5, 0},
		{"a major with no rule", []Foul{{1, true, 0, 0}}, 10, 1},
		// A ranking-point rule carries the ordinary point value as well.
		{"MA2601 and MA2602 for one incident", []Foul{{1, true, 0, ruleIdMa2601}, {2, true, 0, ruleIdMa2602}}, 20, 2},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				summary := (&Score{}).Summarize(&Score{Fouls: testCase.opponentFouls})
				assert.Equal(t, testCase.expectedFoulPoints, summary.FoulPoints)
				assert.Equal(t, testCase.expectedNumOpponentMajorFouls, summary.NumOpponentMajorFouls)
				assert.Equal(t, 0, summary.MatchPoints)
				assert.Equal(t, testCase.expectedFoulPoints, summary.Score)
			},
		)
	}

	// An alliance's own fouls never affect its own score.
	summary := (&Score{Fouls: []Foul{{1, true, 0, ruleIdMa2604}}}).Summarize(&Score{})
	assert.Equal(t, 0, summary.FoulPoints)
	assert.Equal(t, 0, summary.NumOpponentMajorFouls)
}

func TestScorePlayoffDq(t *testing.T) {
	score := Score{
		AutoFloor:           1,
		AutoFirst:           2,
		AutoTop:             3,
		TeleopFloor:         4,
		TeleopFirst:         5,
		TeleopTop:           6,
		TeleopStacked:       7,
		Crown:               CrownTeleopTop,
		LeaveStatuses:       [3]bool{true, true, true},
		AutoBalanceStatuses: [3]bool{true, true, true},
		EndgameStatuses:     [3]EndgameStatus{EndgameBalance, EndgameBalance, EndgamePark},
		Toss:                true,
	}
	opponentScore := Score{Fouls: []Foul{{1, true, 0, ruleIdMa2603}, {2, true, 0, ruleIdMa2601}}}

	// Without the disqualification everything scores.
	summary := score.Summarize(&opponentScore)
	assert.NotEqual(t, 0, summary.Score)
	assert.Equal(t, 3, summary.BonusRankingPoints)

	// With it, the entire summary is zero.
	score.PlayoffDq = true
	assert.Equal(t, &ScoreSummary{PlayoffDq: true}, score.Summarize(&opponentScore))

	// The opponent's own summary is unaffected by the disqualification.
	opponentSummary := opponentScore.Summarize(&score)
	assert.Equal(t, 0, opponentSummary.MatchPoints)
	assert.Equal(t, 0, opponentSummary.FoulPoints)
	assert.False(t, opponentSummary.PlayoffDq)
}

func TestScoreAutonRankingPoint(t *testing.T) {
	testCases := []struct {
		name                string
		score               Score
		opponentScore       Score
		expectedAutonPoints int
		expectedRankingPt   bool
		expectedByFoul      bool
	}{
		// Every auto value is a multiple of 4, so 16 and 20 are the boundary cases.
		{"one step below the threshold", Score{LeaveStatuses: [3]bool{true}, AutoTop: 1}, Score{}, 16, false, false},
		{
			"exactly at the threshold",
			Score{LeaveStatuses: [3]bool{true, true}, AutoBalanceStatuses: [3]bool{true}},
			Score{},
			20,
			true,
			false,
		},
		{"above the threshold", Score{AutoTop: 2}, Score{}, 24, true, false},
		{"nothing in auto", Score{}, Score{}, 0, false, false},
		// The crown bonus counts toward the auton points when the crown was placed in auto.
		{"reaching the threshold with the crown", Score{AutoFirst: 2, Crown: CrownAutoFirst}, Score{}, 24, true, false},
		{"a teleop crown does not count", Score{AutoFirst: 2, Crown: CrownTeleopTop}, Score{}, 16, false, false},
		// Teleop scoring and foul points never count toward the Auton RP.
		{"a huge teleop", Score{TeleopTop: 10}, Score{}, 0, false, false},
		{
			"foul points do not count",
			Score{LeaveStatuses: [3]bool{true}},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2604}, {2, true, 0, ruleIdMa2604}}},
			4,
			false,
			false,
		},
		// The opponent-violation alternative.
		{
			"below the threshold but the opponent committed MA2603",
			Score{LeaveStatuses: [3]bool{true}},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2603}}},
			4,
			true,
			true,
		},
		{
			"the opponent committed other fouls",
			Score{LeaveStatuses: [3]bool{true}},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2604}, {2, false, 0, ruleIdMa2606}}},
			4,
			false,
			false,
		},
		{
			"the opponent committed a foul with no rule",
			Score{LeaveStatuses: [3]bool{true}},
			Score{Fouls: []Foul{{1, true, 0, 0}}},
			4,
			false,
			false,
		},
		{
			"the alliance's own MA2603 does not earn it the RP",
			Score{LeaveStatuses: [3]bool{true}, Fouls: []Foul{{1, true, 0, ruleIdMa2603}}},
			Score{},
			4,
			false,
			false,
		},
		// Earning it both ways is still one ranking point.
		{"earned both ways", Score{AutoTop: 2}, Score{Fouls: []Foul{{1, true, 0, ruleIdMa2603}}}, 24, true, true},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				summary := testCase.score.Summarize(&testCase.opponentScore)
				assert.Equal(t, testCase.expectedAutonPoints, summary.AutonPoints)
				assert.Equal(t, testCase.expectedRankingPt, summary.AutonRankingPoint)
				assert.Equal(t, testCase.expectedByFoul, summary.AutonRankingPointByFoul)
				expectedBonusRankingPoints := 0
				if testCase.expectedRankingPt {
					expectedBonusRankingPoints = 1
				}
				assert.Equal(t, expectedBonusRankingPoints, summary.BonusRankingPoints)
			},
		)
	}
}

func TestScoreScoringRankingPoint(t *testing.T) {
	testCases := []struct {
		name                       string
		score                      Score
		expectedShelfTreasureCount int
		expectedRankingPt          bool
	}{
		{"nothing", Score{}, 0, false},
		{"one below the threshold", Score{TeleopFirst: 6, TeleopTop: 4, TeleopStacked: 1}, 11, false},
		{"exactly at the threshold", Score{TeleopFirst: 6, TeleopTop: 4, TeleopStacked: 2}, 12, true},
		{"above the threshold", Score{TeleopFirst: 13}, 13, true},
		{"all on the first shelf, one short", Score{TeleopFirst: 11}, 11, false},
		{"all stacked", Score{TeleopStacked: 12}, 12, true},
		// Floor treasures and every auto placement are excluded.
		{"floor treasures do not count", Score{TeleopFloor: 20, TeleopFirst: 11}, 11, false},
		{"auto placements do not count", Score{AutoFloor: 5, AutoFirst: 5, AutoTop: 5, TeleopFirst: 11}, 11, false},
		// The crown is one treasure, counted once by the counter it was entered in.
		{"the crown counts once", Score{TeleopTop: 11, Crown: CrownTeleopTop}, 11, false},
		{"the crown counts once, at the threshold", Score{TeleopTop: 12, Crown: CrownTeleopTop}, 12, true},
		// The toss cube is not a shelf treasure.
		{"the toss does not count", Score{TeleopTop: 11, Toss: true}, 11, false},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				// There is no opponent-violation alternative for this ranking point.
				for _, opponentScore := range []Score{
					{},
					{Fouls: []Foul{{1, true, 0, ruleIdMa2601}, {2, true, 0, ruleIdMa2603}}},
				} {
					summary := testCase.score.Summarize(&opponentScore)
					assert.Equal(t, testCase.expectedShelfTreasureCount, summary.ShelfTreasureCount)
					assert.Equal(t, ScoringRpThreshold, summary.ShelfTreasureGoal)
					assert.Equal(t, testCase.expectedRankingPt, summary.ScoringRankingPoint)
				}
			},
		)
	}
}

func TestScoreEndgameRankingPoint(t *testing.T) {
	testCases := []struct {
		name              string
		score             Score
		opponentScore     Score
		expectedRankingPt bool
		expectedByFoul    bool
	}{
		{"nothing", Score{}, Score{}, false, false},
		{
			"parked only",
			Score{EndgameStatuses: [3]EndgameStatus{EndgamePark, EndgamePark, EndgamePark}},
			Score{},
			false,
			false,
		},
		{
			"one robot balanced",
			Score{EndgameStatuses: [3]EndgameStatus{EndgameNone, EndgameBalance, EndgameNone}},
			Score{},
			true,
			false,
		},
		{
			"an auto balance does not count",
			Score{AutoBalanceStatuses: [3]bool{true, false, false}},
			Score{},
			false,
			false,
		},
		{"the opponent committed MA2601", Score{}, Score{Fouls: []Foul{{1, true, 0, ruleIdMa2601}}}, true, true},
		{"the opponent committed MA2602", Score{}, Score{Fouls: []Foul{{1, true, 0, ruleIdMa2602}}}, true, true},
		{
			"the opponent committed both for one incident",
			Score{},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2601}, {2, true, 0, ruleIdMa2602}}},
			true,
			true,
		},
		{
			"the opponent committed other fouls",
			Score{},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2604}, {2, false, 0, ruleIdMa2606}}},
			false,
			false,
		},
		{"the opponent committed a foul with no rule", Score{}, Score{Fouls: []Foul{{1, true, 0, 0}}}, false, false},
		{
			"the alliance's own MA2601 does not earn it the RP",
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2601}}},
			Score{},
			false,
			false,
		},
		{
			"earned both ways",
			Score{EndgameStatuses: [3]EndgameStatus{EndgameBalance}},
			Score{Fouls: []Foul{{1, true, 0, ruleIdMa2602}}},
			true,
			true,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				summary := testCase.score.Summarize(&testCase.opponentScore)
				assert.Equal(t, testCase.expectedRankingPt, summary.EndgameRankingPoint)
				assert.Equal(t, testCase.expectedByFoul, summary.EndgameRankingPointByFoul)
				expectedBonusRankingPoints := 0
				if testCase.expectedRankingPt {
					expectedBonusRankingPoints = 1
				}
				assert.Equal(t, expectedBonusRankingPoints, summary.BonusRankingPoints)
			},
		)
	}
}

func TestScoreBonusRankingPoints(t *testing.T) {
	// All three bonus ranking points at once, for a maximum of 3.
	score := Score{
		AutoTop:         2,
		TeleopFirst:     12,
		EndgameStatuses: [3]EndgameStatus{EndgameBalance, EndgameNone, EndgameNone},
	}
	summary := score.Summarize(&Score{})
	assert.True(t, summary.AutonRankingPoint)
	assert.True(t, summary.ScoringRankingPoint)
	assert.True(t, summary.EndgameRankingPoint)
	assert.Equal(t, 3, summary.BonusRankingPoints)
}

func TestScoreRankingPointThresholdsAreTunable(t *testing.T) {
	originalAutonRpThreshold := AutonRpThreshold
	originalScoringRpThreshold := ScoringRpThreshold
	defer func() {
		AutonRpThreshold = originalAutonRpThreshold
		ScoringRpThreshold = originalScoringRpThreshold
	}()
	assert.Equal(t, 20, originalAutonRpThreshold)
	assert.Equal(t, 12, originalScoringRpThreshold)

	score := Score{LeaveStatuses: [3]bool{true, true, true}, TeleopFirst: 8}
	summary := score.Summarize(&Score{})
	assert.Equal(t, 12, summary.AutonPoints)
	assert.False(t, summary.AutonRankingPoint)
	assert.False(t, summary.ScoringRankingPoint)
	assert.Equal(t, 12, summary.ShelfTreasureGoal)

	AutonRpThreshold = 12
	ScoringRpThreshold = 8
	summary = score.Summarize(&Score{})
	assert.True(t, summary.AutonRankingPoint)
	assert.True(t, summary.ScoringRankingPoint)
	assert.Equal(t, 8, summary.ShelfTreasureGoal)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	// Each mutation must be detected by Equals, so that every field of the score is covered.
	mutators := map[string]func(score *Score){
		"AutoFloor":              func(score *Score) { score.AutoFloor++ },
		"AutoFirst":              func(score *Score) { score.AutoFirst++ },
		"AutoTop":                func(score *Score) { score.AutoTop++ },
		"TeleopFloor":            func(score *Score) { score.TeleopFloor++ },
		"TeleopFirst":            func(score *Score) { score.TeleopFirst++ },
		"TeleopTop":              func(score *Score) { score.TeleopTop++ },
		"TeleopStacked":          func(score *Score) { score.TeleopStacked++ },
		"Crown":                  func(score *Score) { score.Crown = CrownAutoFloor },
		"LeaveStatuses[0]":       func(score *Score) { score.LeaveStatuses[0] = !score.LeaveStatuses[0] },
		"LeaveStatuses[1]":       func(score *Score) { score.LeaveStatuses[1] = !score.LeaveStatuses[1] },
		"LeaveStatuses[2]":       func(score *Score) { score.LeaveStatuses[2] = !score.LeaveStatuses[2] },
		"AutoBalanceStatuses[0]": func(score *Score) { score.AutoBalanceStatuses[0] = !score.AutoBalanceStatuses[0] },
		"AutoBalanceStatuses[1]": func(score *Score) { score.AutoBalanceStatuses[1] = !score.AutoBalanceStatuses[1] },
		"AutoBalanceStatuses[2]": func(score *Score) { score.AutoBalanceStatuses[2] = !score.AutoBalanceStatuses[2] },
		"EndgameStatuses[0]":     func(score *Score) { score.EndgameStatuses[0] = EndgameNone },
		"EndgameStatuses[1]":     func(score *Score) { score.EndgameStatuses[1] = EndgameBalance },
		"EndgameStatuses[2]":     func(score *Score) { score.EndgameStatuses[2] = EndgamePark },
		"Toss":                   func(score *Score) { score.Toss = !score.Toss },
		"Fouls length":           func(score *Score) { score.Fouls = []Foul{} },
		"Foul.FoulId":            func(score *Score) { score.Fouls[0].FoulId++ },
		"Foul.IsMajor":           func(score *Score) { score.Fouls[0].IsMajor = !score.Fouls[0].IsMajor },
		"Foul.TeamId":            func(score *Score) { score.Fouls[0].TeamId++ },
		"Foul.RuleId":            func(score *Score) { score.Fouls[0].RuleId++ },
		"PlayoffDq":              func(score *Score) { score.PlayoffDq = !score.PlayoffDq },
	}

	for name, mutate := range mutators {
		t.Run(
			name, func(t *testing.T) {
				mutated := TestScore1()
				mutate(mutated)
				assert.False(t, score1.Equals(mutated), "Equals did not detect a change to %s", name)
				assert.False(t, mutated.Equals(score1), "Equals did not detect a change to %s", name)
			},
		)
	}
}

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	// Red: leave 4 + 4 = 8; auto balance 12; auto treasure 4*1 + 8*2 + 12*1 = 32; auton 8 + 12 + 32 = 52.
	// Teleop 2*3 + 5*4 + 10*2 + 8*1 = 54, plus the top-shelf crown bonus 10 = 64.
	// Endgame 12 + 2 = 14; toss 2. Match points 52 + 64 + 14 + 2 = 132.
	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 8, redSummary.LeavePoints)
	assert.Equal(t, 12, redSummary.AutoBalancePoints)
	assert.Equal(t, 32, redSummary.AutoTreasurePoints)
	assert.Equal(t, 52, redSummary.AutonPoints)
	assert.Equal(t, 64, redSummary.TeleopTreasurePoints)
	assert.Equal(t, 14, redSummary.EndgamePoints)
	assert.Equal(t, 2, redSummary.TossPoints)
	assert.Equal(t, CrownTeleopTop, redSummary.Crown)
	assert.Equal(t, 10, redSummary.CrownBonusPoints)
	assert.Equal(t, 14, redSummary.TreasureCount)
	assert.Equal(t, 7, redSummary.ShelfTreasureCount)
	assert.Equal(t, 132, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.PostMatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 132, redSummary.Score)
	assert.Equal(t, false, redSummary.PlayoffDq)
	assert.Equal(t, true, redSummary.AutonRankingPoint)
	assert.Equal(t, false, redSummary.ScoringRankingPoint)
	assert.Equal(t, true, redSummary.EndgameRankingPoint)
	assert.Equal(t, 2, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)

	// Blue: leave 4; auto treasure 8; auton 12. Teleop 2*2 + 5*3 + 10*1 = 29. Endgame 2. Match points 43.
	// Red committed 5 majors and 2 minors: 5*10 + 2*5 = 60 foul points for blue.
	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 4, blueSummary.LeavePoints)
	assert.Equal(t, 0, blueSummary.AutoBalancePoints)
	assert.Equal(t, 8, blueSummary.AutoTreasurePoints)
	assert.Equal(t, 12, blueSummary.AutonPoints)
	assert.Equal(t, 29, blueSummary.TeleopTreasurePoints)
	assert.Equal(t, 2, blueSummary.EndgamePoints)
	assert.Equal(t, 0, blueSummary.TossPoints)
	assert.Equal(t, CrownNone, blueSummary.Crown)
	assert.Equal(t, 0, blueSummary.CrownBonusPoints)
	assert.Equal(t, 7, blueSummary.TreasureCount)
	assert.Equal(t, 4, blueSummary.ShelfTreasureCount)
	assert.Equal(t, 43, blueSummary.MatchPoints)
	assert.Equal(t, 0, blueSummary.PostMatchPoints)
	assert.Equal(t, 60, blueSummary.FoulPoints)
	assert.Equal(t, 103, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PlayoffDq)
	assert.Equal(t, 0, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 60, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	redSummary = redScore.Summarize(blueScore)
	assert.Equal(t, 0, redSummary.Score)
	assert.Equal(t, true, redSummary.PlayoffDq)
	// Red's disqualification should not affect blue's own summary.
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 103, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PlayoffDq)
	blueScore.PlayoffDq = true
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 0, blueSummary.Score)
	assert.Equal(t, true, blueSummary.PlayoffDq)
}
