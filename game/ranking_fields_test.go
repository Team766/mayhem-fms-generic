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
		AutonPoints:        24,
		Score:              67,
		BonusRankingPoints: 2,
	}
	blueSummary := &ScoreSummary{
		MatchPoints:        61,
		AutonPoints:        16,
		Score:              81,
		BonusRankingPoints: 1,
	}
	rankingFields := RankingFields{}

	// Add a loss. The tiebreaker sums take the score including foul points, and the auton points.
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	assert.Equal(t, RankingFields{2, 67, 24, 0.9451961492941164, 0, 1, 0, 0, 1}, rankingFields)

	// Add a win.
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	assert.Equal(t, RankingFields{6, 148, 40, 0.24496508529377975, 1, 1, 0, 0, 2}, rankingFields)

	// Add a tie.
	rankingFields.AddScoreSummary(redSummary, redSummary, false)
	assert.Equal(t, RankingFields{9, 215, 64, 0.6559562651954052, 1, 1, 1, 0, 3}, rankingFields)

	// Add a disqualification. It adds one to Played and Disqualifications and nothing else.
	rankingFields.AddScoreSummary(blueSummary, redSummary, true)
	assert.Equal(t, RankingFields{9, 215, 64, 0.05434383959970039, 1, 1, 1, 1, 4}, rankingFields)
}

func TestSortRankings(t *testing.T) {
	// Check each tiebreaker level in turn: ranking points, then score, then auton points, then the random value.
	rankings := make(Rankings, 5)
	rankings[0] = Ranking{1, 0, 0, RankingFields{50, 50, 20, 0.49, 3, 2, 1, 0, 10}}
	rankings[1] = Ranking{2, 0, 0, RankingFields{50, 50, 20, 0.51, 3, 2, 1, 0, 10}}
	rankings[2] = Ranking{3, 0, 0, RankingFields{50, 50, 19, 0.50, 3, 2, 1, 0, 10}}
	rankings[3] = Ranking{4, 0, 0, RankingFields{50, 49, 99, 0.50, 3, 2, 1, 0, 10}}
	rankings[4] = Ranking{5, 0, 0, RankingFields{49, 99, 99, 0.50, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 1, rankings[1].TeamId)
	assert.Equal(t, 3, rankings[2].TeamId)
	assert.Equal(t, 4, rankings[3].TeamId)
	assert.Equal(t, 5, rankings[4].TeamId)

	// Check with unequal numbers of matches played; every criterion is a per-match average.
	rankings = make(Rankings, 3)
	rankings[0] = Ranking{1, 0, 0, RankingFields{10, 25, 5, 0.49, 3, 2, 1, 0, 5}}
	rankings[1] = Ranking{2, 0, 0, RankingFields{19, 50, 10, 0.51, 3, 2, 1, 0, 9}}
	rankings[2] = Ranking{3, 0, 0, RankingFields{20, 50, 10, 0.51, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 3, rankings[1].TeamId)
	assert.Equal(t, 1, rankings[2].TeamId)
}

func TestSortRankingsComparesExactly(t *testing.T) {
	// Both teams average one ranking point per match. Team 1 averages 100/3 = 33.3333... points and team 2 averages
	// 3333/100 = 33.33; rounding to the 2 decimals the manual displays would make them equal, but the comparison is
	// exact, so team 1 ranks first.
	rankings := make(Rankings, 2)
	rankings[0] = Ranking{2, 0, 0, RankingFields{100, 3333, 100, 0.9, 0, 0, 0, 0, 100}}
	rankings[1] = Ranking{1, 0, 0, RankingFields{3, 100, 3, 0.1, 0, 0, 0, 0, 3}}
	sort.Sort(rankings)
	assert.Equal(t, 1, rankings[0].TeamId)
	assert.Equal(t, 2, rankings[1].TeamId)

	// The same for the auton points, once the ranking points and the scores are exactly equal on average.
	rankings = make(Rankings, 2)
	rankings[0] = Ranking{2, 0, 0, RankingFields{100, 3300, 3333, 0.9, 0, 0, 0, 0, 100}}
	rankings[1] = Ranking{1, 0, 0, RankingFields{3, 99, 100, 0.1, 0, 0, 0, 0, 3}}
	sort.Sort(rankings)
	assert.Equal(t, 1, rankings[0].TeamId)
	assert.Equal(t, 2, rankings[1].TeamId)
}

func TestGetWinRankingPoints(t *testing.T) {
	assert.Equal(t, 3, GetWinRankingPoints())
}
