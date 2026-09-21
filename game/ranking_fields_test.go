// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func TestAddScoreSummary(t *testing.T) {
	randomizer := rand.New(rand.NewSource(0))
	RankingRandomFloat64 = randomizer.Float64
	redSummary := &ScoreSummary{
		MatchPoints:        67,
		Score:              67,
		BonusRankingPoints: 2,
	}
	blueSummary := &ScoreSummary{
		MatchPoints:        61,
		Score:              81,
		BonusRankingPoints: 1,
	}
	rankingFields := RankingFields{}

	// Add a loss.
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	assert.Equal(t, RankingFields{2, 67, 0.9451961492941164, 0, 1, 0, 0, 1}, rankingFields)

	// Add a win.
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	assert.Equal(t, RankingFields{6, 128, 0.24496508529377975, 1, 1, 0, 0, 2}, rankingFields)

	// Add a tie.
	rankingFields.AddScoreSummary(redSummary, redSummary, false)
	assert.Equal(t, RankingFields{9, 195, 0.6559562651954052, 1, 1, 1, 0, 3}, rankingFields)

	// Add a disqualification.
	rankingFields.AddScoreSummary(blueSummary, redSummary, true)
	assert.Equal(t, RankingFields{9, 195, 0.05434383959970039, 1, 1, 1, 1, 4}, rankingFields)
}

func TestSortRankings(t *testing.T) {
	// Check tiebreakers.
	rankings := make(Rankings, 4)
	rankings[0] = Ranking{1, 0, 0, RankingFields{50, 50, 0.49, 3, 2, 1, 0, 10}}
	rankings[1] = Ranking{2, 0, 0, RankingFields{50, 50, 0.51, 3, 2, 1, 0, 10}}
	rankings[2] = Ranking{3, 0, 0, RankingFields{50, 49, 0.50, 3, 2, 1, 0, 10}}
	rankings[3] = Ranking{4, 0, 0, RankingFields{49, 50, 0.50, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 1, rankings[1].TeamId)
	assert.Equal(t, 3, rankings[2].TeamId)
	assert.Equal(t, 4, rankings[3].TeamId)

	// Check with unequal number of matches played.
	rankings = make(Rankings, 3)
	rankings[0] = Ranking{1, 0, 0, RankingFields{10, 25, 0.49, 3, 2, 1, 0, 5}}
	rankings[1] = Ranking{2, 0, 0, RankingFields{19, 50, 0.51, 3, 2, 1, 0, 9}}
	rankings[2] = Ranking{3, 0, 0, RankingFields{20, 50, 0.51, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 3, rankings[1].TeamId)
	assert.Equal(t, 1, rankings[2].TeamId)
}
