// Copyright 2026 Team 254. All Rights Reserved.
//
// The worked examples from section 11 of the Medieval Mayhem game spec
// (specs/2026_medieval_mayhem.md). Every expected number below is copied from the spec; if a change to the scoring
// code breaks one of these, either the change or the spec is wrong.

package game

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

// The totals the spec's worked examples list for one alliance.
type expectedSummary struct {
	leavePoints          int
	autoBalancePoints    int
	autoTreasurePoints   int
	autonPoints          int
	teleopTreasurePoints int
	endgamePoints        int
	tossPoints           int
	matchPoints          int
	foulPoints           int
	score                int
	autonRankingPoint    bool
	scoringRankingPoint  bool
	endgameRankingPoint  bool
	rankingPoints        int
	treasureCount        int
	shelfTreasureCount   int
}

func assertWorkedExample(t *testing.T, name string, expected expectedSummary, summary, opponent *ScoreSummary) {
	t.Helper()
	assert.Equal(t, expected.leavePoints, summary.LeavePoints, "%s leave points", name)
	assert.Equal(t, expected.autoBalancePoints, summary.AutoBalancePoints, "%s auto balance points", name)
	assert.Equal(t, expected.autoTreasurePoints, summary.AutoTreasurePoints, "%s auto treasure points", name)
	assert.Equal(t, expected.autonPoints, summary.AutonPoints, "%s auton points", name)
	assert.Equal(t, expected.teleopTreasurePoints, summary.TeleopTreasurePoints, "%s teleop treasure points", name)
	assert.Equal(t, expected.endgamePoints, summary.EndgamePoints, "%s endgame points", name)
	assert.Equal(t, expected.tossPoints, summary.TossPoints, "%s toss points", name)
	assert.Equal(t, expected.matchPoints, summary.MatchPoints, "%s match points", name)
	assert.Equal(t, expected.foulPoints, summary.FoulPoints, "%s foul points", name)
	assert.Equal(t, expected.score, summary.Score, "%s score", name)
	assert.Equal(t, expected.autonRankingPoint, summary.AutonRankingPoint, "%s Auton RP", name)
	assert.Equal(t, expected.scoringRankingPoint, summary.ScoringRankingPoint, "%s Scoring RP", name)
	assert.Equal(t, expected.endgameRankingPoint, summary.EndgameRankingPoint, "%s Endgame RP", name)
	assert.Equal(t, expected.treasureCount, summary.TreasureCount, "%s treasure count", name)
	assert.Equal(t, expected.shelfTreasureCount, summary.ShelfTreasureCount, "%s shelf treasure count", name)
	assert.Equal(
		t, expected.rankingPoints, qualificationRankingPoints(summary, opponent), "%s qualification RP", name,
	)
}

// Returns the ranking points an alliance earns in a qualification match: the match result plus its bonus RPs.
func qualificationRankingPoints(summary, opponentSummary *ScoreSummary) int {
	rankingPoints := summary.BonusRankingPoints
	if summary.Score > opponentSummary.Score {
		rankingPoints += GetWinRankingPoints()
	} else if summary.Score == opponentSummary.Score {
		rankingPoints++
	}
	return rankingPoints
}

// Example A: an ordinary qualification match (2v2).
func exampleAScores() (*Score, *Score) {
	redScore := &Score{
		LeaveStatuses:   [3]bool{true, true, false},
		AutoFirst:       1,
		TeleopFloor:     3,
		TeleopFirst:     4,
		TeleopTop:       2,
		TeleopStacked:   1,
		EndgameStatuses: [3]EndgameStatus{EndgameBalance, EndgamePark, EndgameNone},
		Toss:            true,
		Fouls:           []Foul{{1, false, 0, ruleIdMa2616}},
	}
	blueScore := &Score{
		LeaveStatuses:   [3]bool{true, false, false},
		AutoFloor:       1,
		TeleopFloor:     2,
		TeleopFirst:     3,
		TeleopTop:       1,
		Crown:           CrownTeleopFirst,
		EndgameStatuses: [3]EndgameStatus{EndgamePark, EndgameNone, EndgameNone},
	}
	return redScore, blueScore
}

