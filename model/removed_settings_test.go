// Copyright 2026 Team 766. All Rights Reserved.
//
// Guards what this fork removes from upstream (see docs/DEVELOPMENT.md, "What we remove").

package model

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

// An upstream sync must not bring back settings for the integrations this fork removes. If this fails, a removed
// integration has leaked back in: remove it again rather than editing the list.
func TestEventSettingsHasNoRemovedIntegrations(t *testing.T) {
	removedPrefixes := []string{"Tba", "Nexus", "TeamSign", "Led", "Twitch"}

	settingsType := reflect.TypeOf(EventSettings{})
	for i := 0; i < settingsType.NumField(); i++ {
		name := settingsType.Field(i).Name
		for _, prefix := range removedPrefixes {
			assert.False(t, strings.HasPrefix(name, prefix), "EventSettings.%s belongs to a removed integration", name)
		}
	}
}
