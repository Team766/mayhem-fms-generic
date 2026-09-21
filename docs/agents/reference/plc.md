# M-Ayhem Arduino PLC: bench test and per-team lights

The field's PLC is Team 766's Arduino Mega running `fakeplc-arduino/fakeplc-mega` (branch `plc-cheesy-arena-compat`, as
flashed for M-Ayhem 2025). It is a Modbus TCP server at 10.0.100.15:502 with fixed addresses that match upstream's
generic signal order. The FMS's PLC code is upstream's; see `docs/DEVELOPMENT.md`, "PLC", for the contract and
`plc/mayhem_arduino_test.go` for the address guard. Written September 2026; trust the firmware and the code over this page.

## Bench test

Run this after any upstream sync that touched `plc/` or the arena's PLC handling, and before the event.

Setup: laptop at 10.0.100.5/24 on the same switch as the Arduino; build and run `./cheesy-arena`; Settings > PLC Address =
`10.0.100.15`; turn on 2v2 mode (the hardware has no station-3 stop wiring); open `/setup/field_testing`; Arduino serial
monitor at 115200.

1. **Link and health.** The serial monitor shows one "new client" and does not repeat it (repeating means the connection is flapping: a Modbus request ran past the end of a table the firmware serves). The FMS log has no "PLC error". Match Play shows the PLC healthy and the ArmorBlocks connected.
2. **Inputs, one at a time**, watching the Inputs column: field e-stop; each of R1, R2, B1, B2 driver-station e-stop; each field-side per-team e-stop (it must flip the same team's bit); each a-stop. Pressed reads `false` and the station shows stopped on Match Play. Unplug an e-stop cable: it must read as stopped.
3. **E-stop latch.** In a test match, press and release a team e-stop: the station stays stopped until the match ends. An a-stop only holds through auto.
4. **Field e-stop** during a match aborts the match.
5. **Coils by override** (pre-match only): cycle stack green, red, blue and the field reset light from the field-testing page; check the right light responds and that "auto" returns control. Overrides clear when a match is loaded.
6. **Pre-match cycle in 2v2.** Load a match with nothing connected, then bypass or connect the four stations one by one. With all four ready the stack light goes green and the buzzer sounds. (In 2025 this never happened with the PLC enabled; 2v2 mode now ignores R3 and B3.)
7. **Recovery.** Pull the Arduino's Ethernet for 10 s and replug: the FMS returns to healthy within a few seconds without resetting the Arduino. Kill and restart the FMS: the Arduino accepts the new session.
8. **Soak** for 30 minutes: no "PLC error" lines and no repeated "new client" on the serial monitor. Make sure no second Modbus client (another FMS instance, a test script) is attached; the firmware serves one client.

## Per-team station lights (planned, not built)

Goal: a light at each driver station showing that team's state, as a cheap replacement for team signs.

- **Hardware today:** one RGB light at each of R1, R2, B1, B2, currently driven from the alliance-level stack-light coils. No lights at station 3.
- **What upstream already computes per station:** e-stop, a-stop, bypass, no team, driver-station and robot link, and readiness for a single station (`checkAllianceStationsReady` accepts one station). Upstream's team-sign code has a ready-made state ladder to copy: e-stop amber; a-stop in auto blinking amber; pre-match ready; field reset green; link lost in match blinking.
- **Design:** three coils (R, G, B) per station, appended after the generic coils, plus one "station lights active" coil so the firmware falls back to today's behaviour with an FMS that does not send them. Blinking is done in the FMS with the existing cycle helper. One new setter on the PLC interface and about 40 lines in the arena's PLC loop. The field-testing page lists new coils automatically from the generated names, and the generated compile guard forces `go generate`.
- **This needs a firmware change and new fixed addresses**, so extend `plc/mayhem_arduino_test.go` with them, and size the coil table in the firmware to match (it serves 32 coils today; 8 generic + 19 new fits).
- **Decide first:** the colour legend (red is also an alliance colour; an on/off RGB light cannot show orange distinct from yellow), light polarity, and whether station 3 gets lights.
- **Rough effort:** a day of Go including tests, half a day of firmware, and a 2-3 hour bench session. The lights are advisory; e-stops are enforced by the FMS and the driver stations, not by the Arduino.