func TestWorkedExampleA(t *testing.T) {
	redScore, blueScore := exampleAScores()
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assertWorkedExample(
		t,
		"A red",
		expectedSummary{
			leavePoints: 8, autoBalancePoints: 0, autoTreasurePoints: 8, autonPoints: 16,
			teleopTreasurePoints: 54, endgamePoints: 14, tossPoints: 2,
			matchPoints: 86, foulPoints: 0, score: 86,
			autonRankingPoint: false, scoringRankingPoint: false, endgameRankingPoint: true,
			rankingPoints: 4, treasureCount: 11, shelfTreasureCount: 7,
		},
		redSummary,
		blueSummary,
	)
	assertWorkedExample(
		t,
		"A blue",
		expectedSummary{
			leavePoints: 4, autoBalancePoints: 0, autoTreasurePoints: 4, autonPoints: 8,
			teleopTreasurePoints: 34, endgamePoints: 2, tossPoints: 0,
			matchPoints: 44, foulPoints: 5, score: 49,
			autonRankingPoint: false, scoringRankingPoint: false, endgameRankingPoint: false,
			rankingPoints: 0, treasureCount: 7, shelfTreasureCount: 4,
		},
		blueSummary,
		redSummary,
	)

	status, _ := DetermineMatchStatus(redSummary, blueSummary, false)
	assert.Equal(t, RedWonMatch, status)
}

// Example B: thresholds exactly at the boundary, plus the crown (2v2, qualification).
func exampleBScores() (*Score, *Score) {
	redScore := &Score{
		LeaveStatuses:       [3]bool{true, true, false},
		AutoBalanceStatuses: [3]bool{true, false, false},
		TeleopFloor:         2,
		TeleopFirst:         6,
		TeleopTop:           4,
		TeleopStacked:       2,
		EndgameStatuses:     [3]EndgameStatus{EndgameNone, EndgamePark, EndgameNone},
	}
	blueScore := &Score{
		LeaveStatuses:   [3]bool{true, true, false},
		AutoFirst:       1,
		TeleopFloor:     1,
		TeleopFirst:     5,
		TeleopTop:       4,
		TeleopStacked:   2,
		Crown:           CrownTeleopTop,
		EndgameStatuses: [3]EndgameStatus{EndgameBalance, EndgameBalance, EndgameNone},
		Toss:            true,
	}
	return redScore, blueScore
}

func TestWorkedExampleB(t *testing.T) {
	redScore, blueScore := exampleBScores()
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assertWorkedExample(
		t,
		"B red",
		expectedSummary{
			leavePoints: 8, autoBalancePoints: 12, autoTreasurePoints: 0, autonPoints: 20,
			teleopTreasurePoints: 90, endgamePoints: 2, tossPoints: 0,
			matchPoints: 112, foulPoints: 0, score: 112,
			autonRankingPoint: true, scoringRankingPoint: true, endgameRankingPoint: false,
			rankingPoints: 2, treasureCount: 14, shelfTreasureCount: 12,
		},
		redSummary,
		blueSummary,
	)
	assertWorkedExample(
		t,
		"B blue",
		expectedSummary{
			leavePoints: 8, autoBalancePoints: 0, autoTreasurePoints: 8, autonPoints: 16,
			teleopTreasurePoints: 93, endgamePoints: 24, tossPoints: 2,
			matchPoints: 135, foulPoints: 0, score: 135,
			autonRankingPoint: false, scoringRankingPoint: false, endgameRankingPoint: true,
			rankingPoints: 4, treasureCount: 13, shelfTreasureCount: 11,
		},
		blueSummary,
		redSummary,
	)

	status, _ := DetermineMatchStatus(redSummary, blueSummary, false)
	assert.Equal(t, BlueWonMatch, status)
}

