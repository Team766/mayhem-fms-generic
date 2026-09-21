# Mayhem PLC inventory (for regenerating the generic base)

Read-only research, 2026-09-20. Refs inspected: `mayhem-fms-generic` (`upstream/main` = f2f38f4, `fms2025/main` = 3bc6890,
`origin/main` = 83b987f, fork point aa1cd2b 2025-05-27), `mayhem-fms-2025` (main = 3bc6890, identical to `fms2025/main`),
`mayhem-fms-2026` (HEAD 799f26d, upstream base b573da1), `fakeplc-arduino` (checked-out branch `plc-cheesy-arena-compat`
= 43e1d06, 2025-11-07). Nothing was bench-tested; statements about Modbus exception behaviour are from reading
ArduinoModbus/libmodbus semantics and are marked where unsure.

---

## 1. What the Mayhem PLC is

### 1.1 Firmware summary (`../fakeplc-arduino/fakeplc-mega/fakeplc-mega.ino`, 255 lines)

- Arduino Mega + Ethernet shield (CS pin 10), static IP **10.0.100.15**/24, gw 10.0.100.1, MAC DE:AD:BE:EF:FE:ED,
  Modbus TCP server on **port 502** using `ArduinoModbus` (`ModbusTCPServer`, libmodbus underneath).
- Tables configured in `setup()`: **8 holding registers** (addr 0-7), **32 coils** (0-31), **32 discrete inputs** (0-31).
  No input registers.
- Holding register 0 is written once to `0xFFFF` ("all ArmorBlocks connected") and never changed. Registers 1-7 stay 0.
- Serves **one TCP client at a time**: `loop()` takes `server.available()` and then spins in
  `while (client.connected()) { poll(); processStackLightOutputs(); readInputs(); }`.
- **Inputs** (`readInputs()`), all pins `INPUT_PULLUP`, value = `digitalRead(pin) ^ gpio_input_invert[i]`:
  team E-stop bit = `(driver-station E-stop pin OR field-side per-team E-stop pin) ^ 1`. Field-side per-team E-stop
  pins are 41/43/45/47. Inputs whose pin entry is `200` are skipped, so they are **never written and stay 0**.
