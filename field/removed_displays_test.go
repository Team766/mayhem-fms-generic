// Copyright 2026 Team 766. All Rights Reserved.
//
// Guards what this fork removes from upstream (see docs/DEVELOPMENT.md, "What we remove").

package field

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// An upstream sync must not bring back the Twitch display. If this fails, remove the display again rather than
// editing the test.
func TestTwitchDisplayIsGone(t *testing.T) {
	for _, name := range DisplayTypeNames {
		assert.NotContains(t, strings.ToLower(name), "twitch")
	}
	for _, path := range displayTypePaths {
		assert.NotContains(t, strings.ToLower(path), "twitch")
	}
}
