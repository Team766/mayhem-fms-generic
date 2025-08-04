// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{true, 25, 16},
		{false, 1868, 13},
		{false, 1868, 13},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
	}
	return &Score{
		LeaveStatuses:    [3]LeaveStatus{LeaveFull, LeavePartial, LeaveNone},
		GamePiece1Auton:  2,
		GamePiece1Teleop: 5,
		GamePiece2Auton:  1,
		GamePiece2Teleop: 4,
		EndgameStatuses:  [3]EndgameStatus{EndgamePartial, EndgameNone, EndgameFull},
		Fouls:            fouls,
		PlayoffDq:        false,
	}
}

func TestScore2() *Score {
	return &Score{
		LeaveStatuses:    [3]LeaveStatus{LeaveNone, LeaveFull, LeaveNone},
		GamePiece1Auton:  4,
		GamePiece1Teleop: 8,
		GamePiece2Auton:  5,
		GamePiece2Teleop: 10,
		EndgameStatuses:  [3]EndgameStatus{EndgameFull, EndgamePartial, EndgameFull},
		Fouls:            []Foul{},
		PlayoffDq:        false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}
