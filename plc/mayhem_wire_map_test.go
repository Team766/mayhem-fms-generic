// Copyright 2026 Team 766. All Rights Reserved.

package plc

import (
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// Pins the frozen wire map against the numbers documented in docs/PlcWireMap.md and
// docs/agents/reference/plc.md section 3.2, so that an accidental edit to the map fails loudly.
func TestMayhemWireMapFrozenIndices(t *testing.T) {
	assert.Equal(t, uint16(32), MayhemWireMap.inputCount)
	assert.Equal(t, uint16(8), MayhemWireMap.registerCount)
	assert.Equal(t, uint16(32), MayhemWireMap.coilCount)

	assert.Equal(
		t,
		map[input]uint16{
			fieldEStop:     0,
			red1EStop:      1,
			red1AStop:      2,
			red2EStop:      3,
			red2AStop:      4,
			blue1EStop:     7,
			blue1AStop:     8,
			blue2EStop:     9,
			blue2AStop:     10,
			redConnected1:  13,
			redConnected2:  14,
			redConnected3:  15,
			blueConnected1: 16,
			blueConnected2: 17,
			blueConnected3: 18,
		},
		MayhemWireMap.inputs,
	)
	assert.Equal(
		t,
		map[input]bool{
			red3EStop:  true,
			red3AStop:  true,
			blue3EStop: true,
			blue3AStop: true,
			ftaReady:   true,
		},
		MayhemWireMap.inputDefaults,
	)

	assert.Equal(t, map[register]uint16{fieldIoConnection: 0}, MayhemWireMap.registers)

	assert.Equal(
		t,
		map[coil]uint16{
			heartbeat:        0,
			matchReset:       1,
			stackLightGreen:  2,
			stackLightOrange: 3,
			stackLightRed:    4,
			stackLightBlue:   5,
			stackLightBuzzer: 6,
			fieldResetLight:  7,
		},
		MayhemWireMap.coils,
	)
}

// Every input whose name ends in "EStop" or "AStop", plus fieldEStop, must appear in the wire map, either with a
// wire address or with an explicit default -- never neither. Iterating the enum (rather than listing the
// identifiers here) means that if upstream ever adds a new stop input, this test fails until MayhemWireMap is
// updated to say what should happen for it, instead of that input silently reading as healthy.
func TestMayhemWireMapStopInputCoverage(t *testing.T) {
	checked := 0
	for i := input(0); i < inputCount; i++ {
		name := i.String()
		if i != fieldEStop && !strings.HasSuffix(name, "EStop") && !strings.HasSuffix(name, "AStop") {
			continue
		}
		checked++
		_, hasAddress := MayhemWireMap.inputs[i]
		_, hasDefault := MayhemWireMap.inputDefaults[i]
		assert.True(
			t,
			hasAddress || hasDefault,
			"stop input %q has neither a wire address nor an explicit default in MayhemWireMap", name,
		)
		assert.False(
			t,
			hasAddress && hasDefault,
			"stop input %q has both a wire address and a default in MayhemWireMap; pick one", name,
		)
	}

	// Sanity check that the loop actually found the evergreen stop inputs, so a change to the input enum's names
	// (which would silently make the suffix check above match nothing) doesn't make this test vacuously pass.
	assert.Equal(t, 13, checked, "expected fieldEStop plus 12 team E-Stop/A-Stop inputs")
}

// Drives a real ModbusPlc with MayhemWireMap selected through a fake Modbus client, and checks that values placed
// at the frozen wire addresses land on the right logical signal (scatter), and that logical coil writes land on
// the right wire address (gather).
func TestMayhemWireMapScatterGatherRoundTrip(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}
	plc.wire = MayhemWireMap

	client.inputs[0] = true      // fieldEStop: not pressed.
	client.inputs[1] = false     // red1EStop: pressed.
	client.inputs[2] = true      // red1AStop: not pressed.
	client.inputs[3] = true      // red2EStop: not pressed.
	client.inputs[4] = true      // red2AStop: not pressed.
	client.inputs[7] = true      // blue1EStop: not pressed.
	client.inputs[8] = true      // blue1AStop: not pressed.
	client.inputs[9] = true      // blue2EStop: not pressed.
	client.inputs[10] = true     // blue2AStop: not pressed.
	client.inputs[13] = true     // redConnected1.
	client.registers[0] = 0xFFFF // fieldIoConnection: all ArmorBlocks connected.
	plc.update()

	assert.False(t, plc.GetFieldEStop())
	redEStops, blueEStops := plc.GetTeamEStops()
	redAStops, blueAStops := plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, false, false}, redEStops) // R1 pressed; R2 and the unwired R3 are not.
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{false, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	redConnected, _ := plc.GetEthernetConnected()
	assert.True(t, redConnected[0])
	assert.Equal(
		t,
		map[string]bool{"RedDs": true, "BlueDs": true, "RedIoLink": true, "BlueIoLink": true},
		plc.GetArmorBlockStatuses(),
	)

	// The hardware provides no station-3 stops and no FTA-ready switch; both read as healthy regardless of what is
	// on the wire, per their explicit defaults.
	assert.False(t, redEStops[2])
	assert.False(t, blueEStops[2])
	assert.False(t, redAStops[2])
	assert.False(t, blueAStops[2])
	assert.True(t, plc.IsFtaReady())

	// Coils 0-7 gather onto their frozen wire addresses.
	plc.SetStackLights(true, false, false, false)
	plc.update()
	assert.True(t, client.coils[4])  // stackLightRed
	assert.False(t, client.coils[5]) // stackLightBlue

	// A logical coil the map does not implement (e.g. awardsModeLight, a season/upstream coil) never reaches the
	// wire at all -- it is simply absent, not zeroed over something else.
	plc.SetAwardsModeLight(true)
	plc.update()
	assert.False(t, client.coils[8])
}

// The coil-override feature (used by the field-testing page) must still reach the wire when a wire map is
// selected, and must be applied after the map (i.e. keyed by the logical coil, same as upstream).
func TestMayhemWireMapCoilOverrideReachesWire(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}
	plc.wire = MayhemWireMap

	plc.SetFieldResetLight(false)
	plc.SetCoilOverride(int(fieldResetLight), true)
	plc.update()
	assert.True(t, client.coils[7])

	plc.ClearCoilOverride(int(fieldResetLight))
	plc.update()
	assert.False(t, client.coils[7])
}

// SetWireMap must select MayhemWireMap for "mayhem" and nil (unmodified upstream behavior) for everything else,
// including "upstream" and an unrecognized value.
func TestPlcSetWireMap(t *testing.T) {
	var plc ModbusPlc
	assert.Nil(t, plc.wire)

	plc.SetWireMap("mayhem")
	assert.Same(t, MayhemWireMap, plc.wire)

	plc.SetWireMap("upstream")
	assert.Nil(t, plc.wire)

	plc.SetWireMap("mayhem")
	assert.Same(t, MayhemWireMap, plc.wire)
	plc.SetWireMap("something-else")
	assert.Nil(t, plc.wire)
}
