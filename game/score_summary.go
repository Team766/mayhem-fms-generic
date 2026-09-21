// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	LeavePoints               int
	AutoBalancePoints         int
	AutoTreasurePoints        int
	AutonPoints               int
	TeleopTreasurePoints      int
	EndgamePoints             int
	TossPoints                int
	Crown                     CrownPlacement
	CrownBonusPoints          int
	TreasureCount             int
	ShelfTreasureCount        int
	ShelfTreasureGoal         int
	MatchPoints               int
	PostMatchPoints           int
	FoulPoints                int
	Score                     int
	PlayoffDq                 bool
	AutonRankingPoint         bool
	ScoringRankingPoint       bool
	EndgameRankingPoint       bool
	AutonRankingPointByFoul   bool
	EndgameRankingPointByFoul bool
	BonusRankingPoints        int
	NumOpponentMajorFouls     int
}

type MatchStatus int

const (
	MatchScheduled MatchStatus = iota
	MatchHidden
	RedWonMatch
	BlueWonMatch
	TieMatch
)

func (t MatchStatus) Get() MatchStatus {
	return t
}

// Determines the winner of the match given the score summaries for both alliances, and returns a display string
// indicating the playoff tiebreaker criterion used if the primary score is tied.
func DetermineMatchStatus(
	redScoreSummary, blueScoreSummary *ScoreSummary,
	applyPlayoffTiebreakers bool,
) (MatchStatus, string) {
	if redScoreSummary.PlayoffDq != blueScoreSummary.PlayoffDq {
		if redScoreSummary.PlayoffDq {
			return BlueWonMatch, ""
		}
		return RedWonMatch, ""
	}

	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status, ""
	}

	if applyPlayoffTiebreakers {
		// Check scoring breakdowns to resolve playoff ties. The alliance that committed fewer major fouls wins, which
		// is the alliance with the higher count of major fouls committed by its opponent.
		if status := comparePoints(
			redScoreSummary.NumOpponentMajorFouls, blueScoreSummary.NumOpponentMajorFouls,
		); status != TieMatch {
			return status, "TIEBREAK: MAJOR FOULS"
		}
		if status := comparePoints(redScoreSummary.AutonPoints, blueScoreSummary.AutonPoints); status != TieMatch {
			return status, "TIEBREAK: AUTON POINTS"
		}
		if status := comparePoints(redScoreSummary.MatchPoints, blueScoreSummary.MatchPoints); status != TieMatch {
			return status, "TIEBREAK: MATCH POINTS"
		}
		return TieMatch, "TRUE TIE"
	}

	return TieMatch, ""
}

// Helper method to compare the red and blue alliance point totals and return the appropriate MatchStatus.
func comparePoints(redPoints, bluePoints int) MatchStatus {
	if redPoints > bluePoints {
		return RedWonMatch
	}
	if redPoints < bluePoints {
		return BlueWonMatch
	}
	return TieMatch
}