// Example C: ranking points earned through opponent violations, with fouls (2v2, qualification).
func exampleCScores() (*Score, *Score) {
	redScore := &Score{
		LeaveStatuses:   [3]bool{true, true, false},
		AutoFirst:       1,
		Crown:           CrownAutoFirst,
		TeleopFirst:     2,
		TeleopTop:       3,
		EndgameStatuses: [3]EndgameStatus{EndgamePark, EndgamePark, EndgameNone},
		Toss:            true,
		Fouls: []Foul{
			{1, true, 0, ruleIdMa2603},
			{2, true, 0, ruleIdMa2601},
			{3, true, 0, ruleIdMa2602},
		},
	}
	blueScore := &Score{
		LeaveStatuses: [3]bool{true, false, false},
		TeleopFloor:   4,
		TeleopFirst:   1,
		TeleopTop:     1,
		Fouls: []Foul{
			{1, false, 0, ruleIdMa2606},
			{2, true, 0, ruleIdMa2604},
			{3, true, 0, ruleIdMa2604},
		},
	}
	return redScore, blueScore
}

func TestWorkedExampleC(t *testing.T) {
	redScore, blueScore := exampleCScores()
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assertWorkedExample(
		t,
		"C red",
		expectedSummary{
			leavePoints: 8, autoBalancePoints: 0, autoTreasurePoints: 16, autonPoints: 24,
			teleopTreasurePoints: 40, endgamePoints: 4, tossPoints: 2,
			matchPoints: 70, foulPoints: 25, score: 95,
			autonRankingPoint: true, scoringRankingPoint: false, endgameRankingPoint: false,
			rankingPoints: 4, treasureCount: 6, shelfTreasureCount: 5,
		},
		redSummary,
		blueSummary,
	)
	assertWorkedExample(
		t,
		"C blue",
		expectedSummary{
			leavePoints: 4, autoBalancePoints: 0, autoTreasurePoints: 0, autonPoints: 4,
			teleopTreasurePoints: 23, endgamePoints: 0, tossPoints: 0,
			matchPoints: 27, foulPoints: 30, score: 57,
			autonRankingPoint: true, scoringRankingPoint: false, endgameRankingPoint: true,
			rankingPoints: 2, treasureCount: 6, shelfTreasureCount: 2,
		},
		blueSummary,
		redSummary,
	)

	// Blue earns both of its ranking points through red's violations; red earns its own on points.
	assert.True(t, blueSummary.AutonRankingPointByFoul)
	assert.True(t, blueSummary.EndgameRankingPointByFoul)
	assert.False(t, redSummary.AutonRankingPointByFoul)
	assert.False(t, redSummary.EndgameRankingPointByFoul)

	status, _ := DetermineMatchStatus(redSummary, blueSummary, false)
	assert.Equal(t, RedWonMatch, status)
}

// Variant C2: red's auto foul is entered as a major foul with no rule selected. Blue keeps its 30 foul points and its
// score of 57, but loses the Auton RP.
func TestWorkedExampleC2(t *testing.T) {
	redScore, blueScore := exampleCScores()
	redScore.Fouls[0].RuleId = 0
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assert.Equal(t, 30, blueSummary.FoulPoints)
	assert.Equal(t, 57, blueSummary.Score)
	assert.False(t, blueSummary.AutonRankingPoint)
	assert.False(t, blueSummary.AutonRankingPointByFoul)
	assert.True(t, blueSummary.EndgameRankingPoint)
	assert.Equal(t, 1, qualificationRankingPoints(blueSummary, redSummary))

	// Selecting MA2603 in Edit Match Result restores the ranking point.
	redScore.Fouls[0].RuleId = ruleIdMa2603
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 2, qualificationRankingPoints(blueSummary, redScore.Summarize(blueScore)))
}

