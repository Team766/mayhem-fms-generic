package plc

import (
	"github.com/Team254/cheesy-arena/websocket"
)
type ArduinoPlc struct {
	address          string
	isHealthy        bool
	ioChangeNotifier *websocket.Notifier
}

func NewArduinoPlc() *ArduinoPlc {
	return &ArduinoPlc{
		ioChangeNotifier: websocket.NewNotifier("arduinoPlcIoChange", nil),
	}
}

func (plc *ArduinoPlc) SetAddress(address string) {
	plc.address = address
}

func (plc *ArduinoPlc) IsEnabled() bool {
	return plc.address != ""
}

func (plc *ArduinoPlc) IsHealthy() bool {
	return plc.isHealthy
}

func (plc *ArduinoPlc) IoChangeNotifier() *websocket.Notifier {
	return plc.ioChangeNotifier
}

func (plc *ArduinoPlc) Run() {
	// TODO: Implement Arduino PLC communication loop
}

func (plc *ArduinoPlc) GetArmorBlockStatuses() map[string]bool {
	// TODO: Implement armor block status check for Arduino PLC
	return make(map[string]bool)
}

func (plc *ArduinoPlc) GetFieldEStop() bool {
	// TODO: Implement field E-stop check for Arduino PLC
	return false
}

func (plc *ArduinoPlc) GetTeamEStops() ([3]bool, [3]bool) {
	// TODO: Implement team E-stop checks for Arduino PLC
	return [3]bool{}, [3]bool{}
}

func (plc *ArduinoPlc) GetTeamAStops() ([3]bool, [3]bool) {
	// TODO: Implement team A-stop checks for Arduino PLC
	return [3]bool{}, [3]bool{}
}

func (plc *ArduinoPlc) GetEthernetConnected() ([3]bool, [3]bool) {
	// TODO: Implement Ethernet connection checks for Arduino PLC
	return [3]bool{}, [3]bool{}
}

func (plc *ArduinoPlc) ResetMatch() {
	// TODO: Implement match reset for Arduino PLC
}

func (plc *ArduinoPlc) SetStackLights(red, blue, orange, green bool) {
	// TODO: Implement stack light control for Arduino PLC
}

func (plc *ArduinoPlc) SetStackBuzzer(state bool) {
	// TODO: Implement stack buzzer control for Arduino PLC
}

func (plc *ArduinoPlc) SetFieldResetLight(state bool) {
	// TODO: Implement field reset light control for Arduino PLC
}

func (plc *ArduinoPlc) GetCycleState(max, index, duration int) bool {
	// TODO: Implement cycle state for Arduino PLC
	return false
}

func (plc *ArduinoPlc) GetInputNames() []string {
	// TODO: Return input names for Arduino PLC
	return []string{}
}

func (plc *ArduinoPlc) GetRegisterNames() []string {
	// TODO: Return register names for Arduino PLC
	return []string{}
}

func (plc *ArduinoPlc) GetCoilNames() []string {
	// TODO: Return coil names for Arduino PLC
	return []string{}
}

func (plc *ArduinoPlc) SetTrussLights(redLights, blueLights [3]bool) {
	// TODO: Implement truss light control for Arduino PLC
}
