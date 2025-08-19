// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game specific constants, set up as placeholders that can be easily customized.
package game

type Mayhem struct {
	AutoGamepiece1Level1Count   int
	TeleopGamepiece1Level1Count int
	AutoGamepiece1Level2Count   int
	TeleopGamepiece1Level2Count int
	AutoGamepiece2Count         int
	TeleopGamepiece2Count       int
}

const (
	LeavePoints                  = 5
	ParkPoints                   = 5
	AutoGamepiece1Level1Points   = 3
	TeleopGamepiece1Level1Points = 1
	AutoGamepiece1Level2Points   = 5
	TeleopGamepiece1Level2Points = 3
	AutoGamepiece2Points         = 4
	TeleopGamepiece2Points       = 2

	Gamepiece1BonusThreshold = 8
)
