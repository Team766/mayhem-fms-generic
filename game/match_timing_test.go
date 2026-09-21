package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTeleopDurationSec(t *testing.T) {
	assert.Equal(t, 140, GetTeleopDurationSec())

	originalTeleopDurationSec := MatchTiming.TeleopDurationSec
	defer func() {
		MatchTiming.TeleopDurationSec = originalTeleopDurationSec
	}()

	MatchTiming.TeleopDurationSec = 106

	assert.Equal(t, 106, GetTeleopDurationSec())
}
