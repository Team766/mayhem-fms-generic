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

		RobotsBypassed: [3]bool{false, false, true},
		LeaveStatuses:  [3]bool{true, true, false},

		Mayhem: Mayhem{
			AutoGamepiece1Level1Count:   1,
			TeleopGamepiece1Level1Count: 2,
			AutoGamepiece1Level2Count:   2,
			TeleopGamepiece1Level2Count: 4,
			AutoGamepiece2Count:         2,
			TeleopGamepiece2Count:       4,
		},
		ParkStatuses: [3]bool{true, true, false},
		Fouls:        fouls,
		PlayoffDq:    false,
	}
}

func TestScore2() *Score {
	return &Score{

		RobotsBypassed: [3]bool{false, false, false},
		LeaveStatuses:  [3]bool{false, true, false},
		Mayhem: Mayhem{
			AutoGamepiece1Level1Count:   2,
			TeleopGamepiece1Level1Count: 4,
			AutoGamepiece1Level2Count:   3,
			TeleopGamepiece1Level2Count: 5,
			AutoGamepiece2Count:         1,
			TeleopGamepiece2Count:       4,
		},
		ParkStatuses: [3]bool{false, true, false},
		Fouls:        []Foul{},
		PlayoffDq:    false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{TeamId: 254, Rank: 1, PreviousRank: 0, RankingFields: RankingFields{RankingPoints: 20, MatchPoints: 625, AutoPoints: 90, Gamepiece2Points: 40, Wins: 3, Losses: 2, Ties: 1, Disqualifications: 0, Played: 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{TeamId: 1114, Rank: 2, PreviousRank: 1, RankingFields: RankingFields{RankingPoints: 18, MatchPoints: 700, AutoPoints: 100, Gamepiece2Points: 50, Wins: 1, Losses: 3, Ties: 2, Disqualifications: 0, Played: 10}}
}
