// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

// create constants for each of the scoring elements
const (
	LeaveNonePoints        = 0
	LeavePartialPoints     = 3
	LeaveFullPoints        = 6
	GamePiece1AutonPoints  = 5
	GamePiece2AutonPoints  = 7
	GamePiece1TeleopPoints = 6
	GamePiece2TeleopPoints = 8
	EndgameNonePoints      = 0
	EndgamePartialPoints   = 2
	EndgameFullPoints      = 5
)

// LeaveStatus represents the autonomous performance of a robot.
type LeaveStatus int

const (
	LeaveNone LeaveStatus = iota
	LeavePartial
	LeaveFull
)

// EndgameStatus represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgamePartial
	EndgameFull
)

type Score struct {
	LeaveStatuses    [3]LeaveStatus
	GamePiece1Auton  int
	GamePiece1Teleop int
	GamePiece2Auton  int
	GamePiece2Teleop int
	EndgameStatuses  [3]EndgameStatus
	Fouls            []Foul
	PlayoffDq        bool
}

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	summary.LeavePoints = 0
	for _, status := range score.LeaveStatuses {
		switch status {
		case LeavePartial:
			summary.LeavePoints += LeavePartialPoints
		case LeaveFull:
			summary.LeavePoints += LeaveFullPoints
		default:
		}
	}

	// Calculate Game Piece 1 points.
	summary.GamePiece1Points = (score.GamePiece1Auton * GamePiece1AutonPoints) + (score.GamePiece1Teleop * GamePiece1TeleopPoints)

	// Calculate Game Piece 2 points.
	summary.GamePiece2Points = (score.GamePiece2Auton * GamePiece2AutonPoints) + (score.GamePiece2Teleop * GamePiece2TeleopPoints)

	// Calculate endgame points.
	summary.EndgamePoints = 0
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgamePartial:
			summary.EndgamePoints += EndgamePartialPoints
		case EndgameFull:
			summary.EndgamePoints += EndgameFullPoints
		default:
		}
	}

	summary.MatchPoints = summary.LeavePoints + summary.GamePiece1Points + summary.GamePiece2Points + summary.EndgamePoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}

		// TODO: Review foul.Rule() and IsRankingPoint for generic game
		/* rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				// This section will need to be updated for generic ranking points
			}
		} */
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	// TODO: This section needs to be updated based on generic game ranking rules.
	// For now, bonus ranking points will be 0 or based on simple criteria.
	summary.BonusRankingPoints = 0

	// Example: A simple RP for achieving a certain score
	// if summary.MatchPoints >= 50 { // Arbitrary threshold
	// 	summary.BonusRankingPoints++
	// }

	// Example: A simple RP for all robots achieving Full Leave
	// allRobotsFullLeave := true
	// for _, status := range score.AutonStatuses {
	// 	if status != LeaveFull {
	// 		allRobotsFullLeave = false
	// 		break
	// 	}
	// }
	// if allRobotsFullLeave {
	// 	summary.BonusRankingPoints++
	// }

	return summary
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.LeaveStatuses != other.LeaveStatuses ||
		score.GamePiece1Auton != other.GamePiece1Auton ||
		score.GamePiece1Teleop != other.GamePiece1Teleop ||
		score.GamePiece2Auton != other.GamePiece2Auton ||
		score.GamePiece2Teleop != other.GamePiece2Teleop ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.PlayoffDq != other.PlayoffDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
