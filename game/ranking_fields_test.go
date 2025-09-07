// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddScoreSummary(t *testing.T) {
	rand.Seed(0)
	redSummary := &ScoreSummary{
		AutoPoints:         30,
		Gamepiece2Points:   20,
		MatchPoints:        64,
		Score:              64,
		BonusRankingPoints: 2,
	}
	blueSummary := &ScoreSummary{
		AutoPoints:         16,
		Gamepiece2Points:   40,
		MatchPoints:        63,
		Score:              83,
		BonusRankingPoints: 1,
	}
	rankingFields := RankingFields{}

	// Add a loss.
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	expectedRankingFields := RankingFields{
		RankingPoints:    2, // 0 for loss + 2 bonus ranking points
		MatchPoints:      64,
		AutoPoints:       30,
		Gamepiece2Points: 20,
		Losses:           1,
		Played:           1,
		// Random is set by the function and can't be predicted
	}
	// Set the random value to match for comparison
	expectedRankingFields.Random = rankingFields.Random
	assert.Equal(t, expectedRankingFields, rankingFields)

	// Add a win.
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	expectedRankingFields = RankingFields{
		RankingPoints:    6,       // 2 (previous) + 3 (win) + 1 (bonus ranking point)
		MatchPoints:      64 + 63, // Previous + new match points
		AutoPoints:       30 + 16, // Previous + new auto points
		Gamepiece2Points: 20 + 40, // Previous + new gamepiece2 points
		Wins:             1,
		Losses:           1,
		Played:           2,
		// Random is set by the function and can't be predicted
	}
	// Set the random value to match for comparison
	expectedRankingFields.Random = rankingFields.Random
	assert.Equal(t, expectedRankingFields, rankingFields)

	// Add a tie.
	rankingFields.AddScoreSummary(redSummary, redSummary, false)
	expectedRankingFields = RankingFields{
		RankingPoints:    9,            // 6 (previous) + 1 (tie) + 2 (bonus ranking points)
		MatchPoints:      64 + 63 + 64, // Previous + new match points
		AutoPoints:       30 + 16 + 30, // Previous + new auto points
		Gamepiece2Points: 20 + 40 + 20, // Previous + new gamepiece2 points
		Wins:             1,
		Losses:           1,
		Ties:             1,
		Played:           3,
		// Random is set by the function and can't be predicted
	}
	// Set the random value to match for comparison
	expectedRankingFields.Random = rankingFields.Random
	assert.Equal(t, expectedRankingFields, rankingFields)

	// Add a disqualification.
	rankingFields.AddScoreSummary(blueSummary, redSummary, true)
	expectedRankingFields = RankingFields{
		RankingPoints:     9,            // No change from previous since disqualified
		MatchPoints:       64 + 63 + 64, // No change from previous since disqualified
		AutoPoints:        30 + 16 + 30, // No change from previous since disqualified
		Gamepiece2Points:  20 + 40 + 20, // No change from previous since disqualified
		Wins:              1,
		Losses:            1,
		Ties:              1,
		Disqualifications: 1,
		Played:            4, // Still increments played counter
		// Random is set by the function and can't be predicted
	}
	// Set the random value to match for comparison
	expectedRankingFields.Random = rankingFields.Random
	assert.Equal(t, expectedRankingFields, rankingFields)
}

func TestSortRankings(t *testing.T) {
	// Check tiebreakers.
	rankings := make(Rankings, 10)
	rankings[0] = Ranking{TeamId: 1, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.49}}
	rankings[1] = Ranking{TeamId: 2, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.51}}
	rankings[2] = Ranking{TeamId: 3, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 49, Random: 0.50}}
	rankings[3] = Ranking{TeamId: 4, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 51, Random: 0.50}}
	rankings[4] = Ranking{TeamId: 5, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 49, Gamepiece2Points: 50, Random: 0.50}}
	rankings[5] = Ranking{TeamId: 6, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 50, AutoPoints: 51, Gamepiece2Points: 50, Random: 0.50}}
	rankings[6] = Ranking{TeamId: 7, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 49, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.50}}
	rankings[7] = Ranking{TeamId: 8, RankingFields: RankingFields{RankingPoints: 50, MatchPoints: 51, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.50}}
	rankings[8] = Ranking{TeamId: 9, RankingFields: RankingFields{RankingPoints: 49, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.50}}
	rankings[9] = Ranking{TeamId: 10, RankingFields: RankingFields{RankingPoints: 51, MatchPoints: 50, AutoPoints: 50, Gamepiece2Points: 50, Random: 0.50}}
	for i := range rankings {
		rankings[i].Played = 10 // Set played matches for all to make averages easy
	}
	sort.Sort(rankings)
	assert.Equal(t, 10, rankings[0].TeamId)
	assert.Equal(t, 8, rankings[1].TeamId)
	assert.Equal(t, 6, rankings[2].TeamId)
	assert.Equal(t, 4, rankings[3].TeamId)
	assert.Equal(t, 2, rankings[4].TeamId)
	assert.Equal(t, 1, rankings[5].TeamId)
	assert.Equal(t, 3, rankings[6].TeamId)
	assert.Equal(t, 5, rankings[7].TeamId)
	assert.Equal(t, 7, rankings[8].TeamId)
	assert.Equal(t, 9, rankings[9].TeamId)
}
