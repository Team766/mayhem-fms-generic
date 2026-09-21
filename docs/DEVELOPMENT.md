# Developing the M-Ayhem FMS

How this repository relates to upstream Cheesy Arena, what it removes and adds, where a year's game lives, and how
both are kept current. Read this before changing anything.

This repository is Team 766's base for the Mechanical M-Ayhem field management system. It is
[Team254/cheesy-arena](https://github.com/Team254/cheesy-arena) **minus** a strip list, **plus** a short
feature list, with a small placeholder game. Each year's FMS is this base with that year's game applied.

- How the base is kept current: [agents/sync-upstream.md](agents/sync-upstream.md). Checkpoint: [`UPSTREAM.md`](../UPSTREAM.md).
- How a year's game is applied: [agents/apply-game.md](agents/apply-game.md), from a spec ([agents/game-spec-format.md](agents/game-spec-format.md)).
- Detailed, dated inventories behind this page: [agents/reference/](agents/reference/).

## Why it works this way

Earlier attempts: a fork of Cheesy Arena Lite (2024; Lite was dormant and lagged), a hand-maintained generic
fork (2025; merging a year of upstream hit 39 conflicted files, all in game code upstream rewrites every
January), a config-to-code generator and then a runtime YAML game engine (2026; thousands of lines of
machinery to avoid about a thousand lines of yearly edits, and still unable to express the real game).
Team 254 now maintains Lite by having an agent port game-agnostic upstream changes from a recorded
checkpoint. This base does the same, and adds a second playbook for the game.

## Strip list (removed completely: code, settings, templates, routes, JS, tests)

| Removed | Notes |
|---------|-------|
| The Blue Alliance | Publishing and team-info download. Keep `model.TbaMatchKey` (inert, threaded through `playoff/`), team avatars (`static/img/avatars`, avatar API; move `AvatarsDir` out of `partner/`) |
| Nexus | Lineups and auto-queue |
| Team signs | `field/team_sign.go` and settings |
| Twitch display | Handler, template, JS, display type |
| Upstream's season game | Scoring model, arena hooks, game PLC I/O, game settings, sounds, artwork |
| LED/DMX | `led/`, `field/arena_leds.go`, settings, field-testing UI. This is 2026 upstream code that drives DMX lighting on that season's field element; it is not the team signs |

## Keep list (as upstream, even though Lite removed some)

Fouls, the rules list and the full referee flow; per-position scoring panels; cards and playoff DQ;
playoffs and alliance selection; every display; reports; match logs; awards, lower thirds, sponsor slides;
network (AP, switch); PLC generic signals and the field-testing page; Companion and Blackmagic clients;
driver-station protocol code exactly as upstream.

## Never reintroduce

The identifier lists are in [agents/reference/strip-list.md](agents/reference/strip-list.md) section 5 and are
checked by the leftover grep in `agents/sync-upstream.md`. In short: nothing from the strip list, nothing
Lite-specific (`/api/scores`, points-only `Score`, renamed match-state strings, Lite naming), no generic
game engine, no build tags or mode flags that switch between games.

## Feature list (added on top of upstream)

| Feature | Contract |
|---------|----------|
| 2v2 mode | [TwoVTwo.md](TwoVTwo.md) |
| M-Ayhem Arduino PLC, unmodified | Below |
| Single-game playoff rounds | A third playoff type: single elimination where every round before the final advances on one win; the final stays best-of-three. Implemented by setting `NumWinsToAdvance = 1` on the non-final matchups; the existing matchup logic then schedules one match and reveals the next only after a tie. No other bracket changes |
| Per-team station lights | Planned; design notes in [agents/reference/plc.md](agents/reference/plc.md), which also has the PLC bench test |

### PLC

The field uses Team 766's Arduino Modbus PLC (`fakeplc-arduino`, branch `plc-cheesy-arena-compat`, as flashed for
M-Ayhem 2025), not upstream's Allen-Bradley program. **The FMS's PLC code is upstream's, unchanged, as it was in
2025**, and the Arduino needs no reflashing. Three facts make that work:

- The firmware serves fixed addresses that match upstream's generic signal order (field e-stop, team e-stops and a-stops, connected inputs, the field I/O register, heartbeat, match reset, stack lights, buzzer, field reset light). `plc/mayhem_arduino_test.go` fails if an upstream sync moves any of them or grows a table past what the firmware serves. Season I/O always goes after the generic block, and the strip removes it.
- The hardware has no station-3 stop wiring, so those inputs read as pressed. 2v2 mode ignores R3 and B3, so **the supported configuration with the PLC enabled is 2v2**. 3v3 with this PLC would show station 3 as e-stopped.
- Upstream's "FTA ready" switch (a start permission added in May 2026, not a safety stop) does not exist on this field. Its start condition and Match Play badge are removed; the PLC input itself stays in the signal list so nothing is renumbered.

