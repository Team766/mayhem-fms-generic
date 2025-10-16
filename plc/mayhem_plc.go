package plc

type MayhemPlc struct {
	*ModbusPlc
}

// NewMayhemPlc creates a new MayhemPlc with custom pin mappings.
// It panics if the pin mappings are invalid.
func NewMayhemPlc() *MayhemPlc {
	// Define Mayhem-specific pin mappings
	inputMap := InputMap{
		// Input mappings - adjust these to match your actual hardware configuration
		fieldEStop:     0,
		red1EStop:      1,
		red1AStop:      2,
		red2EStop:      3,
		red2AStop:      4,
		red3EStop:      5,
		red3AStop:      6,
		blue1EStop:     7,
		blue1AStop:     8,
		blue2EStop:     9,
		blue2AStop:     10,
		blue3EStop:     11,
		blue3AStop:     12,
		redConnected1:  13,
		redConnected2:  14,
		redConnected3:  15,
		blueConnected1: 16,
		blueConnected2: 17,
		blueConnected3: 18,
	}

	coilMap := CoilMap{
		// Coil mappings - adjust these to match your actual hardware configuration
		heartbeat:            0,
		matchReset:           1,
		stackLightGreen:      2,
		stackLightOrange:     3,
		stackLightRed:        4,
		stackLightBlue:       5,
		stackLightBuzzer:     6,
		fieldResetLight:      7,
		redTrussLightOuter:   8,
		redTrussLightMiddle:  9,
		redTrussLightInner:   10,
		blueTrussLightOuter:  11,
		blueTrussLightMiddle: 12,
		blueTrussLightInner:  13,
	}

	return &MayhemPlc{
		ModbusPlc: NewModbusPlcWithMaps(inputMap, coilMap),
	}
}
