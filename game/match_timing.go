// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

var MatchTiming = struct {
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TimeoutDurationSec          int
}{20, 3, 140, 30, 0}

func GetTeleopDurationSec() int {
	return MatchTiming.TeleopDurationSec
}

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(
		MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec,
	) * time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(
		MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+GetTeleopDurationSec(),
	) * time.Second
}
