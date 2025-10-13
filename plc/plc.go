// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with the field PLC.
package plc

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
)

type Plc interface {
	SetAddress(address string)
	IsEnabled() bool
	IsHealthy() bool
	IoChangeNotifier() *websocket.Notifier
	Run()
	GetArmorBlockStatuses() map[string]bool
	GetFieldEStop() bool
	GetTeamEStops() ([3]bool, [3]bool)
	GetTeamAStops() ([3]bool, [3]bool)
	GetEthernetConnected() ([3]bool, [3]bool)
	ResetMatch()
	SetStackLights(red, blue, orange, green bool)
	SetStackBuzzer(state bool)
	SetFieldResetLight(state bool)
	GetCycleState(max, index, duration int) bool
	GetInputNames() []string
	GetRegisterNames() []string
	GetCoilNames() []string
	SetTrussLights(redLights, blueLights [3]bool)
}

// InputMap defines the mapping from logical input to physical pin number
type InputMap map[input]int

// CoilMap defines the mapping from logical coil to physical pin number
type CoilMap map[coil]int

type ModbusPlc struct {
	address          string
	handler          *modbus.TCPClientHandler
	client           modbus.Client
	isHealthy        bool
	ioChangeNotifier *websocket.Notifier
	hasValidMappings bool
	inputs           [inputCount]bool
	registers        [registerCount]uint16
	coils            [coilCount]bool
	oldInputs        [inputCount]bool
	oldRegisters     [registerCount]uint16
	oldCoils         [coilCount]bool
	cycleCounter     int
	matchResetCycles int
	inputMap         InputMap
	coilMap          CoilMap
}

const (
	modbusPort         = 502
	plcLoopPeriodMs    = 100
	plcRetryIntevalSec = 3
	cycleCounterMax    = 100
)

// Discrete inputs
//
//go:generate stringer -type=input
type input int

const (
	fieldEStop input = iota
	red1EStop
	red1AStop
	red2EStop
	red2AStop
	red3EStop
	red3AStop
	blue1EStop
	blue1AStop
	blue2EStop
	blue2AStop
	blue3EStop
	blue3AStop
	redConnected1
	redConnected2
	redConnected3
	blueConnected1
	blueConnected2
	blueConnected3
	inputCount
)

// 16-bit registers
//
//go:generate stringer -type=register
type register int

const (
	fieldIoConnection register = iota
	redProcessor
	blueProcessor
	registerCount
)

// Coils
//
//go:generate stringer -type=coil
type coil int

const (
	heartbeat coil = iota
	matchReset
	stackLightGreen
	stackLightOrange
	stackLightRed
	stackLightBlue
	stackLightBuzzer
	fieldResetLight
	redTrussLightOuter
	redTrussLightMiddle
	redTrussLightInner
	blueTrussLightOuter
	blueTrussLightMiddle
	blueTrussLightInner
	coilCount
)

// Bitmask for decoding fieldIoConnection into individual ArmorBlock connection statuses.
//
//go:generate stringer -type=armorBlock
type armorBlock int

const (
	redDs armorBlock = iota
	blueDs
	redIoLink
	blueIoLink
	armorBlockCount
)

// NewModbusPlc creates a new ModbusPlc with default 1:1 pin mappings.
// For custom pin mappings, use NewModbusPlcWithMaps.
func NewModbusPlc() *ModbusPlc {
	return NewModbusPlcWithMaps(nil, nil)
}

// NewModbusPlcWithMaps creates a new ModbusPlc with the given input and coil mappings.
// If nil is passed for either map, default 1:1 mappings will be used.
// Invalid mappings will be logged and will cause IsHealthy() to return false.
func NewModbusPlcWithMaps(inputMap InputMap, coilMap CoilMap) *ModbusPlc {
	// Create default 1:1 mappings if none provided
	if inputMap == nil {
		inputMap = make(InputMap, inputCount)
		for i := 0; i < int(inputCount); i++ {
			inputMap[input(i)] = i
		}
	}

	if coilMap == nil {
		coilMap = make(CoilMap, coilCount)
		for i := 0; i < int(coilCount); i++ {
			coilMap[coil(i)] = i
		}
	}

	// Initialize the PLC with zero values for all arrays
	plc := &ModbusPlc{
		inputMap:         inputMap,
		coilMap:          coilMap,
		inputs:           [inputCount]bool{},
		oldInputs:        [inputCount]bool{},
		coils:            [coilCount]bool{},
		oldCoils:         [coilCount]bool{},
		hasValidMappings: len(inputMap) == int(inputCount) && len(coilMap) == int(coilCount),
	}

	// Create the notifier with the generateIoChangeMessage method as the message producer
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)

	if !plc.hasValidMappings {
		log.Printf("Warning: Invalid PLC pin mappings - input count: %d (expected %d), coil count: %d (expected %d)",
			len(inputMap), inputCount, len(coilMap), coilCount)
	}

	return plc
}

