// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUniqueMatchSounds(t *testing.T) {
	UpdateMatchSounds()

	uniqueSounds := UniqueMatchSounds()

	assert.Equal(
		t,
		[]string{
			"start",
			"end",
			"resume",
			"warning",
			"toss",
			"abort",
			"match_result",
			"pick_clock",
			"pick_clock_expired",
			"field_reset",
		},
		matchSoundNames(uniqueSounds),
	)
	assert.Len(t, uniqueSounds, 10)
	assert.Same(t, MatchSounds[0], uniqueSounds[0])
	assert.Same(t, MatchSounds[1], uniqueSounds[1])
	assert.Same(t, MatchSounds[3], uniqueSounds[3])
}

func TestMatchSoundTimes(t *testing.T) {
	UpdateMatchSounds()

	// Seconds into the match, with the default timing: auto ends at 15, teleop starts at 18 and ends at 138. The
	// warning marks the start of the endgame at 30 s remaining and the Toss cue sounds at 20 s remaining.
	expectedTimes := []struct {
		name         string
		matchTimeSec float64
	}{
		{"start", 0},
		{"end", 15},
		{"resume", 18},
		{"warning", 108},
		{"toss", 118},
		{"end", 138},
		{"abort", -1},
		{"match_result", -1},
		{"pick_clock", -1},
		{"pick_clock_expired", -1},
		{"field_reset", -1},
	}

	if assert.Len(t, MatchSounds, len(expectedTimes)) {
		for i, expected := range expectedTimes {
			assert.Equal(t, expected.name, MatchSounds[i].Name, "sound %d", i)
			assert.Equal(t, expected.matchTimeSec, MatchSounds[i].MatchTimeSec, "sound %s", expected.name)
			assert.Equal(t, "wav", MatchSounds[i].FileExtension)
		}
	}
}

func matchSoundNames(matchSounds []*MatchSound) []string {
	names := make([]string, 0, len(matchSounds))
	for _, sound := range matchSounds {
		names = append(names, sound.Name)
	}
	return names
}