- **Outputs**: the generic coil->pin copy (`updateOutputs()`) is **commented out**. The only output logic is
  `processStackLightOutputs()`, which reads three coils (2 = green, 4 = red, 5 = blue) and drives **four RGB lights**
  (one per driver station: Red1, Red2, Blue1, Blue2; pins 26/28/30, 32/34/36, 38/40/42, 44/46/48 = R/G/B each):
  - green coil set -> all four lights green
  - red && blue -> red stations red, blue stations blue
  - red only -> red stations red, blue stations green; blue only -> the mirror
  - none of the three set (e.g. PostMatch orange-only, or the "off" half of the blinking green) -> **no branch, lights hold
    the previous state** (so the FMS's blinking green shows as solid green).
  In the light helpers HIGH = lit (this contradicts the `setup()` comment "optoisolator board is active low"; I did not
  resolve which is right for the wiring, the code as written treats HIGH as on for lights).
- **Not implemented**: heartbeat (coil 0 is ignored; there is no watchdog, lights just hold last state when the FMS goes
  away), matchReset (coil 1), orange stack light (3), buzzer (6), **field reset light (7)**, any game-specific coil,
  any counting register, `ftaReady`, station 3 E/A-stops, blue "Ethernet connected".
  Output pins 22, 24, 4, 5 are configured and left HIGH forever.
- `tick()`: blinks pin 13 (500 ms idle / 100 ms with a client). With `debug_print = true` (the committed default) it
  prints ~70 chars to Serial on **every loop iteration** (the print is outside the rate limit), which slows the poll loop.

Firmware defects worth knowing (from reading, not observed):
1. `gpio_input_pins[16]` has only **15 initialisers** (pin 49 was dropped relative to `master`), so index 13 = pin 2,
   14 = pin 3, **15 = pin 0 (Serial RX)**. `redConnected1..3` therefore read pins 2, 3, 0.
2. `pinMode(200, INPUT_PULLUP)` is executed for the four placeholder entries; pin 200 does not exist on a Mega
   (out-of-range PROGMEM lookup; undefined, probably harmless, should be guarded).
3. Single-client blocking loop: if the FMS TCP session half-dies without FIN, the sketch can stay in the inner loop and
   never service a new connection until reset. Unverified, but it is the first thing to suspect if "PLC unhealthy" sticks.
4. Field E-stop (pin 23) has invert 0: an **unwired pin reads "not e-stopped"** (not fail-safe), while team E-stops are
   fail-safe (unwired = stopped).

Firmware repo branches: `master` f7b0e3e (2024 event: plain 16-in/16-out passthrough), `origin/M-Ayhem2025` 772e619
(2025-11-05, macro rewrite by students, has its own bugs), `plc-changes-for-eastops-and-lights` 42e5920,
`plc-cheesy-arena-compat` 43e1d06 (2025-11-07 23:37, commit message: "get fakeplc working with standard cheesy-arena
modbus addresses ... remap PLC stack light to team lights"). **I cannot tell from git which sketch was flashed at the
2025 event**; the compat branch is the newest and is what is checked out, and it is the only one consistent with an
unremapped FMS, so I treat it as the shipped firmware. Confirm with the team.

### 1.2 Modbus map: upstream vs Arduino

FMS side (unchanged in all shipped code): `ReadDiscreteInputs(0, inputCount)` (FC2), `ReadHoldingRegisters(0, registerCount)`
(FC3), `WriteMultipleCoils(0, coilCount)` (FC15), unit id 0xFF, 100 ms loop, 1 s timeout (`plc/plc.go` `readInputs`,
`readRegisters`, `writeCoils`). Input polarity: E-stop/A-stop inputs are **true = not stopped** (`GetTeamEStops` negates).

| Logical signal | Upstream index (2025 fork point aa1cd2b) | Upstream index (upstream/main 2026) | Arduino index / pin | Match? |
|---|---|---|---|---|
| **Discrete inputs** | | | | |
| fieldEStop | 0 | 0 | 0 / pin 23 (not inverted) | yes (unwired = OK, not fail-safe) |
| red1EStop | 1 | 1 | 1 / pin 25 OR pin 41, inverted | yes |
| red1AStop | 2 | 2 | 2 / pin 27, inverted | yes |
| red2EStop | 3 | 3 | 3 / pin 29 OR pin 43, inverted | yes |
| red2AStop | 4 | 4 | 4 / pin 31, inverted | yes |
| red3EStop, red3AStop | 5, 6 | 5, 6 | 5, 6 / none (never written = 0) | index yes, **value = permanently stopped** |
| blue1EStop | 7 | 7 | 7 / pin 33 OR pin 45, inverted | yes |
| blue1AStop | 8 | 8 | 8 / pin 35, inverted | yes |
| blue2EStop | 9 | 9 | 9 / pin 37 OR pin 47, inverted | yes |
| blue2AStop | 10 | 10 | 10 / pin 39, inverted | yes |
| blue3EStop, blue3AStop | 11, 12 | 11, 12 | 11, 12 / none (0) | index yes, **value = permanently stopped** |
| redConnected1..3 | 13-15 | 13-15 | 13 / pin 2, 14 / pin 3, 15 / pin 0 (array bug) | index yes, values meaningless (display only) |
| blueConnected1..3 | 16-18 | 16-18 | never written (0) | index yes, always false (display only) |
| ftaReady | n/a | **19** | never written (0) | **NO: always "not ready"** |
| red/blueHubSensor1-4 | n/a | 20-27 | never written (0) | n/a (game specific) |
| **Holding registers** | | | | |
| fieldIoConnection | 0 | 0 | 0 = 0xFFFF constant | yes (all 4 ArmorBlocks "connected") |
| game registers | 1-2 (processor counts) | 1-10 (hub totals/counts) | 1-7 exist (0); **8-10 do not exist** | **NO on 2026: read of 11 registers exceeds the 8 configured** |
| **Coils** | | | | |
| heartbeat | 0 | 0 | accepted, ignored | index yes, no watchdog |
| matchReset | 1 | 1 | accepted, ignored | n/a (no counters) |
| stackLightGreen | 2 | 2 | 2 -> all 4 station lights green | yes |
| stackLightOrange | 3 | 3 | ignored | not shown |
| stackLightRed | 4 | 4 | 4 -> Red1/Red2 lights red | yes |
| stackLightBlue | 5 | 5 | 5 -> Blue1/Blue2 lights blue | yes |
| stackLightBuzzer | 6 | 6 | ignored | not implemented |
| fieldResetLight | 7 | 7 | ignored (pins 22/24 unused) | not implemented |
| game coils | 8-13 truss lights | 8 awardsModeLight, 9-12 hub motors/lights | accepted (32 coils), ignored | harmless |

### 1.3 Is it wire compatible with UNMODIFIED current upstream?

- **With the 2025 upstream (fork point) it was**: 19 inputs <= 32, 3 registers <= 8, 14 coils <= 32, and the evergreen
  indices line up. That is why no Go remap was needed at the 2025 event (see section 2).
- **With unmodified `upstream/main` today it is NOT usable**, for three independent reasons:
  1. `registerCount` is now **11**; the sketch configures 8 holding registers. libmodbus answers a read of 0..10 with
     exception 02 (illegal data address), so `readRegisters()` returns false every cycle -> `resetConnection()` ->
     `IsHealthy()` false forever -> `getStartMatchConditions()` reports "PLC is not healthy" and the match cannot start
     (`field/arena.go` ~1132). Inputs/coils are still exchanged each reconnect, but the connection flaps at 10 Hz.
     (Confidence: high on the libmodbus bounds check, not bench-verified.)
  2. New evergreen input **`ftaReady` (index 19)**, added upstream in f1be400: `getStartMatchConditions()` requires
     `Plc.IsFtaReady()`; the Arduino never sets input 19 -> "FTA ready switch is not active" blocks every match.
  3. Station 3 E-stop/A-stop inputs are always 0 = stopped. Unmodified upstream calls `handleTeamStop("R3"/"B3", ...)`
     unconditionally, so R3/B3 are permanently e-stopped and `getAllianceStationStartConditions` blocks the match
     (bypass does not clear an E-stop). The 2025 fork hid this with `TwoVsTwoMode` guards (section 2).
- **What breaks when upstream's enum order changes**: the firmware hard-codes wire indices (coils 2/4/5, inputs 0-15).
  History of the enums (from `upstream/chezy_champs_20xx:plc/plc.go`):
  - coils 0-7 (heartbeat .. fieldResetLight) have been stable since at least 2022; everything from index 8 up is
    replaced every year (2022 hubMotors; 2023 charge station lights; 2024 12 speaker/amp coils; 2025 6 truss lights;
    2026 awardsModeLight + 4 hub coils).
  - inputs were **renumbered in 2024** when A-stops were interleaved (2023: `redEstop1..3` at 1-3; 2024+: E/A pairs at
    1-12), then game inputs come and go at the tail (2024: 19-22 amplify/coop; 2025: none; 2026: ftaReady at 19 + 8 hub
    sensors).
  - registers: only index 0 is evergreen; the count swings (17, 1, 6, 3, 11).
  So the three failure modes are (a) **count growth past the firmware's table sizes** (what bites now, on registers),
  (b) **new evergreen gating inputs** such as ftaReady, (c) a **renumbering of the evergreen prefix** (happened once in
  five years; would silently swap E-stops/A-stops, which is the dangerous one).

---

## 2. Which Go-side PLC customisations actually shipped in 2025

`mayhem-fms-2025` main == `fms2025/main` (3bc6890). No branch in that repo differs from main under `plc/`.
`fms2025/main` descends from generic `origin/main` 83b987f, so it contains both 0cd5d47 (remap, PR #12, 2025-10-15) **and**
its revert 3618ee4 (PR #14, 2025-10-18). Net effect: **no remapping shipped.**

Diff `aa1cd2b..fms2025/main` for PLC-related files:

| File | Shipped change |
|---|---|
| `plc/plc.go` (+3/-9) | import reordering; removed `GetProcessorCounts()` from the `Plc` interface and `ModbusPlc` (game genericising, commit 0d265a4). Enums, I/O code, truss-light coils all unchanged from upstream 2025. |
| `plc/plc_test.go` (+2/-26) | removed `TestPlcRegistersGameSpecific`. |
| `plc/*_string.go`, `plc/fake_modbus_client_test.go` | unchanged. |
| `field/fake_plc_test.go` | unchanged vs aa1cd2b (generic `origin/main` and fms2025 identical). |
| `field/arena.go` `handlePlcInputOutput` | `handleTeamStop("R3"/"B3")` wrapped in `if !arena.EventSettings.TwoVsTwoMode` (38f8664); processor-count scoring and coral co-op truss logic removed; truss lights simplified to on-during-match + warning sequence; `scoreReady` uses four near/far scoring positions. `arena.TeamSigns.Update` and `SetNextMatchTeams` calls were **removed** (team signs dropped in the fork). |
| `field/arena.go` elsewhere | `checkCanStartMatch` checks only R1,R2,B1,B2 in 2v2; `LoadMatch`/`ResetMatch` force `R3/B3.Bypass = true`, `Team = nil` in 2v2. |
| `web/setup_field_testing.go`, template, JS | unchanged vs aa1cd2b. |
| `model/event_settings.go` | `PlcAddress` only (upstream); no PLC-type setting. |

Not shipped (exists only on unmerged generic branches):
- `feature-add-arduino-plc` (571e8f3): the work squashed into 0cd5d47: `InputMap`/`CoilMap` types,
  `NewModbusPlcWithMaps`, `getInputPin`/`getCoilPin` used in **every accessor** (+157/-? lines in `plc/plc.go`), and a
  55-line `plc/mayhem_plc.go` with an identity map. Reverted on main.
- `feature-add-mayhem-plc-pins-coils` == `feature-add-mayhem-plc-team-stacklights` (same commit 8ab0014, 3 commits on top
  of 0cd5d47, **never merged**): replaces the maps with `MayhemPlc{*ModbusPlc}` that "overrides" `getInputPin`/`getCoilPin`
  and ORs four "field e-stop" inputs into `GetTeamEStops`. Problems: (1) Go embedding has no virtual dispatch, so the
  embedded `ModbusPlc` methods never call `MayhemPlc.getInputPin` - the override is dead for everything except the one
  overridden `GetTeamEStops`; (2) `fieldRed1Estop..fieldBlue2Estop = 9..12` collide with `blue2EStop..blue3AStop`;
  (3) `NewMayhemPlc()` is never instantiated (`field/arena.go:109` still `plc.NewModbusPlc()`); (4) despite the branch
  name there is **no per-team stack-light code** in it. The field-E-stop OR was subsequently moved into firmware
  (43e1d06), which made the branch moot.
- `revert-pin-remapping` (2d47862): the revert PR branch. `customizable-mayhem`, `feature-2v2`, `ui-cleanup`,
  `cleanup/merge-fms-2025-into-generic`/HEAD (a2a286c): no `plc/` differences from `origin/main`.
- `plc/mayhem_plc.go` exists only on `feature-add-arduino-plc` and the two `feature-add-mayhem-plc-*` branches. It is
  **not** on `fms2025/main`, `origin/main`, `upstream/main`, or HEAD. It never shipped.
- `mayhem-fms-2026` (feature/medieval-game and main): **zero** changes under `plc/`, `field/fake_plc_test.go` or the
  field-testing files relative to its upstream base b573da1. The `cheesy-arena` custom-games-runtime branch likewise has
  no `plc/` diff.

Was remapping reverted for good? Yes as far as the code shows: reverted on generic main three days after merging,
replaced two weeks later by the firmware-side approach ("standard cheesy-arena modbus addresses"). Nothing re-introduces it.

Latent bug in the shipped 2v2 code (code reading, not observed; please confirm from 2025 field experience): with the PLC
enabled and `TwoVsTwoMode` on, `assignTeam` sets `R3/B3.aStopReset = false` and the skipped `handleTeamStop` never sets it
back to true, yet `handlePlcInputOutput` still computes `redAllianceReady := checkAllianceStationsReady("R1","R2","R3")`.
That makes both alliances permanently "not ready" for PLC purposes: stack-light coils stay red+blue, green/buzzer never
fire, `FieldReadyAt` is never stamped, and the station lights would have stayed alliance-coloured pre-match. Match start
was unaffected because `checkCanStartMatch` used the 4-station list. Do not port the `TwoVsTwoMode` guard as-is.

---

## 3. Recommendation for the regenerated base

Principle: keep upstream's `plc/plc.go` logic and enums untouched and translate **only at the wire boundary**. The 0cd5d47
design translated inside every accessor, which (a) must be redone for every new game accessor each year, (b) limits
physical indices to `< inputCount/coilCount`, and (c) makes the field-testing page show wire values under the wrong
logical names. Avoid it.

### 3.1 Layer 0 - firmware made tolerant (needed regardless, ~25 lines)

1. Oversize the tables: `configureHoldingRegisters(0, 64)`, `configureCoils(0, 128)`, `configureDiscreteInputs(0, 128)`
   (about 400 bytes of the Mega's 8 KB). Removes failure mode (a) for good.
2. Write **1 (healthy)** for every E-stop/A-stop input that has no pin (stations 3), instead of skipping it. This removes
   the need for the `TwoVsTwoMode` guards around `handleTeamStop` and fixes the `aStopReset` bug above for free.
3. Assert `ftaReady` (wire input 19) = 1, or wire a real switch to a spare pin.
4. Fix the 15-element `gpio_input_pins` initialiser, skip `pinMode` for placeholder pins, set `debug_print = false`
   (or rate-limit it inside the `millis()` check), and add a client-idle timeout (drop the client if no Modbus request
   for ~2 s) so a half-dead session cannot wedge the sketch.
5. Optional: implement heartbeat watchdog (coil 0 not refreshed for 1 s -> lights off/amber) and drive the field reset
   light from coil 7 on pin 22.

With only Layer 0, the Arduino works against **unmodified** upstream as long as the evergreen prefix keeps its numbering.

### 3.2 Layer 1 - fixed Mayhem wire map + boundary remap in Go (small, makes the base immune to renumbering)

Document one frozen map (put it in the firmware header and `docs/`), e.g.:

- Discrete inputs: 0 fieldEStop; 1-12 R1E,R1A,R2E,R2A,R3E,R3A,B1E,B1A,B2E,B2A,B3E,B3A; 13-18 red/blueConnected1-3;
  19 ftaReady; 20-31 reserved.
- Holding registers: 0 fieldIoConnection; 1-15 reserved.
- Coils: 0 heartbeat; 1 matchReset; 2 green; 3 orange; 4 red; 5 blue; 6 buzzer; 7 fieldResetLight; 8-14 reserved;
  15 `stationLightsActive`; 16-33 per-station lights (section 4); 34-63 reserved.

This is deliberately identical to today's upstream evergreen numbering, so Layer 0 alone still interoperates.

Go changes (new file + three tiny hooks; nothing else in `plc/plc.go` moves):

- **New `plc/mayhem_wire_map.go`** (~70-90 lines, same package so it can use the unexported identifiers):
  `type wireMap struct { inputs map[input]uint16; coils map[coil]uint16; inputCount, coilCount uint16; inputDefaults map[input]bool }`
  and `var MayhemWireMap = &wireMap{...}` keyed by **identifier** (`fieldEStop: 0, red1EStop: 1, ... stackLightGreen: 2`).
  Because it is keyed by name, an upstream renumber is harmless and an upstream rename/removal is a **compile error**
  (the loud failure you want from the playbook). Game-specific inputs/coils/registers are simply absent from the map:
  unmapped coils are not sent, unmapped inputs take a default (true for `*EStop`/`*AStop`, false otherwise), registers
  other than `fieldIoConnection` stay 0.
- **Seams in `plc/plc.go`** (upstream/main line refs approximate):
  - `ModbusPlc` struct: add one field `wire *wireMap` (nil = upstream behaviour).
  - `readInputs()`: if `plc.wire != nil`, read `wire.inputCount` bits and scatter into `plc.inputs` via the map (+8 lines).
  - `readRegisters()`: if `plc.wire != nil`, read 1 register into `plc.registers[fieldIoConnection]` (+6 lines).
  - `writeCoils()`: if `plc.wire != nil`, build a `wire.coilCount` bool slice from `getEffectiveCoils()` via the map and
    write that (+8 lines). Doing it after `getEffectiveCoils()` keeps the upstream coil-override feature working.
  - Everything the UI uses (`GetInputNames/GetCoilNames`, `generateIoChangeMessage`, `SetCoilOverride`) stays logical,
    so the field-testing page keeps correct labels.
- **Selection seam**: `field/arena.go` `NewArena`: `arena.Plc = new(plc.ModbusPlc)` (upstream/main line 129) and
  `LoadSettings`: `arena.Plc.SetAddress(settings.PlcAddress)` (line 220). Smallest robust option: add
  `EventSettings.PlcWireMap string` ("" / "upstream" / "mayhem"; default "mayhem" in the fork) in
  `model/event_settings.go`, a select in `templates/setup_settings.html` next to "PLC Address", parse it in
  `web/setup_settings.go`, and in `LoadSettings` do
  `if m, ok := arena.Plc.(*plc.ModbusPlc); ok { m.SetWireMap(settings.PlcWireMap) }`. The type assertion avoids adding a
  method to the `plc.Plc` interface, so `field/fake_plc_test.go` does not change. (~20 lines + 1 settings test line.)
  If the team will never own an Allen-Bradley PLC, skip the setting and set the map unconditionally in `NewArena`
  (1 line); I would still keep the setting because it lets you bisect "is it the map or the firmware" on the bench.
- **Guard test** (`plc/mayhem_wire_map_test.go`, ~40 lines): with `MayhemWireMap` set, drive `FakeModbusClient` and assert
  each evergreen logical signal lands on the frozen wire index, that unmapped stop inputs default to not-stopped, and
  that every evergreen identifier is present in the map. Enlarge `FakeModbusClient` arrays from 32 to 64/128
  (`plc/fake_modbus_client_test.go`, 3 lines).

Do **not** carry forward: `InputMap/CoilMap/NewModbusPlcWithMaps`, `hasValidMappings`, `getInputPin/getCoilPin`,
`plc/mayhem_plc.go` (any version), the `TwoVsTwoMode` guards around `handleTeamStop`.

What else must be carried into the new base from the 2025 fork, PLC-wise: only the decision that game-specific PLC
accessors are stubbed when the game is genericised (2025 removed `GetProcessorCounts`; for 2026 the equivalents are
`GetHubCounts`, `SetHubMotors`, `SetHubLights` and the hub block at the end of `handlePlcInputOutput`). With the wire map
these can simply be left in place: they read zeros and write coils that are never sent.

---

## 4. Per-team lights feasibility

### 4.1 What already exists

- **Hardware**: the 2025 field already has one RGB light per populated station (Red1, Red2, Blue1, Blue2), 12 output
  pins. Today they only mirror alliance-level state because the FMS only sends alliance-level coils. Stations 3 have no
  light; pins 22, 24, 4, 5 are the only spare opto channels in the current 16-channel output list (enough for two
  2-colour lights, not two RGB lights). Unsure about the actual opto board channel count; confirm.
- **Per-station state upstream already computes** (all in `field/arena.go`, `AllianceStation` struct): `EStop`, `AStop`
  (latched by `handleTeamStop`), `Bypass`, `Team == nil`, `Ethernet`, `DsConn.DsLinked/RadioLinked/RobotLinked`, and
  `arena.checkAllianceStationsReady(station)` / `getAllianceStationStartConditions(station)` which **already work for a
  single station**. The exact per-station colour state machine you want already exists upstream in
  `field/team_sign.go` `generateTeamNumberTexts` (~lines 350-368): E-stop -> orange; A-stop in auto -> blinking orange;
  PreMatch and station ready -> alliance colour; FieldReset -> green; FieldVolunteers -> purple; robot link lost
  in-match -> blinking alliance colour. The fork removed the `TeamSigns.Update` call, so nothing consumes it today.

### 4.2 Design

Coils: **3 per station (R, G, B) x 6 stations = 18 logical coils**, appended to the `coil` enum immediately before
`coilCount` (`red1LightRed, red1LightGreen, red1LightBlue, ... blue3LightBlue`), mapped to frozen wire coils 16-33 by
`MayhemWireMap`, plus `stationLightsActive` at wire coil 15 so the firmware can fall back to today's alliance-derived
behaviour when talking to an FMS that does not send them. RGB-per-channel (rather than an encoded state) keeps the
firmware a dumb coil->pin copy and lets the field-testing page exercise every channel individually. Blink is done
FMS-side with `GetCycleState(2, 0, 2)`, as the green stack light already does.

Suggested semantics (pick once, document): no team / bypassed -> off; E-stop -> amber (R+G) solid; A-stop during auto ->
amber blinking; pre-match not ready -> alliance colour; pre-match ready -> green; in match linked -> green (or alliance
colour); in match robot link lost -> alliance colour blinking; FieldReset -> green, FieldVolunteers -> magenta (R+B).

Go changes:

| File / function | Change | Rough lines |
|---|---|---|
| `plc/plc.go` coil enum | 18 (+1) new constants before `coilCount` | +19 |
| `plc/plc.go` `Plc` interface + `ModbusPlc` | `SetStationLights(red, blue [3][3]bool)` (or a small `StationLight` struct) | +15 |
| `plc/coil_string.go` | regenerate: `go generate ./plc` (stringer v0.43.0 pinned in the `//go:generate` line). The generated guard `_ = x[coilCount-13]` makes the build **fail until regenerated**, so this cannot be forgotten. | generated |
| `plc/mayhem_wire_map.go` | 19 map entries | +19 |
| `field/fake_plc_test.go` | `stationLights` field + setter on `FakePlc` | +8 |
| `field/arena.go` | new `getStationLightState(station string, isRed bool) [3]bool` (~30 lines, port of the `team_sign.go` decision ladder) + 6-line call block in `handlePlcInputOutput` right after `redAllianceReady/blueAllianceReady` | +40 |
| `field/arena_test.go` | new `TestPlcStationLights` covering no-team, bypass, e-stop latch, a-stop in auto, ready, in-match link loss | +70-90 |
| `plc/plc_test.go` | extend the expected list in `TestPlcGetNames`; new `TestPlcStationLightCoils`; existing `TestPlcCoils*` index assertions are unaffected because the coils are appended | +45 |
| `web/setup_field_testing_test.go` | none expected (it does not enumerate names; verify) | 0 |

Field-testing page: **picks the new coils up automatically.** `web/setup_field_testing.go` passes `plc.GetCoilNames()`
(stringer output) to `templates/setup_field_testing.html`, which ranges over it and emits `#coil{{$i}}` buttons;
`static/js/setup_field_testing.js` `handlePlcIoChange` iterates whatever `data.Coils` contains, and the click-to-override
cycle (auto/on/off, upstream 40ba048) works per index up to `coilCount`. No template/JS/handler edits. Only cosmetic
cost: the Coils column grows from 13 to 32 rows. Overrides are rejected during a match by design.

Firmware changes (~40 lines): read coil 15; if set, `for station, channel: digitalWrite(lightPin[s][c], coilRead(16+3*s+c))`;
else keep `processStackLightOutputs()`. Add a default branch so lights do not "hold last state". Decide polarity once
(HIGH-on vs active-low) and encode it in one helper. Stations 3 need hardware (two more lights + 6 channels) or are
left unmapped.

### 4.3 Effort and risks (honest)

- Go side incl. tests, on top of the section 3 wire map: **about 1 focused day** (roughly 220-260 lines, two thirds of it
  tests). Without the wire map (just appending coils at upstream indices 13-30) it is half a day, but then the firmware
  addresses move every year when upstream changes game coils, so I do not recommend that.
- Section 3 Layer 1 itself: **0.5-1 day** including the settings plumbing and guard test.
- Firmware Layer 0 + lights: **half a day** of coding, plus **a bench session (2-3 h)** with the real opto board and
  lights; polarity and pin mapping are where time goes.
- Yearly upkeep: one merge conflict at the tail of the `coil` const block + `go generate ./plc` + the
  `TestPlcGetNames` expected list. Small, mechanical, and the compile guard catches omissions.

Risks:
1. **Ready-state semantics in 2v2**: `checkAllianceStationsReady(station)` depends on `aStopReset`; with the 2025-style
   `TwoVsTwoMode` guards R3/B3 never become ready. Fix via firmware defaults (3.1 item 2) rather than Go guards, and make
   station 3 lights "off" when `Team == nil`.
2. **Colour language**: an RGB on/off light cannot do orange distinct from yellow; teams read red as "problem", but red
   is also an alliance colour. Agree the legend with the FTA before coding; print it on the field.
3. **Latency/blink**: 100 ms PLC loop + firmware loop slowed by Serial debug; fine for status lights once
   `debug_print` is off. Blink at 2 cycles x 100 ms = 5 Hz toggle may be too fast for some lights; use
   `GetCycleState(2, 0, 5)` (1 Hz) as the hub lights do.
4. **Single-client firmware**: if a laptop runs a second FMS instance or Modbus poller, the real FMS is locked out.
5. **Not safety-rated**: lights are advisory. E-stop remains enforced by the FMS->DS path, not by the Arduino.
6. Unknown: whether the 2025 event actually ran 43e1d06 and whether the lights behaved as the code suggests (see the
   2v2 latent bug in section 2). Ask the 2025 FTA whether lights ever went green pre-match.

---

## 5. Verification checklist after regeneration

### 5.1 Unit tests (all exist upstream unless marked NEW)

`go generate ./plc && git diff --exit-code plc/*_string.go` (stringers current), then `go test ./plc ./field ./web`:

- `plc/plc_test.go`: `TestPlcInitialization`, `TestPlcGetCycleState`, `TestPlcGetNames` (update expected lists),
  `TestPlcInputs`, `TestPlcInputsGameSpecific`, `TestPlcRegisters`, `TestPlcRegistersGameSpecific`, `TestPlcCoils`,
  `TestPlcCoilsGameSpecific`, `TestPlcCoilOverrides`, `TestPlcCoilOverridesResetOnResetMatch`,
  `TestPlcInvalidCoilOverrideIndex`, `TestPlcIsHealthy`, `TestByteToBool`, `TestByteToUint`, `TestBoolToByte`.
  The `*GameSpecific` ones must be adapted/removed in step with whatever the genericising playbook does to hub I/O.
- `field/arena_test.go`: `TestArenaCheckCanStartMatch` (includes PLC healthy / field E-stop / ftaReady / ArmorBlock
  conditions), `TestPlcEStopAStop`, `TestPlcEStopAStopWithPlcDisabled`, `TestPlcFieldEStop`,
  `TestPlcFieldEStopWithPlcDisabled`, `TestPlcMatchCycleEvergreen` (stack lights, buzzer, field reset light),
  `TestPlcMatchCycleGameSpecific`. Add a 2v2 variant of `TestPlcMatchCycleEvergreen` with `FakePlc.isEnabled = true`
  asserting the stack lights **do** go green with four stations ready (this would have caught the 2025 latent bug). NEW.
- `field/fake_plc_test.go` must still satisfy `plc.Plc` (compile check).
- `web/setup_field_testing_test.go`: `TestSetupFieldTesting`, `TestSetupFieldTestingWebsocket`,
  `TestSetupFieldTestingWebsocketSetPlcCoilOverride` (+ `AllowedStates`, `DisallowedStates`, `InvalidArgs`).
- NEW: `plc/mayhem_wire_map_test.go` (frozen indices, defaults, completeness), `TestPlcStationLightCoils`,
  `TestPlcStationLights`, and a settings round-trip for `PlcWireMap` in `web/setup_settings_test.go`.

### 5.2 Bench procedure with the Arduino

Setup: laptop 10.0.100.5/24 on the same switch as the Mega (10.0.100.15); build the binary as `./cheesy-arena`; Settings ->
PLC Address = `10.0.100.15`, PLC wire map = mayhem; open `/setup/field_testing`; Arduino serial monitor at 115200.

1. **Link/health**: serial shows "new client" once and does **not** repeat (repeating = connection flapping, i.e. a
   count/exception problem). FMS log has no "PLC error reading registers". Match Play page shows PLC healthy and all four
   ArmorBlocks connected (register 0 = 65535 on the testing page).
2. **Counts**: temporarily point an FMS built from unmodified upstream at the Arduino; with Layer 0 firmware it must also
   stay healthy (proves tolerance to `registerCount`/`inputCount` growth).
3. **Inputs, one at a time** (watch the Inputs column): field E-stop; each of R1/R2/B1/B2 driver-station E-stop;
   each field-side per-team E-stop (pins 41/43/45/47) must flip the **same** team E-stop bit; each A-stop. Confirm
   polarity: pressed -> input `false` -> station shows E-stop on Match Play. Unplug an E-stop cable: must read as stopped
   (fail-safe). Stations 3 E/A inputs read `true` with nothing wired. `ftaReady` reads `true` (or follows its switch).
4. **E-stop latch**: start a test match with all stations bypassed except one simulated; press a team E-stop mid-match,
   release it; station stays E-stopped until match end (`handleTeamStop`). A-stop only latches through auto.
5. **Field E-stop** during a match aborts the match.
6. **Coils via override** (pre-match only): cycle each coil on/off from the field-testing page: stack green/red/blue,
   field reset light, and each of the 18 station-light channels; verify the right physical light and colour; verify
   "auto" returns control. Verify overrides clear on match load (`ResetMatch`).
7. **Pre-match cycle**: load a 2v2 match, nothing connected -> station lights alliance colour; bypass/connect stations one
   by one -> that station's light changes alone; all ready -> green stack state, `FieldReadyAt` set, field reset light
   off. Un-bypass one -> only that light reverts.
8. **Fallback**: with `stationLightsActive` forced off (override), lights revert to alliance-derived behaviour.
9. **Disconnect/recovery**: pull the Arduino's Ethernet for 10 s and replug; FMS must return to healthy within ~5 s
   without resetting the Arduino. Kill the FMS process without closing the socket (kill -9) and restart it; the Arduino
   must accept the new session (tests the idle-timeout fix). With a heartbeat watchdog, lights go to the safe pattern
   within 1 s of FMS loss.
10. **Soak**: leave connected 30 min with `debug_print` off; no "PLC error" lines, no client reconnects in serial log.
11. **Single-client check**: confirm no second Modbus client (another FMS instance, a test script) is attached during
    the event.