// getInputPin returns the physical pin number for a logical input.
// Panics if the input is not found in the mapping.
func (plc *ModbusPlc) getInputPin(in input) int {
	pin, ok := plc.inputMap[in]
	if !ok {
		panic(fmt.Sprintf("No mapping found for input %v", in))
	}
	if pin < 0 || pin >= len(plc.inputs) {
		panic(fmt.Sprintf("Invalid pin number %d for input %v (must be 0-%d)", pin, in, len(plc.inputs)-1))
	}
	return pin
}

// getCoilPin returns the physical pin number for a logical coil.
// Panics if the coil is not found in the mapping.
func (plc *ModbusPlc) getCoilPin(c coil) int {
	pin, ok := plc.coilMap[c]
	if !ok {
		panic(fmt.Sprintf("No mapping found for coil %v", c))
	}
	if pin < 0 || pin >= len(plc.coils) {
		panic(fmt.Sprintf("Invalid pin number %d for coil %v (must be 0-%d)", pin, c, len(plc.coils)-1))
	}
	return pin
}

func (plc *ModbusPlc) SetAddress(address string) {
	plc.address = address
	plc.resetConnection()

	if plc.ioChangeNotifier == nil {
		// Register a notifier that listeners can subscribe to to get websocket updates about I/O value changes.
		plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)
	}
}

// Returns true if the PLC is enabled in the configurations.
func (plc *ModbusPlc) IsEnabled() bool {
	return plc.address != ""
}

// Returns true if the PLC is connected, responding to requests, and has valid pin mappings.
func (plc *ModbusPlc) IsHealthy() bool {
	return plc.isHealthy && plc.hasValidMappings
}

// Returns a notifier which fires whenever the I/O values change.
func (plc *ModbusPlc) IoChangeNotifier() *websocket.Notifier {
	return plc.ioChangeNotifier
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *ModbusPlc) Run() {
	for {
		if plc.handler == nil {
			if !plc.IsEnabled() {
				// No PLC is configured; just allow the loop to continue to simulate inputs and outputs.
				plc.isHealthy = false
			} else {
				err := plc.connect()
				if err != nil {
					log.Printf("PLC error: %v", err)
					time.Sleep(time.Second * plcRetryIntevalSec)
					plc.isHealthy = false
					continue
				}
			}
		}

		startTime := time.Now()
		plc.update()
		time.Sleep(time.Until(startTime.Add(time.Millisecond * plcLoopPeriodMs)))
	}
}

// Returns a map of ArmorBlocks I/O module names to whether they are connected properly.
func (plc *ModbusPlc) GetArmorBlockStatuses() map[string]bool {
	statuses := make(map[string]bool, armorBlockCount)
	for i := 0; i < int(armorBlockCount); i++ {
		statuses[strings.Title(armorBlock(i).String())] = plc.registers[fieldIoConnection]&(1<<i) > 0
	}
	return statuses
}

// Returns the state of the field emergency stop button (true if e-stop is active).
func (plc *ModbusPlc) GetFieldEStop() bool {
	return !plc.inputs[plc.getInputPin(fieldEStop)]
}

// Returns the state of the red and blue driver station emergency stop buttons (true if E-stop is active).
func (plc *ModbusPlc) GetTeamEStops() ([3]bool, [3]bool) {
	var redEStops, blueEStops [3]bool
	redEStops[0] = !plc.inputs[plc.getInputPin(red1EStop)]
	redEStops[1] = !plc.inputs[plc.getInputPin(red2EStop)]
	redEStops[2] = !plc.inputs[plc.getInputPin(red3EStop)]
	blueEStops[0] = !plc.inputs[plc.getInputPin(blue1EStop)]
	blueEStops[1] = !plc.inputs[plc.getInputPin(blue2EStop)]
	blueEStops[2] = !plc.inputs[plc.getInputPin(blue3EStop)]
	return redEStops, blueEStops
}

