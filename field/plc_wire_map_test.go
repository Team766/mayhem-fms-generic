// Copyright 2026 Team 766. All Rights Reserved.
//
// An arena-level integration test for the M-Ayhem wire map (plc/mayhem_wire_map.go): drives a real
// *plc.ModbusPlc, configured with the Mayhem wire map, through a fake Modbus client and checks that a wired team
// E-Stop still reaches AllianceStation.EStop and blocks match start, exactly as it would with upstream's own wire
// numbering. The wire-level details (frozen indices, unwired-station defaults, coil override, etc.) are covered
// by plc/mayhem_wire_map_test.go; this test exists to prove nothing is lost or miswired crossing the plc/field
// package seam.

package field

import (
	"errors"
	"github.com/Team254/cheesy-arena/plc"
	"github.com/stretchr/testify/assert"
	"testing"
)

// fakeWireModbusClient is a minimal modbus.Client, sized to the Mayhem wire map's fixed tables (32 inputs, 16
// registers), used only to drive a real ModbusPlc through Arena without a live PLC connection.
type fakeWireModbusClient struct {
	inputs    [32]bool
	registers [16]uint16
}

func (c *fakeWireModbusClient) ReadCoils(address, quantity uint16) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	if address != 0 {
		return nil, errors.New("unexpected address")
	}
	bytes := make([]byte, (len(c.inputs)+7)/8)
	for i, bit := range c.inputs {
		if bit {
			bytes[i/8] |= 1 << uint(i%8)
		}
	}
	return bytes, nil
}

func (c *fakeWireModbusClient) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) ReadInputRegisters(address, quantity uint16) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) ReadHoldingRegisters(address, quantity uint16) (results []byte, err error) {
	if address != 0 {
		return nil, errors.New("unexpected address")
	}
	bytes := make([]byte, len(c.registers)*2)
	for i, value := range c.registers {
		bytes[2*i] = byte(value >> 8)
		bytes[2*i+1] = byte(value)
	}
	return bytes, nil
}

func (c *fakeWireModbusClient) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) WriteMultipleRegisters(
	address, quantity uint16, value []byte,
) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) ReadWriteMultipleRegisters(
	readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte,
) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) MaskWriteRegister(address, andMask, orMask uint16) (results []byte, err error) {
	return nil, nil
}

func (c *fakeWireModbusClient) ReadFIFOQueue(address uint16) (results []byte, err error) {
	return nil, nil
}

// With the Mayhem wire map selected and 2v2 mode off (a plain 3v3 event), pressing a WIRED team E-Stop (as
// opposed to an unwired station-3 one, which reads healthy by design) still stops that alliance station and
// blocks the match from starting.
func TestPlcMayhemWireMapWiredEStopBlocksMatchStart(t *testing.T) {
	arena := setupTestArena(t)
	assert.False(t, arena.EventSettings.TwoVsTwoMode) // 3v3: the 2v2 setting is off.
	assert.Equal(t, "mayhem", arena.EventSettings.PlcWireMap)

	modbusPlc, ok := arena.Plc.(*plc.ModbusPlc)
	if !assert.True(t, ok, "Arena.Plc should be a *plc.ModbusPlc by default") {
		return
	}
	modbusPlc.SetAddress("10.0.100.15") // Enables the PLC; no network call is actually made in this test.

	client := &fakeWireModbusClient{}
	// Every wired input reads healthy except red1EStop (wire address 1), which is pressed.
	client.inputs[0] = true  // fieldEStop
	client.inputs[1] = false // red1EStop: pressed.
	client.inputs[2] = true  // red1AStop
	client.inputs[3] = true  // red2EStop
	client.inputs[4] = true  // red2AStop
	client.inputs[7] = true  // blue1EStop
	client.inputs[8] = true  // blue1AStop
	client.inputs[9] = true  // blue2EStop
	client.inputs[10] = true // blue2AStop
	client.registers[0] = 0xFFFF
	modbusPlc.PollWithClientForTesting(client)

	// Bypass everything so the only thing that can block match start is the E-Stop.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true

	arena.Update()
	assert.True(t, arena.AllianceStations["R1"].EStop)
	assert.False(t, arena.AllianceStations["R2"].EStop)
	assert.False(t, arena.AllianceStations["R3"].EStop) // Unwired station 3 reads healthy, not stopped.

	err := arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "an emergency stop is active (R1)")
	}

	// Release the E-Stop and confirm the match can now start.
	client.inputs[1] = true
	modbusPlc.PollWithClientForTesting(client)
	arena.Update()
	assert.False(t, arena.AllianceStations["R1"].EStop)
	assert.Nil(t, arena.StartMatch())
}
