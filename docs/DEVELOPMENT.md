# Developing the M-Ayhem FMS

How this repository relates to upstream Cheesy Arena, what it removes and adds, where a year's game lives, and how
both are kept current. Read this before changing anything.

This is Team 766's field management system for Mechanical M-Ayhem. It is
[Team254/cheesy-arena](https://github.com/Team254/cheesy-arena) **minus** a few things M-Ayhem does not use,
**plus** a few it needs, with the current year's game applied on top.

- Keeping up with upstream: [agents/sync-upstream.md](agents/sync-upstream.md). The checkpoint is in [`UPSTREAM.md`](../UPSTREAM.md).
- Applying a year's game from a spec: [agents/apply-game.md](agents/apply-game.md); spec format in [agents/game-spec-format.md](agents/game-spec-format.md).

This page describes intent and the checks that prove it. It does not list files or call sites: upstream moves
them every year, and `git grep` finds them faster and more reliably than a list that has gone stale.

## Why it works this way

We never merge upstream. Upstream rewrites its game code every January, so a fork that also changes that code
conflicts everywhere, and separate per-year repos made catching up harder still. Instead, this one repo holds the
current year's FMS, and two playbooks re-apply our changes onto each new upstream: one for the base, one for the
year's game. Team 254 maintains Cheesy Arena Lite the same way.

## What we remove

Removed means gone completely: no code, settings, page sections, routes, scripts, styles, enum values or tests
left behind. Each entry says what must survive and how to tell the job is done.

| Removed | Why | Must survive | Done when |
|---------|-----|--------------|-----------|
| The Blue Alliance | M-Ayhem is not a published FRC event | Match keys on matches and in the playoff code (inert, deeply threaded). Team avatars served from disk. Adding teams by hand or CSV. An event code setting, because the driver station's event-name packet needs one | No settings, pages, routes or background tasks mention TBA or publishing; adding a team works offline |
| Nexus | Not used | Manual substitution on Match Play | No lineup fetching or auto-queue |
| Team signs | The field has none | The alliance station *displays* (browser pages), which are a different thing | No sign settings or sign updates in the arena loop |
| Twitch display | Not used | Every other display type | The display type is gone from the list of displays |
| Upstream's season game | Replaced by ours | Everything generic listed under "What we keep" | The tree is in the no-game state below |
| LED/DMX control | It drives lighting on upstream's season field element; we have no such hardware | The PLC's stack lights and field-reset light | No LED package, settings or field-testing controls |

Also never bring in: anything specific to Cheesy Arena Lite (its points-only score, its score API, its renamed
match-state strings, its naming), a generic or configurable game engine, a code generator, or build tags and
mode flags that switch between games.

## What we keep, exactly as upstream

Fouls, the rules list and the whole referee flow; per-position scoring panels and their commit status; cards
and playoff disqualification; playoffs and alliance selection; every display; reports; match logs; awards,
lower thirds and sponsor slides; network configuration; the PLC's generic signals and the field-testing page;
the Companion and Blackmagic clients; the driver-station protocol code.

## What we add

| Feature | Contract |
|---------|----------|
| 2v2 mode | [TwoVTwo.md](TwoVTwo.md) |
| The M-Ayhem Arduino PLC, unmodified | Below |
| Single-game playoff rounds | A third playoff type: single elimination where every round before the final is decided by one game and the final stays best-of-three. It reuses upstream's bracket and only changes how many wins a pre-final series needs. If a game ends in a tie, the next game of that series is played |
| Per-team station lights | Planned; notes in [ArduinoPlc.md](ArduinoPlc.md) |

### PLC

The field uses Team 766's Arduino Modbus PLC, not upstream's Allen-Bradley program. **The FMS's PLC code is
upstream's, unchanged, as it was in 2025**, and the Arduino runs as flashed for M-Ayhem 2025. That works because:

- The firmware serves fixed addresses that match upstream's generic signal order. A guard test (`plc/mayhem_arduino_test.go`) fails if an upstream sync moves one of those signals or grows a table past what the firmware serves. Season signals always come after the generic ones, and we remove them.
- The hardware has no station-3 stop wiring, so those inputs read as pressed. 2v2 mode ignores an empty station 3, so **the supported configuration with the PLC enabled is 2v2**. (A station 3 that holds a team would read as e-stopped.)
- Upstream's "FTA ready" switch (a start permission, not a safety stop) does not exist on this field, so its start condition and Match Play badge are removed. The input stays in the signal list so nothing is renumbered.

The bench test is in [ArduinoPlc.md](ArduinoPlc.md). Run it after any sync that touches the PLC or the arena's PLC handling.

## The game

The tree always carries exactly one game: the current year's. Its spec is in `specs/`, and `specs/CURRENT` names
it. `apply-game` reads the `CURRENT` spec to learn what is being replaced, applies the new spec, and points
`CURRENT` at it. Old specs stay as history and examples.

**What counts as game code (the seam):** the score model and its math, ranking order and tiebreaks, the rules
list and foul values, match timing and sounds; every screen or report that lets someone enter a score, or that
reads the score, the summary or the rankings; tunable game settings; the game's logos and sounds; and the test
fixtures that embed a score. Wiring the game's timing and tunable settings into the arena's settings loader counts
as game code too. A game change that touches anything else (arena behaviour, the PLC, networking, playoffs, 2v2)
is a base change and gets its own pull request.

**The no-game state.** Right after a regeneration, before a game is applied, the tree must still build, pass its
tests and run a match: a score is only fouls and the playoff disqualification flag; ranking is by ranking points
(win 3, tie 1), then match points; a tied playoff match goes to the alliance with fewer major fouls; timing is
auto, pause, teleop and a warning; the scoring panel shows only Commit; displays show teams, score and timer;
the rules list keeps only the general rules ([`specs/GENERAL_RULES.md`](../specs/GENERAL_RULES.md)) and foul values stay upstream's, until a game replaces them; no game settings.

## What survives a regeneration untouched

`docs/`, `specs/`, `UPSTREAM.md`, the M-Ayhem section of `AGENTS.md`, `.claude/skills/` (thin pointers to the
playbooks), and the 2v2 schedule templates.

## Open decisions

1. The ArmorBlock link inputs block match start when the PLC is enabled. The Arduino reports them as connected, so they are harmless today; decide whether they are generic or belong to upstream's field hardware.
2. Driver-station game data: kept as upstream with an empty payload. Revisit only if a game needs it.