// Returns the state of the red and blue driver station autonomous stop buttons (true if A-stop is active).
func (plc *ModbusPlc) GetTeamAStops() ([3]bool, [3]bool) {
	var redAStops, blueAStops [3]bool
	redAStops[0] = !plc.inputs[plc.getInputPin(red1AStop)]
	redAStops[1] = !plc.inputs[plc.getInputPin(red2AStop)]
	redAStops[2] = !plc.inputs[plc.getInputPin(red3AStop)]
	blueAStops[0] = !plc.inputs[plc.getInputPin(blue1AStop)]
	blueAStops[1] = !plc.inputs[plc.getInputPin(blue2AStop)]
	blueAStops[2] = !plc.inputs[plc.getInputPin(blue3AStop)]
	return redAStops, blueAStops
}

// Returns whether anything is connected to each station's designated Ethernet port on the SCC.
func (plc *ModbusPlc) GetEthernetConnected() ([3]bool, [3]bool) {
	return [3]bool{
			plc.inputs[plc.getInputPin(redConnected1)],
			plc.inputs[plc.getInputPin(redConnected2)],
			plc.inputs[plc.getInputPin(redConnected3)],
		},
		[3]bool{
			plc.inputs[plc.getInputPin(blueConnected1)],
			plc.inputs[plc.getInputPin(blueConnected2)],
			plc.inputs[plc.getInputPin(blueConnected3)],
		}
}

// Resets the internal state of the PLC to start a new match.
func (plc *ModbusPlc) ResetMatch() {
	plc.coils[plc.getCoilPin(matchReset)] = true
	plc.matchResetCycles = 0

	// Clear register variables (other than fieldIoConnection) so that any values from pre-match testing don't carry
	// over.
	for i := 1; i < int(registerCount); i++ {
		plc.registers[i] = 0
	}
}

// Sets the on/off state of the stack lights on the scoring table.
func (plc *ModbusPlc) SetStackLights(red, blue, orange, green bool) {
	plc.coils[plc.getCoilPin(stackLightRed)] = red
	plc.coils[plc.getCoilPin(stackLightBlue)] = blue
	plc.coils[plc.getCoilPin(stackLightOrange)] = orange
	plc.coils[plc.getCoilPin(stackLightGreen)] = green
}

// Triggers the "match ready" chime if the state is true.
func (plc *ModbusPlc) SetStackBuzzer(state bool) {
	plc.coils[plc.getCoilPin(stackLightBuzzer)] = state
}

// Sets the on/off state of the field reset light.
func (plc *ModbusPlc) SetFieldResetLight(state bool) {
	plc.coils[plc.getCoilPin(fieldResetLight)] = state
}

func (plc *ModbusPlc) GetCycleState(max, index, duration int) bool {
	return plc.cycleCounter/duration%max == index
}

func (plc *ModbusPlc) GetInputNames() []string {
	inputNames := make([]string, inputCount)
	for i := range plc.inputs {
		inputNames[i] = input(i).String()
	}
	return inputNames
}

func (plc *ModbusPlc) GetRegisterNames() []string {
	registerNames := make([]string, registerCount)
	for i := range plc.registers {
		registerNames[i] = register(i).String()
	}
	return registerNames
}

func (plc *ModbusPlc) GetCoilNames() []string {
	coilNames := make([]string, coilCount)
	for i := range plc.coils {
		coilNames[i] = coil(i).String()
	}
	return coilNames
}

// Sets the state of the red and blue truss lights. Each array represents the outer, middle, and inner lights,
// respectively.
func (plc *ModbusPlc) SetTrussLights(redLights, blueLights [3]bool) {
	plc.coils[plc.getCoilPin(redTrussLightOuter)] = redLights[0]
	plc.coils[plc.getCoilPin(redTrussLightMiddle)] = redLights[1]
	plc.coils[plc.getCoilPin(redTrussLightInner)] = redLights[2]
	plc.coils[plc.getCoilPin(blueTrussLightOuter)] = blueLights[0]
	plc.coils[plc.getCoilPin(blueTrussLightMiddle)] = blueLights[1]
	plc.coils[plc.getCoilPin(blueTrussLightInner)] = blueLights[2]
}
func (plc *ModbusPlc) connect() error {
	address := fmt.Sprintf("%s:%d", plc.address, modbusPort)
	handler := modbus.NewTCPClientHandler(address)
	handler.Timeout = 1 * time.Second
	handler.SlaveId = 0xFF
	err := handler.Connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to PLC at %s", address)

	plc.handler = handler
	plc.client = modbus.NewClient(plc.handler)
	plc.writeCoils() // Force initial write of the coils upon connection since they may not be triggered by a change.
	return nil
}