Do not resurrect the 2025 remap attempts (`InputMap`/`CoilMap`, `getInputPin`/`getCoilPin`, `plc/mayhem_plc.go`).

## The current game and the game seam

The tree always carries exactly one game: the current year's. Its spec is checked in under `specs/`, and
`specs/CURRENT` names it. `apply-game` takes the new spec, reads the `CURRENT` spec to learn what is being
replaced (its ids are the vocabulary to remove), and then points `CURRENT` at the new spec. Old specs stay in
`specs/` as history and examples. There is no separate placeholder game and no maintained word list.

`specs/high_seas_havoc.md` (M-Ayhem 2025) is kept as a second example spec; it has no code in this tree.

### The neutral state (no game)

Right after a regeneration, before `apply-game` runs, the tree has no game. It must still build, pass its
tests and run a match. "No game" means:

- `game.Score` holds only `Fouls` and `PlayoffDq`. `Summarize()` computes foul points from the opponent's fouls, `NumOpponentMajorFouls`, `MatchPoints = 0`, `Score = FoulPoints`, and no bonus ranking points. No game settings, thresholds or special-case rules.
- Ranking: win 3 RP, tie 1. Sort by ranking points per match, then match points per match, then the random value. Playoff tie: fewer major fouls committed, otherwise a tie.
- `game/rule.go` keeps the `Rule` type and lookups with a two-entry list (one minor, one major) so the referee flow stays testable. `game/foul.go` keeps upstream's point values and has no rule-specific exceptions.
- Match timing is auto, pause, teleop and a warning time before the end; sounds are start, resume, warning, end, abort and the non-match sounds. No shifts.
- Scoring panels show the match, the foul dialog and Commit; the referee panel shows cards and fouls; the audience overlay shows teams, score and timer; the final score shows the score, a Foul row and ranking points or playoff wins; edit-result has fouls and cards; rankings and reports have Rank, Team, RP, Match points, W-L-T, DQ, Played.
- The arena has no game hooks, the PLC has only the generic signals, and there is no "Game-Specific" settings fieldset.

Files a game may touch (the seam):

- `game/`: `score.go`, `score_summary.go`, `ranking_fields.go`, `match_timing.go`, `match_sounds.go`, `rule.go` (list), `foul.go` (point values), `test_helpers.go`, and tests.
- Entry: `web/scoring_panel.go` (commands, positions), `templates/scoring_panel.html`, `static/js/scoring_panel.js`, `static/css/scoring_panel.css`; status rows in `templates/referee_panel.html` and `static/js/referee_panel.js`; `templates/edit_match_result.html`, `templates/match_review.html`, `static/js/match_review.js`, game parts of `web/match_review.go`.
- Score consumers (**all of them, every time**): `templates/audience_display.html`, `static/js/audience_display.js`, `static/js/display_shared.js`, `templates/wall_display.html`, `static/js/wall_display.js`, `templates/announcer_display_score_posted.html`, `static/js/announcer_display.js`, `static/js/alliance_station_display.js`, `templates/rankings_display.html`, `static/js/rankings_display.js`, ranking columns in `web/reports.go` and `templates/rankings.csv`, RP symbols in `templates/match_review.html`.
- Settings: the "Game-Specific" fieldset in `templates/setup_settings.html` with its `EventSettings` fields and `LoadSettings()` lines, only for thresholds the spec marks tunable.
- Fixture fallout in `tournament/`, `web/`, `model/`, `field/` tests.
- Assets: `static/img/game-logo.png`, `static/img/blinds-logo.png`, sounds named by the spec.

Anything else changed by a game is a defect unless the PR calls it out.

## Files the base owns (survive a regeneration untouched)

`docs/DEVELOPMENT.md`, `docs/TwoVTwo.md`, `docs/agents/`, `specs/`, `UPSTREAM.md`, the M-Ayhem
section of `AGENTS.md`, `.claude/skills/` (thin wrappers), `README.md` identity, `schedules/2p_*.csv`.

## Open decisions

1. ArmorBlock `redIoLink`/`blueIoLink` inputs block match start when the PLC is enabled: generic field hardware, or strip with the season game?
2. Keep driver-station game-data plumbing with an empty payload (recommended) or remove as Lite did.
3. Keep a neutral event-code setting for the driver-station event-name packet once TBA settings are gone.
