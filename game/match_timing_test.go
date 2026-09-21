package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMatchTimingDefaults(t *testing.T) {
	// 15 s auto, a 3 s pause, then 120 s of teleop which covers the manual's teleop and endgame periods. The warning
	// cue marks the start of the endgame at 30 s remaining, and the Toss opens at 20 s remaining.
	assert.Equal(t, 15, MatchTiming.AutoDurationSec)
	assert.Equal(t, 3, MatchTiming.PauseDurationSec)
	assert.Equal(t, 120, MatchTiming.TeleopDurationSec)
	assert.Equal(t, 30, MatchTiming.WarningRemainingDurationSec)
	assert.Equal(t, 20, MatchTiming.TossRemainingDurationSec)
	assert.Equal(t, 0, MatchTiming.TimeoutDurationSec)

	assert.Equal(t, 15*time.Second, GetDurationToAutoEnd())
	assert.Equal(t, 18*time.Second, GetDurationToTeleopStart())
	assert.Equal(t, 138*time.Second, GetDurationToTeleopEnd())
}

func TestGetTeleopDurationSec(t *testing.T) {
	assert.Equal(t, 120, GetTeleopDurationSec())

	originalTeleopDurationSec := MatchTiming.TeleopDurationSec
	defer func() {
		MatchTiming.TeleopDurationSec = originalTeleopDurationSec
	}()

	MatchTiming.TeleopDurationSec = 106

	assert.Equal(t, 106, GetTeleopDurationSec())
}