func (plc *ModbusPlc) resetConnection() {
	if plc.handler != nil {
		plc.handler.Close()
		plc.handler = nil
	}
}

// Performs a single iteration of reading inputs from and writing outputs to the PLC.
func (plc *ModbusPlc) update() {
	if plc.handler != nil {
		isHealthy := true
		isHealthy = isHealthy && plc.writeCoils()
		isHealthy = isHealthy && plc.readInputs()
		isHealthy = isHealthy && plc.readRegisters()
		if !isHealthy {
			plc.resetConnection()
		}
		plc.isHealthy = isHealthy
	}

	plc.cycleCounter++
	if plc.cycleCounter == cycleCounterMax {
		plc.cycleCounter = 0
	}

	// Detect any changes in input or output and notify listeners if so.
	if plc.inputs != plc.oldInputs || plc.registers != plc.oldRegisters || plc.coils != plc.oldCoils {
		plc.ioChangeNotifier.Notify()
		plc.oldInputs = plc.inputs
		plc.oldRegisters = plc.registers
		plc.oldCoils = plc.coils
	}
}

func (plc *ModbusPlc) readInputs() bool {
	if len(plc.inputs) == 0 {
		return true
	}
	inputs, err := plc.client.ReadDiscreteInputs(0, uint16(len(plc.inputs)))
	if err != nil {
		log.Printf("PLC error reading inputs: %v", err)
		return false
	}
	if len(inputs)*8 < len(plc.inputs) {
		log.Printf("Insufficient length of PLC inputs: got %d bytes, expected %d bits.", len(inputs), len(plc.inputs))
		return false
	}
	copy(plc.inputs[:], byteToBool(inputs, len(plc.inputs)))
	return true
}

func (plc *ModbusPlc) readRegisters() bool {
	if len(plc.registers) == 0 {
		return true
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(len(plc.registers)))
	if err != nil {
		log.Printf("PLC error reading registers: %v", err)
		return false
	}
	if len(registers)/2 < len(plc.registers) {
		log.Printf(
			"Insufficient length of PLC registers: got %d bytes, expected %d words.",
			len(registers),
			len(plc.registers),
		)
		return false
	}

	copy(plc.registers[:], byteToUint(registers, len(plc.registers)))
	return true
}

func (plc *ModbusPlc) writeCoils() bool {
	// Send a heartbeat to the PLC so that it can disable outputs if the connection is lost.
	plc.coils[plc.getCoilPin(heartbeat)] = true

	coils := boolToByte(plc.coils[:])
	_, err := plc.client.WriteMultipleCoils(0, uint16(len(plc.coils)), coils)
	if err != nil {
		log.Printf("PLC error writing coils: %v", err)
		return false
	}

	if plc.matchResetCycles > 5 {
		// Only need a short pulse to reset the internal state of the PLC.
		plc.coils[plc.getCoilPin(matchReset)] = false
	} else {
		plc.matchResetCycles++
	}

	return true
}

func (plc *ModbusPlc) generateIoChangeMessage() any {
	return &struct {
		Inputs    []bool
		Registers []uint16
		Coils     []bool
	}{plc.inputs[:], plc.registers[:], plc.coils[:]}
}

func byteToBool(bytes []byte, size int) []bool {
	bools := make([]bool, size)
	for i := 0; i < size; i++ {
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		bitMask := byte(1 << bitIndex)
		bools[i] = bytes[byteIndex]&bitMask != 0
	}
	return bools
}

func byteToUint(bytes []byte, size int) []uint16 {
	uints := make([]uint16, size)
	for i := 0; i < size; i++ {
		uints[i] = uint16(bytes[2*i])<<8 + uint16(bytes[2*i+1])
	}
	return uints
}

func boolToByte(bools []bool) []byte {
	bytes := make([]byte, (len(bools)+7)/8)
	for i, bit := range bools {
		if bit {
			bytes[i/8] |= 1 << uint(i%8)
		}
	}
	return bytes
}
