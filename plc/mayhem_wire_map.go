// Copyright 2026 Team 766. All Rights Reserved.
//
// The frozen M-Ayhem wire map: translates between upstream's logical input/register/coil identifiers and the
// fixed Modbus addresses wired into Team 766's Arduino "fakeplc" hardware (fakeplc-arduino, branch
// plc-cheesy-arena-compat). See docs/PlcWireMap.md for the tables this file implements, and
// docs/agents/reference/plc.md section 3.2 for how it was derived.
//
// This file only ever translates at the wire boundary (readInputs, readRegisters, writeCoils in plc.go); it does
// not change any logical behavior. It is keyed by identifier, not by number, so an upstream renumber of the
// input/register/coil enums is harmless (the map just keeps pointing at the same signal) and an upstream rename
// or removal of an identifier this file references is a compile error, not a silent runtime bug.

package plc

// Fixed sizes of the tables exchanged with the Arduino over the wire. These are independent of
// inputCount/registerCount/coilCount (upstream's own enum lengths), so upstream growing its I/O tables in a
// future year cannot change what is sent to, or expected back from, this specific piece of hardware. They must
// never exceed what the firmware serves (fakeplc-mega configures 32 discrete inputs, 8 holding registers and 32
// coils): a Modbus request past the end of a table is rejected as a whole, which would mark the PLC unhealthy.
const (
	mayhemWireInputCount    = 32
	mayhemWireRegisterCount = 8
	mayhemWireCoilCount     = 32
)

// wireMap translates between upstream's logical input/register/coil identifiers and fixed physical wire
// addresses for one piece of PLC hardware. A nil *wireMap on a ModbusPlc means "no translation": talk unmodified
// upstream Modbus, at upstream's own table sizes.
type wireMap struct {
	// inputCount, registerCount, and coilCount are the sizes of the tables exchanged with the wire, independent of
	// the upstream enum lengths.
	inputCount, registerCount, coilCount uint16

	// inputs maps a logical input to the wire discrete-input address that provides it. An input absent from this
	// map is never read from the wire; its value comes from inputDefaults instead.
	inputs map[input]uint16

	// inputDefaults gives the raw value to use, in place of a wire reading, for every input the hardware does not
	// provide. It is keyed the same as `inputs`, and every input must appear in exactly one of the two maps (see
	// TestMayhemWireMapStopInputCoverage): there is no implicit fallback. The value here is the raw bit that would
	// have come off the wire -- upstream's accessors (GetFieldEStop, GetTeamEStops, GetTeamAStops) then negate it,
	// so `true` here reads as "not stopped" / healthy, matching how upstream interprets an active-low, wired input.
	inputDefaults map[input]bool

	// registers maps a logical register to the wire holding-register address that provides it. A register absent
	// from this map is never read from the wire and stays at its zero value.
	registers map[register]uint16

	// coils maps a logical coil to the wire coil address it is written to. A coil absent from this map is simply
	// never sent to the wire.
	coils map[coil]uint16
}

// MayhemWireMap is the frozen wire map for the M-Ayhem Arduino PLC. It implements only what the firmware
// (fakeplc-mega, branch plc-cheesy-arena-compat) actually wires up: discrete inputs 0-19, holding register 0, and
// coils 0-7. Everything else -- including the per-station light coils reserved from wire coil 16 up for the
// feature described in docs/agents/reference/plc.md section 4 -- is deliberately left unimplemented rather than
// guessed at.
var MayhemWireMap = &wireMap{
	inputCount:    mayhemWireInputCount,
	registerCount: mayhemWireRegisterCount,
	coilCount:     mayhemWireCoilCount,

	inputs: map[input]uint16{
		fieldEStop: 0,
		red1EStop:  1,
		red1AStop:  2,
		red2EStop:  3,
		red2AStop:  4,
		// red3EStop, red3AStop: no wiring for station 3 -- see inputDefaults below.
		blue1EStop: 7,
		blue1AStop: 8,
		blue2EStop: 9,
		blue2AStop: 10,
		// blue3EStop, blue3AStop: no wiring for station 3 -- see inputDefaults below.
		redConnected1:  13,
		redConnected2:  14,
		redConnected3:  15,
		blueConnected1: 16,
		blueConnected2: 17,
		blueConnected3: 18,
		// ftaReady: no FTA switch wired up -- see inputDefaults below.
	},
	inputDefaults: map[input]bool{
		// The field has no station-3 E-stop/A-stop wiring (only four driver stations are populated: Red1, Red2,
		// Blue1, Blue2). Per docs/BASE.md's PLC contract, a signal the hardware does not provide reads as healthy,
		// so station 3 reads as "not stopped" rather than permanently e-stopped.
		red3EStop:  true,
		red3AStop:  true,
		blue3EStop: true,
		blue3AStop: true,
		// The Arduino has no FTA ready dead-man switch wired up. Per the same contract, it reads as "ready" rather
		// than blocking every match with "FTA ready switch is not active".
		ftaReady: true,
	},

	registers: map[register]uint16{
		fieldIoConnection: 0,
	},

	coils: map[coil]uint16{
		heartbeat:        0,
		matchReset:       1,
		stackLightGreen:  2,
		stackLightOrange: 3,
		stackLightRed:    4,
		stackLightBlue:   5,
		stackLightBuzzer: 6,
		fieldResetLight:  7,
		// awardsModeLight and any other upstream/season coil are intentionally absent: upstream/season I/O never
		// gets a wire address in this map (see docs/PlcWireMap.md). They are simply not sent to the wire.
	},
}

// scatterInputs translates raw wire discrete-input bytes (as returned by a Modbus FC2 read, sized to
// m.inputCount bits) into upstream's logical input array, via the wire map.
func (m *wireMap) scatterInputs(wireBytes []byte, logicalInputs *[inputCount]bool) {
	wireBits := byteToBool(wireBytes, int(m.inputCount))
	for i := input(0); i < inputCount; i++ {
		if address, ok := m.inputs[i]; ok {
			logicalInputs[i] = wireBits[address]
		} else {
			// Explicitly declared default (see inputDefaults above); never an implicit zero value for a signal
			// this map is responsible for.
			logicalInputs[i] = m.inputDefaults[i]
		}
	}
}

// scatterRegisters translates raw wire holding-register words (sized to m.registerCount) into upstream's logical
// register array, via the wire map. A register absent from the map stays at its zero value.
func (m *wireMap) scatterRegisters(wireWords []uint16, logicalRegisters *[registerCount]uint16) {
	for i := register(0); i < registerCount; i++ {
		if address, ok := m.registers[i]; ok {
			logicalRegisters[i] = wireWords[address]
		}
	}
}

// gatherCoils translates upstream's logical coil array into a wire-sized (m.coilCount) slice of coil states, via
// the wire map. A logical coil absent from the map is simply not represented on the wire.
func (m *wireMap) gatherCoils(logicalCoils [coilCount]bool) []bool {
	wireBits := make([]bool, m.coilCount)
	for i := coil(0); i < coilCount; i++ {
		if address, ok := m.coils[i]; ok {
			wireBits[address] = logicalCoils[i]
		}
	}
	return wireBits
}
