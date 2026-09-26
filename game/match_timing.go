// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

// TossRemainingDurationSec is the point in the match, given as seconds remaining, at which the human players may throw
// their marked toss cube into the treasure chest. It is announced by the "toss" sound.
var MatchTiming = struct {
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TossRemainingDurationSec    int
	TimeoutDurationSec          int
}{15, 3, 120, 30, 20, 0}

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