// Example D1: a tied playoff match broken by the auton points.
func exampleD1Scores() (*Score, *Score) {
	redScore := &Score{
		LeaveStatuses:       [3]bool{true, true, false},
		AutoBalanceStatuses: [3]bool{true, false, false},
		AutoFirst:           1,
		TeleopFloor:         1,
		TeleopFirst:         2,
		TeleopTop:           1,
		EndgameStatuses:     [3]EndgameStatus{EndgameBalance, EndgameNone, EndgameNone},
	}
	blueScore := &Score{
		LeaveStatuses:   [3]bool{true, true, false},
		AutoFirst:       1,
		TeleopFloor:     1,
		TeleopFirst:     2,
		TeleopTop:       2,
		TeleopStacked:   1,
		EndgameStatuses: [3]EndgameStatus{EndgamePark, EndgamePark, EndgameNone},
		Toss:            true,
	}
	return redScore, blueScore
}

func TestWorkedExampleD1(t *testing.T) {
	redScore, blueScore := exampleD1Scores()
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assert.Equal(t, 28, redSummary.AutonPoints)
	assert.Equal(t, 22, redSummary.TeleopTreasurePoints)
	assert.Equal(t, 12, redSummary.EndgamePoints)
	assert.Equal(t, 0, redSummary.TossPoints)
	assert.Equal(t, 62, redSummary.MatchPoints)
	assert.Equal(t, 62, redSummary.Score)

	assert.Equal(t, 16, blueSummary.AutonPoints)
	assert.Equal(t, 40, blueSummary.TeleopTreasurePoints)
	assert.Equal(t, 4, blueSummary.EndgamePoints)
	assert.Equal(t, 2, blueSummary.TossPoints)
	assert.Equal(t, 62, blueSummary.MatchPoints)
	assert.Equal(t, 62, blueSummary.Score)

	// As a playoff match, the tie is broken by the auton points since neither alliance committed a major foul.
	status, tiebreakReason := DetermineMatchStatus(redSummary, blueSummary, true)
	assert.Equal(t, RedWonMatch, status)
	assert.Equal(t, "TIEBREAK: AUTON POINTS", tiebreakReason)

	// The same inputs as a qualification match are a tie, and both alliances keep their bonus ranking points.
	status, tiebreakReason = DetermineMatchStatus(redSummary, blueSummary, false)
	assert.Equal(t, TieMatch, status)
	assert.Equal(t, "", tiebreakReason)
	assert.Equal(t, 3, qualificationRankingPoints(redSummary, blueSummary))
	assert.Equal(t, 3, redSummary.ShelfTreasureCount)
	assert.False(t, redSummary.ScoringRankingPoint)
	assert.True(t, redSummary.AutonRankingPoint)
	assert.True(t, redSummary.EndgameRankingPoint)
	assert.Equal(t, 1, qualificationRankingPoints(blueSummary, redSummary))
	assert.Equal(t, 5, blueSummary.ShelfTreasureCount)
	assert.False(t, blueSummary.AutonRankingPoint)
	assert.False(t, blueSummary.ScoringRankingPoint)
	assert.False(t, blueSummary.EndgameRankingPoint)
}

// Example D2: the same match decided by the major-foul tiebreaker, even though red has more auton points.
func TestWorkedExampleD2(t *testing.T) {
	redScore, blueScore := exampleD1Scores()
	blueScore.TeleopTop = 1
	redScore.Fouls = []Foul{{1, true, 0, ruleIdG415}}
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assert.Equal(t, 30, blueSummary.TeleopTreasurePoints)
	assert.Equal(t, 52, blueSummary.MatchPoints)
	assert.Equal(t, 10, blueSummary.FoulPoints)
	assert.Equal(t, 62, blueSummary.Score)
	assert.Equal(t, 62, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 62, redSummary.Score)
	assert.Equal(t, 28, redSummary.AutonPoints)
	assert.Equal(t, 16, blueSummary.AutonPoints)

	status, tiebreakReason := DetermineMatchStatus(redSummary, blueSummary, true)
	assert.Equal(t, BlueWonMatch, status)
	assert.Equal(t, "TIEBREAK: MAJOR FOULS", tiebreakReason)
}

