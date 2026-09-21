# M-Ayhem PLC wire map

The frozen Modbus wire map for Team 766's Arduino field PLC (`fakeplc-arduino`, branch
`plc-cheesy-arena-compat`). Implemented in `plc/mayhem_wire_map.go`, applied only at the wire boundary
(`readInputs`/`readRegisters`/`writeCoils` in `plc/plc.go`), selected by `EventSettings.PlcWireMap` ("mayhem",
the default, or "upstream"). Derivation: [agents/reference/plc.md](agents/reference/plc.md) section 3.2.

## Discrete inputs

| Wire address | Signal | Provided by the Arduino | Default if not |
|---|---|---|---|
| 0 | `fieldEStop` | yes | |
| 1 | `red1EStop` | yes | |
| 2 | `red1AStop` | yes | |
| 3 | `red2EStop` | yes | |
| 4 | `red2AStop` | yes | |
| 5-6 | `red3EStop`, `red3AStop` | no | healthy (not stopped) |
| 7 | `blue1EStop` | yes | |
| 8 | `blue1AStop` | yes | |
| 9 | `blue2EStop` | yes | |
| 10 | `blue2AStop` | yes | |
| 11-12 | `blue3EStop`, `blue3AStop` | no | healthy (not stopped) |
| 13-18 | `redConnected1..3`, `blueConnected1..3` | yes (display only) | |
| 19 | `ftaReady` | no | healthy (ready) |
| 20-31 | reserved | -- | -- |

## Holding registers

| Wire address | Signal | Provided by the Arduino | Default if not |
|---|---|---|---|
| 0 | `fieldIoConnection` | yes (constant `0xFFFF`) | |
| 1-15 | reserved | -- | -- |

## Coils

| Wire address | Signal | Provided by the Arduino | Default if not |
|---|---|---|---|
| 0 | `heartbeat` | accepted, ignored (no watchdog) | |
| 1 | `matchReset` | accepted, ignored | |
| 2 | `stackLightGreen` | yes | |
| 3 | `stackLightOrange` | accepted, ignored | |
| 4 | `stackLightRed` | yes | |
| 5 | `stackLightBlue` | yes | |
| 6 | `stackLightBuzzer` | accepted, ignored | |
| 7 | `fieldResetLight` | accepted, ignored | |
| 8-15 | reserved (15 = future `stationLightsActive`) | -- | -- |
| 16-31 | reserved (future per-station lights need 18 coils, so that feature must raise the coil table size here and in the firmware together) | -- | -- |
| 34-63 | reserved | -- | -- |

## Rules

- Upstream/season I/O (game registers, hub sensors, truss/hub coils, `awardsModeLight`, etc.) never gets a wire
  address here. Absent from the tables above means never sent to or read from the wire.
- To add a signal: append it at the next free reserved address, in the table above and in
  `plc/mayhem_wire_map.go`. Never renumber or reuse an address; the firmware's addressing is fixed.
- Every `*EStop`/`*AStop` input, plus `fieldEStop`, must have a wire address or an explicit default;
  `plc/mayhem_wire_map_test.go` fails if a future upstream addition has neither.
- Arduino-side firmware changes this map assumes are in [agents/reference/plc.md](agents/reference/plc.md)
  section 3.1.

## Table sizes

The FMS reads 32 discrete inputs and 8 holding registers and writes 32 coils, which is exactly what the `fakeplc-mega` firmware serves. A Modbus request that runs past the end of a table is rejected as a whole and the PLC shows as unhealthy, so these sizes change only together with the firmware.
