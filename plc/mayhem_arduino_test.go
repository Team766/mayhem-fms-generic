// Copyright 2026 Team 766. All Rights Reserved.
//
// Guard for the M-Ayhem Arduino PLC, which is flashed with fixed Modbus addresses that match the signal order below.

package plc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// The M-Ayhem field's Arduino PLC (fakeplc-arduino) serves fixed Modbus addresses and cannot follow a renumbering. If
// this test fails after an upstream sync, upstream has moved a signal the hardware depends on: do not change the
// expected numbers here; keep the generic signals in this order (see docs/DEVELOPMENT.md, "PLC") or update the firmware.
func TestMayhemArduinoSignalAddresses(t *testing.T) {
	inputs := []input{
		fieldEStop, red1EStop, red1AStop, red2EStop, red2AStop, red3EStop, red3AStop, blue1EStop, blue1AStop,
		blue2EStop, blue2AStop, blue3EStop, blue3AStop, redConnected1, redConnected2, redConnected3, blueConnected1,
		blueConnected2, blueConnected3,
	}
	for address, signal := range inputs {
		assert.Equal(t, address, int(signal), "input %s moved", signal.String())
	}

	assert.Equal(t, 0, int(fieldIoConnection))

	coils := []coil{
		heartbeat, matchReset, stackLightGreen, stackLightOrange, stackLightRed, stackLightBlue, stackLightBuzzer,
		fieldResetLight,
	}
	for address, signal := range coils {
		assert.Equal(t, address, int(signal), "coil %s moved", signal.String())
	}

	// The firmware serves 32 discrete inputs, 8 holding registers and 32 coils. A Modbus request that runs past the
	// end of a table is rejected as a whole, which would show the PLC as unhealthy.
	assert.LessOrEqual(t, int(inputCount), 32)
	assert.LessOrEqual(t, int(registerCount), 8)
	assert.LessOrEqual(t, int(coilCount), 32)
}