// Example D3: the same score and the same auton points, decided by the match points.
func TestWorkedExampleD3(t *testing.T) {
	redScore := &Score{
		LeaveStatuses: [3]bool{true, true, false},
		AutoFirst:     1,
		TeleopFloor:   2,
		TeleopTop:     4,
	}
	blueScore := &Score{
		LeaveStatuses: [3]bool{true, true, false},
		AutoFirst:     1,
		TeleopFloor:   2,
		TeleopFirst:   1,
		TeleopTop:     4,
		Fouls:         []Foul{{1, false, 0, ruleIdMa2607}},
	}
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assert.Equal(t, 16, redSummary.AutonPoints)
	assert.Equal(t, 44, redSummary.TeleopTreasurePoints)
	assert.Equal(t, 60, redSummary.MatchPoints)
	assert.Equal(t, 65, redSummary.Score)
	assert.Equal(t, 16, blueSummary.AutonPoints)
	assert.Equal(t, 49, blueSummary.TeleopTreasurePoints)
	assert.Equal(t, 65, blueSummary.MatchPoints)
	assert.Equal(t, 65, blueSummary.Score)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)
	assert.Equal(t, 0, blueSummary.NumOpponentMajorFouls)

	status, tiebreakReason := DetermineMatchStatus(redSummary, blueSummary, true)
	assert.Equal(t, BlueWonMatch, status)
	assert.Equal(t, "TIEBREAK: MATCH POINTS", tiebreakReason)
}

// Example D4: identical inputs leave every tiebreak level equal, so the match is a true tie and is replayed.
func TestWorkedExampleD4(t *testing.T) {
	redScore, _ := exampleD1Scores()
	blueScore, _ := exampleD1Scores()
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	status, tiebreakReason := DetermineMatchStatus(redSummary, blueSummary, true)
	assert.Equal(t, TieMatch, status)
	assert.Equal(t, "TRUE TIE", tiebreakReason)
}

// Example E: the 3v3 fallback, with blue's third robot bypassed.
func TestWorkedExampleE(t *testing.T) {
	redScore := &Score{
		LeaveStatuses:       [3]bool{true, true, true},
		AutoBalanceStatuses: [3]bool{false, false, true},
		TeleopFirst:         3,
		TeleopTop:           2,
		EndgameStatuses:     [3]EndgameStatus{EndgameBalance, EndgameBalance, EndgamePark},
		Toss:                true,
	}
	blueScore := &Score{
		// The bypassed robot in station 3 leaves its statuses at their zero values.
		LeaveStatuses:   [3]bool{true, true, false},
		AutoTop:         1,
		TeleopFirst:     8,
		TeleopTop:       3,
		TeleopStacked:   1,
		EndgameStatuses: [3]EndgameStatus{EndgameBalance, EndgameNone, EndgameNone},
	}
	redSummary := redScore.Summarize(blueScore)
	blueSummary := blueScore.Summarize(redScore)

	assertWorkedExample(
		t,
		"E red",
		expectedSummary{
			leavePoints: 12, autoBalancePoints: 12, autoTreasurePoints: 0, autonPoints: 24,
			teleopTreasurePoints: 35, endgamePoints: 26, tossPoints: 2,
			matchPoints: 87, foulPoints: 0, score: 87,
			autonRankingPoint: true, scoringRankingPoint: false, endgameRankingPoint: true,
			rankingPoints: 2, treasureCount: 5, shelfTreasureCount: 5,
		},
		redSummary,
		blueSummary,
	)
	assertWorkedExample(
		t,
		"E blue",
		expectedSummary{
			leavePoints: 8, autoBalancePoints: 0, autoTreasurePoints: 12, autonPoints: 20,
			teleopTreasurePoints: 78, endgamePoints: 12, tossPoints: 0,
			matchPoints: 110, foulPoints: 0, score: 110,
			autonRankingPoint: true, scoringRankingPoint: true, endgameRankingPoint: true,
			rankingPoints: 6, treasureCount: 13, shelfTreasureCount: 12,
		},
		blueSummary,
		redSummary,
	)

	status, _ := DetermineMatchStatus(redSummary, blueSummary, false)
	assert.Equal(t, BlueWonMatch, status)
}

