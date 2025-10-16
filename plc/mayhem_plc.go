package plc

type MayhemPlc struct {
	*ModbusPlc
}

// NewMayhemPlc creates a new MayhemPlc with custom pin mappings.
func NewMayhemPlc() *MayhemPlc {
	return &MayhemPlc{
		ModbusPlc: NewModbusPlc(),
	}
}

// getInputPin returns the physical pin number for a logical input.
// This overrides the default 1:1 mapping with Mayhem-specific mappings.
func (plc *MayhemPlc) getInputPin(in input) int {
	switch in {
	// Input mappings - adjust these to match your actual hardware configuration
	case fieldEStop:
		return 0
	case red1EStop:
		return 1
	case red1AStop:
		return 2
	case red2EStop:
		return 3
	case red2AStop:
		return 4
	case red3EStop:
		return 5
	case red3AStop:
		return 6
	case blue1EStop:
		return 7
	case blue1AStop:
		return 8
	case blue2EStop:
		return 9
	case blue2AStop:
		return 10
	case blue3EStop:
		return 11
	case blue3AStop:
		return 12
	case redConnected1:
		return 13
	case redConnected2:
		return 14
	case redConnected3:
		return 15
	case blueConnected1:
		return 16
	case blueConnected2:
		return 17
	case blueConnected3:
		return 18
	default:
		// Fall back to default implementation for any inputs we don't explicitly map
		return plc.ModbusPlc.getInputPin(in)
	}
}

// getCoilPin returns the physical pin number for a logical coil.
// This overrides the default 1:1 mapping with Mayhem-specific mappings.
func (plc *MayhemPlc) getCoilPin(c coil) int {
	switch c {
	// Coil mappings - adjust these to match your actual hardware configuration
	case heartbeat:
		return 0
	case matchReset:
		return 1
	case stackLightGreen:
		return 2
	case stackLightOrange:
		return 3
	case stackLightRed:
		return 4
	case stackLightBlue:
		return 5
	case stackLightBuzzer:
		return 6
	case fieldResetLight:
		return 7
	case redTrussLightOuter:
		return 8
	case redTrussLightMiddle:
		return 9
	case redTrussLightInner:
		return 10
	case blueTrussLightOuter:
		return 11
	case blueTrussLightMiddle:
		return 12
	case blueTrussLightInner:
		return 13
	default:
		// Fall back to default implementation for any coils we don't explicitly map
		return plc.ModbusPlc.getCoilPin(c)
	}
}
