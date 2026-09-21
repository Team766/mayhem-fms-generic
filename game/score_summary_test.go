// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	assertMatchStatus := func(
		expectedStatus MatchStatus,
		expectedTiebreaker string,
		redScoreSummary *ScoreSummary,
		blueScoreSummary *ScoreSummary,
		applyPlayoffTiebreakers bool,
	) {
		t.Helper()
		status, tiebreaker := DetermineMatchStatus(redScoreSummary, blueScoreSummary, applyPlayoffTiebreakers)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, expectedTiebreaker, tiebreaker)
	}

	redScoreSummary := &ScoreSummary{Score: 10}
	blueScoreSummary := &ScoreSummary{Score: 10}
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.Score = 11
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.Score = 12
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, true)

	// Tiebreak level 1: the alliance that committed fewer major fouls wins, which is the alliance whose
	// NumOpponentMajorFouls is higher.
	redScoreSummary.Score = 12
	redScoreSummary.NumOpponentMajorFouls = 11
	blueScoreSummary.NumOpponentMajorFouls = 10
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: MAJOR FOULS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.NumOpponentMajorFouls = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "TIEBREAK: MAJOR FOULS", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.NumOpponentMajorFouls = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)

	// Tiebreak level 2: more auton points wins.
	redScoreSummary.AutonPoints = 20
	blueScoreSummary.AutonPoints = 16
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: AUTON POINTS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.AutonPoints = 24
	assertMatchStatus(BlueWonMatch, "TIEBREAK: AUTON POINTS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.AutonPoints = 20
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)

	// Tiebreak level 3: more match points, that is the score without the foul points, wins.
	redScoreSummary.MatchPoints = 60
	blueScoreSummary.MatchPoints = 55
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: MATCH POINTS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.MatchPoints = 65
	assertMatchStatus(BlueWonMatch, "TIEBREAK: MATCH POINTS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.MatchPoints = 60
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)

	// A level is only reached when every level above it is equal.
	redScoreSummary = &ScoreSummary{Score: 50, NumOpponentMajorFouls: 1, AutonPoints: 0, MatchPoints: 0}
	blueScoreSummary = &ScoreSummary{Score: 50, NumOpponentMajorFouls: 0, AutonPoints: 40, MatchPoints: 50}
	assertMatchStatus(RedWonMatch, "TIEBREAK: MAJOR FOULS", redScoreSummary, blueScoreSummary, true)
	redScoreSummary = &ScoreSummary{Score: 50, NumOpponentMajorFouls: 1, AutonPoints: 0, MatchPoints: 0}
	blueScoreSummary = &ScoreSummary{Score: 50, NumOpponentMajorFouls: 1, AutonPoints: 40, MatchPoints: 50}
	assertMatchStatus(BlueWonMatch, "TIEBREAK: AUTON POINTS", redScoreSummary, blueScoreSummary, true)

	redScoreSummary = &ScoreSummary{Score: 0, PlayoffDq: true}
	blueScoreSummary = &ScoreSummary{Score: 0}
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, true)

	redScoreSummary = &ScoreSummary{Score: 0}
	blueScoreSummary = &ScoreSummary{Score: 0, PlayoffDq: true}
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, true)

	redScoreSummary = &ScoreSummary{Score: 0, PlayoffDq: true}
	blueScoreSummary = &ScoreSummary{Score: 0, PlayoffDq: true}
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)
}
