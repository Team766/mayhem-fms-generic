// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

// A well-populated score: every field is set to something other than its zero value. Rule 17 is MA2604 (major), rule 29
// is MA2616 (minor) and rule 6 is G415 (major); none of them grants the opposing alliance a ranking point.
// Match points: leave 8 + auto balance 12 + auto treasure 32 + teleop treasure 64 + endgame 14 + toss 2 = 132.
func TestScore1() *Score {
	fouls := []Foul{
		{1, true, 25, 17},
		{2, false, 1868, 29},
		{3, false, 1868, 29},
		{4, true, 25, 6},
		{5, true, 25, 6},
		{6, true, 25, 6},
		{7, true, 25, 6},
	}
	return &Score{
		AutoFloor:           1,
		AutoFirst:           2,
		AutoTop:             1,
		TeleopFloor:         3,
		TeleopFirst:         4,
		TeleopTop:           2,
		TeleopStacked:       1,
		Crown:               CrownTeleopTop,
		LeaveStatuses:       [3]bool{true, true, false},
		AutoBalanceStatuses: [3]bool{false, true, false},
		EndgameStatuses:     [3]EndgameStatus{EndgameBalance, EndgamePark, EndgameNone},
		Toss:                true,
		Fouls:               fouls,
		PlayoffDq:           false,
	}
}

// A more modest score with no fouls and no crown.
// Match points: leave 4 + auto treasure 8 + teleop treasure 29 + endgame 2 = 43.
func TestScore2() *Score {
	return &Score{
		AutoFirst:       1,
		TeleopFloor:     2,
		TeleopFirst:     3,
		TeleopTop:       1,
		LeaveStatuses:   [3]bool{true, false, false},
		EndgameStatuses: [3]EndgameStatus{EndgamePark, EndgameNone, EndgameNone},
		Fouls:           []Foul{},
		PlayoffDq:       false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 180, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 220, 0.1114, 1, 3, 2, 0, 10}}
}