// The ranking check from examples A, B and C: one qualification match each, sorted by ranking points, then average
// score, then average auton points.
func TestWorkedExampleRankings(t *testing.T) {
	type workedExampleMatch struct {
		redTeamIds  [2]int
		blueTeamIds [2]int
		redScore    *Score
		blueScore   *Score
	}
	aRed, aBlue := exampleAScores()
	bRed, bBlue := exampleBScores()
	cRed, cBlue := exampleCScores()
	matches := []workedExampleMatch{
		{[2]int{101, 102}, [2]int{103, 104}, aRed, aBlue},
		{[2]int{105, 106}, [2]int{107, 108}, bRed, bBlue},
		{[2]int{109, 110}, [2]int{111, 112}, cRed, cBlue},
	}

	// Give each team a distinct random value so that the partner ordering is deterministic: the lower team number of
	// each pair sorts first.
	randomValues := map[int]float64{}
	for i, teamId := range []int{101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112} {
		randomValues[teamId] = 1.0 - float64(i)/100.0
	}
	originalRankingRandomFloat64 := RankingRandomFloat64
	defer func() { RankingRandomFloat64 = originalRankingRandomFloat64 }()

	rankingsByTeam := map[int]*Ranking{}
	addTeam := func(teamId int, ownSummary, opponentSummary *ScoreSummary) {
		RankingRandomFloat64 = func() float64 { return randomValues[teamId] }
		ranking := &Ranking{TeamId: teamId}
		ranking.AddScoreSummary(ownSummary, opponentSummary, false)
		rankingsByTeam[teamId] = ranking
	}
	for _, match := range matches {
		redSummary := match.redScore.Summarize(match.blueScore)
		blueSummary := match.blueScore.Summarize(match.redScore)
		for _, teamId := range match.redTeamIds {
			addTeam(teamId, redSummary, blueSummary)
		}
		for _, teamId := range match.blueTeamIds {
			addTeam(teamId, blueSummary, redSummary)
		}
	}

	rankings := make(Rankings, 0, len(rankingsByTeam))
	for _, teamId := range []int{101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112} {
		rankings = append(rankings, *rankingsByTeam[teamId])
	}
	sort.Sort(rankings)

	// The table from the spec: ranking points, then score, then auton points.
	expectedRankings := []struct {
		teamId        int
		rankingPoints int
		scorePoints   int
		autonPoints   int
	}{
		{107, 4, 135, 16},
		{108, 4, 135, 16},
		{109, 4, 95, 24},
		{110, 4, 95, 24},
		{101, 4, 86, 16},
		{102, 4, 86, 16},
		{105, 2, 112, 20},
		{106, 2, 112, 20},
		{111, 2, 57, 4},
		{112, 2, 57, 4},
		{103, 0, 49, 8},
		{104, 0, 49, 8},
	}
	if assert.Equal(t, len(expectedRankings), len(rankings)) {
		for i, expected := range expectedRankings {
			assert.Equal(t, expected.teamId, rankings[i].TeamId, "rank %d", i+1)
			assert.Equal(t, expected.rankingPoints, rankings[i].RankingPoints, "team %d RP", expected.teamId)
			assert.Equal(t, expected.scorePoints, rankings[i].ScorePoints, "team %d score", expected.teamId)
			assert.Equal(t, expected.autonPoints, rankings[i].AutonPoints, "team %d auton", expected.teamId)
		}
	}
}
