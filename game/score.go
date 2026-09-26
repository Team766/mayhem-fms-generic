// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoFloor           int
	AutoFirst           int
	AutoTop             int
	TeleopFloor         int
	TeleopFirst         int
	TeleopTop           int
	TeleopStacked       int
	Crown               CrownPlacement
	LeaveStatuses       [3]bool
	AutoBalanceStatuses [3]bool
	EndgameStatuses     [3]EndgameStatus
	Toss                bool
	Fouls               []Foul
	PlayoffDq           bool
}

// Game-specific settings that can be changed via the settings.
var AutonRpThreshold = 20
var ScoringRpThreshold = 12

// Represents where on the field the dragon's crown ended the match, if the alliance scored it. The crown is counted in
// the ordinary counter for its location like any other treasure; this selection adds that same value again as a bonus.
type CrownPlacement int

const (
	CrownNone CrownPlacement = iota
	CrownAutoFloor
	CrownAutoFirst
	CrownAutoTop
	CrownTeleopFloor
	CrownTeleopFirst
	CrownTeleopTop
	CrownTeleopStacked
)

// Returns the bonus that the crown adds to the alliance's score, which is the value of the placement it made.
func (placement CrownPlacement) PointValue() int {
	switch placement {
	case CrownAutoFloor:
		return 4
	case CrownAutoFirst:
		return 8
	case CrownAutoTop:
		return 12
	case CrownTeleopFloor:
		return 2
	case CrownTeleopFirst:
		return 5
	case CrownTeleopTop:
		return 10
	case CrownTeleopStacked:
		return 8
	default:
		return 0
	}
}

// Returns true if the crown was placed during the autonomous period, meaning that its bonus belongs to the auto
// treasure points and counts toward the Auton ranking point.
func (placement CrownPlacement) IsAuto() bool {
	return placement == CrownAutoFloor || placement == CrownAutoFirst || placement == CrownAutoTop
}

// Returns the name of the placement, for display.
func (placement CrownPlacement) String() string {
	switch placement {
	case CrownAutoFloor:
		return "Auto Floor"
	case CrownAutoFirst:
		return "Auto First"
	case CrownAutoTop:
		return "Auto Top"
	case CrownTeleopFloor:
		return "Teleop Floor"
	case CrownTeleopFirst:
		return "Teleop First"
	case CrownTeleopTop:
		return "Teleop Top"
	case CrownTeleopStacked:
		return "Teleop Stacked"
	default:
		return "None"
	}
}

// Represents where a robot ended the match: nowhere in particular, parked in its own safe house, or balanced on its own
// mountain top. Park and balance are mutually exclusive.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgamePark
	EndgameBalance
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)
	summary.PlayoffDq = score.PlayoffDq

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points. Leaving the safe house and balancing on the mountain top are independent
	// per-robot awards; a robot that does both earns 4 + 12 = 16.
	for _, leave := range score.LeaveStatuses {
		if leave {
			summary.LeavePoints += 4
		}
	}
	for _, balanced := range score.AutoBalanceStatuses {
		if balanced {
			summary.AutoBalancePoints += 12
		}
	}
	summary.AutoTreasurePoints = 4*score.AutoFloor + 8*score.AutoFirst + 12*score.AutoTop

	// Calculate teleoperated period points. The teleop counters also cover the endgame period, which has no placement
	// values of its own.
	summary.TeleopTreasurePoints =
		2*score.TeleopFloor + 5*score.TeleopFirst + 10*score.TeleopTop + 8*score.TeleopStacked

	// The dragon's crown scores double. It is already counted in the ordinary counter for its location, so the bonus
	// adds that same value again, to the period in which the crown was placed.
	summary.Crown = score.Crown
	summary.CrownBonusPoints = score.Crown.PointValue()
	if score.Crown.IsAuto() {
		summary.AutoTreasurePoints += summary.CrownBonusPoints
	} else {
		// CrownNone is worth zero, so adding it here is harmless.
		summary.TeleopTreasurePoints += summary.CrownBonusPoints
	}

	summary.AutonPoints = summary.LeavePoints + summary.AutoBalancePoints + summary.AutoTreasurePoints

	// Calculate endgame points, which are also per robot.
	numRobotsBalanced := 0
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgamePark:
			summary.EndgamePoints += 2
		case EndgameBalance:
			summary.EndgamePoints += 12
			numRobotsBalanced++
		default:
		}
	}

	// The Toss is worth 2 points, once per alliance whatever the alliance size.
	if score.Toss {
		summary.TossPoints = 2
	}

	summary.MatchPoints =
		summary.AutonPoints + summary.TeleopTreasurePoints + summary.EndgamePoints + summary.TossPoints

	// Count treasures for the live displays and for the Scoring ranking point. The crown is one treasure and is counted
	// once, by the counter it was entered in.
	summary.TreasureCount = score.AutoFloor + score.AutoFirst + score.AutoTop + score.TeleopFloor +
		score.TeleopFirst + score.TeleopTop + score.TeleopStacked
	summary.ShelfTreasureCount = score.TeleopFirst + score.TeleopTop + score.TeleopStacked
	summary.ShelfTreasureGoal = ScoringRpThreshold

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Auton ranking point: at least the threshold in autonomous points, or the opponent entered the alliance's safe
	// house or safe zone during auto (MA2603).
	summary.AutonRankingPointByFoul = opponentScore.hasFoulForRule("MA2603")
	summary.AutonRankingPoint = summary.AutonPoints >= AutonRpThreshold || summary.AutonRankingPointByFoul

	// Scoring ranking point: at least the threshold in treasures placed during teleop on the first shelf, on the top
	// shelf, or stacked. There is no opponent-violation alternative.
	summary.ScoringRankingPoint = summary.ShelfTreasureCount >= ScoringRpThreshold

	// Endgame ranking point: at least one robot balanced on its mountain top, or the opponent contacted the alliance's
	// balance beam (MA2601) or a robot on its own beam (MA2602) during endgame. Parking does not count.
	summary.EndgameRankingPointByFoul = opponentScore.hasFoulForRule("MA2601", "MA2602")
	summary.EndgameRankingPoint = numRobotsBalanced > 0 || summary.EndgameRankingPointByFoul

	// Add up the bonus ranking points.
	if summary.AutonRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.ScoringRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.EndgameRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Returns true if the alliance committed at least one foul for any of the given rule numbers. A foul entered without a
// rule never satisfies this.
func (score *Score) hasFoulForRule(ruleNumbers ...string) bool {
	for _, foul := range score.Fouls {
		rule := foul.Rule()
		if rule == nil {
			continue
		}
		for _, ruleNumber := range ruleNumbers {
			if rule.RuleNumber == ruleNumber {
				return true
			}
		}
	}
	return false
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.AutoFloor != other.AutoFloor ||
		score.AutoFirst != other.AutoFirst ||
		score.AutoTop != other.AutoTop ||
		score.TeleopFloor != other.TeleopFloor ||
		score.TeleopFirst != other.TeleopFirst ||
		score.TeleopTop != other.TeleopTop ||
		score.TeleopStacked != other.TeleopStacked ||
		score.Crown != other.Crown ||
		score.LeaveStatuses != other.LeaveStatuses ||
		score.AutoBalanceStatuses != other.AutoBalanceStatuses ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.Toss != other.Toss ||
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
