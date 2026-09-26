// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific fields by which teams are ranked and the logic for sorting rankings.

package game

import "math/rand"

type RankingFields struct {
	RankingPoints     int
	ScorePoints       int
	AutonPoints       int
	Random            float64
	Wins              int
	Losses            int
	Ties              int
	Disqualifications int
	Played            int
}

type Ranking struct {
	TeamId       int `db:"id,manual"`
	Rank         int
	PreviousRank int
	RankingFields
}

type Rankings []Ranking

var RankingRandomFloat64 = rand.Float64

func GetWinRankingPoints() int {
	return 3
}

func (fields *RankingFields) AddScoreSummary(ownScore *ScoreSummary, opponentScore *ScoreSummary, disqualified bool) {
	fields.Played += 1

	// Store a random value to be used as the last tiebreaker if necessary.
	fields.Random = RankingRandomFloat64()

	if disqualified {
		// Don't award any points.
		fields.Disqualifications += 1
		return
	}

	// Assign ranking points and wins/losses/ties.
	if ownScore.Score > opponentScore.Score {
		fields.RankingPoints += GetWinRankingPoints()
		fields.Wins += 1
	} else if ownScore.Score == opponentScore.Score {
		fields.RankingPoints += 1
		fields.Ties += 1
	} else {
		fields.Losses += 1
	}
	fields.RankingPoints += ownScore.BonusRankingPoints

	// Assign tiebreaker points. The first tiebreaker is the average match score, which includes the foul points the
	// alliance received; the second is the average autonomous score.
	fields.ScorePoints += ownScore.Score
	fields.AutonPoints += ownScore.AutonPoints
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Len() int {
	return len(rankings)
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Less(i, j int) bool {
	a := rankings[i]
	b := rankings[j]

	// Every criterion is a per-match average; use cross-multiplication to keep it in integer math, so that rounding
	// for display can never create or break a tie.
	if a.RankingPoints*b.Played == b.RankingPoints*a.Played {
		if a.ScorePoints*b.Played == b.ScorePoints*a.Played {
			if a.AutonPoints*b.Played == b.AutonPoints*a.Played {
				return a.Random > b.Random
			}
			return a.AutonPoints*b.Played > b.AutonPoints*a.Played
		}
		return a.ScorePoints*b.Played > b.ScorePoints*a.Played
	}
	return a.RankingPoints*b.Played > b.RankingPoints*a.Played
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Swap(i, j int) {
	rankings[i], rankings[j] = rankings[j], rankings[i]
}
