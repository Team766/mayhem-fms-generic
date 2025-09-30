// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotsBypassed [3]bool
	Mayhem         Mayhem
	Fouls          []Foul
	PlayoffDq      bool
}

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, status := range score.Mayhem.LeaveStatuses {
		if status {
			summary.LeavePoints += LeavePoints
		}
	}

	summary.AutoPoints = summary.LeavePoints +
		score.Mayhem.AutoGamepiece1Level1Count*AutoGamepiece1Level1Points +
		score.Mayhem.AutoGamepiece1Level2Count*AutoGamepiece1Level2Points +
		score.Mayhem.AutoGamepiece2Count*AutoGamepiece2Points

	summary.NumGamepiece1 = score.Mayhem.AutoGamepiece1Level1Count + score.Mayhem.AutoGamepiece1Level2Count +
		score.Mayhem.TeleopGamepiece1Level1Count + score.Mayhem.TeleopGamepiece1Level2Count

	summary.Gamepiece1Points = score.Mayhem.AutoGamepiece1Level1Count*AutoGamepiece1Level1Points +
		score.Mayhem.AutoGamepiece1Level2Count*AutoGamepiece1Level2Points +
		score.Mayhem.TeleopGamepiece1Level1Count*TeleopGamepiece1Level1Points +
		score.Mayhem.TeleopGamepiece1Level2Count*TeleopGamepiece1Level2Points

	summary.NumGamepiece2 = score.Mayhem.AutoGamepiece2Count + score.Mayhem.TeleopGamepiece2Count

	summary.Gamepiece2Points = score.Mayhem.AutoGamepiece2Count*AutoGamepiece2Points +
		score.Mayhem.TeleopGamepiece2Count*TeleopGamepiece2Points

	// Calculate park points.
	for _, status := range score.Mayhem.ParkStatuses {
		if status {
			summary.ParkPoints += ParkPoints
		}
	}

	summary.MatchPoints = summary.LeavePoints + summary.Gamepiece1Points + summary.Gamepiece2Points + summary.ParkPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	// Leave Bonus RP
	allRobotsLeft := true
	for i, left := range score.Mayhem.LeaveStatuses {
		if !left && !score.RobotsBypassed[i] {
			allRobotsLeft = false
			break
		}
	}
	if allRobotsLeft {
		summary.LeaveBonusRankingPoint = true
	}

	// Gamepiece 1 Bonus RP
	if summary.NumGamepiece1 >= Gamepiece1RPThreshold {
		summary.Gamepiece1BonusRankingPoint = true
	}

	// Park Bonus RP
	allRobotsParked := true
	for i, parked := range score.Mayhem.ParkStatuses {
		if !parked && !score.RobotsBypassed[i] {
			allRobotsParked = false
			break
		}
	}
	if allRobotsParked {
		summary.ParkBonusRankingPoint = true
	}

	// Add up the bonus ranking points.
	if summary.LeaveBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.Gamepiece1BonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.ParkBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.Mayhem.LeaveStatuses != other.Mayhem.LeaveStatuses ||
		score.Mayhem.AutoGamepiece1Level1Count != other.Mayhem.AutoGamepiece1Level1Count ||
		score.Mayhem.TeleopGamepiece1Level1Count != other.Mayhem.TeleopGamepiece1Level1Count ||
		score.Mayhem.AutoGamepiece1Level2Count != other.Mayhem.AutoGamepiece1Level2Count ||
		score.Mayhem.TeleopGamepiece1Level2Count != other.Mayhem.TeleopGamepiece1Level2Count ||
		score.Mayhem.AutoGamepiece2Count != other.Mayhem.AutoGamepiece2Count ||
		score.Mayhem.TeleopGamepiece2Count != other.Mayhem.TeleopGamepiece2Count ||
		score.Mayhem.ParkStatuses != other.Mayhem.ParkStatuses ||
		score.RobotsBypassed != other.RobotsBypassed ||
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
