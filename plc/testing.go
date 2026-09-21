// Copyright 2026 Team 766. All Rights Reserved.
//
// A test-only entry point, exported because ModbusPlc's client/handler fields are private: without it, a test in
// another package (e.g. field's Arena tests, which need to exercise the M-Ayhem wire map end-to-end through a
// real ModbusPlc) would have no way to attach a fake Modbus client and drive a poll cycle. Nothing outside tests
// calls this.

package plc

import (
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
)

// PollWithClientForTesting attaches the given Modbus client to the PLC if it is not already connected, then
// performs one poll cycle (read inputs/registers, write coils) against it -- the same work Run() does against a
// real network connection, minus the network. Intended for tests only.
func (plc *ModbusPlc) PollWithClientForTesting(client modbus.Client) {
	plc.client = client
	if plc.handler == nil {
		plc.handler = modbus.NewTCPClientHandler("dummy")
	}
	if plc.ioChangeNotifier == nil {
		plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)
	}
	plc.update()
}
