# 2v2 mode

M-Ayhem is sometimes played 2v2 and sometimes 3v3. The base supports both permanently, selected per event.
This document is the contract that [agents/sync-upstream.md](agents/sync-upstream.md) re-applies onto each
new upstream. The file-by-file inventory of the 2025 implementation, and how it maps onto upstream as of
September 2026, is in [agents/reference/two-v-two.md](agents/reference/two-v-two.md).

## Design

1. **One event setting.** `EventSettings.TwoVsTwoMode bool`, default false, on Setup > Settings. It is not a build flag: one binary and one test run cover both sizes.
2. **Zero impact when off.** With the setting off, behaviour and screens are upstream's 3v3, including the fourth (off-field) team row on the audience final score.
3. **No schema change.** `model.Match` keeps six team slots.
4. **One representation: `Red3 = Blue3 = 0`.** A 2v2 match has no third team anywhere it is stored: scheduled matches, playoff matches, substitutions. No surrogate filler team, no duplicated captain. (The 2025 code used all three representations; that is what made later screens fragile.)
5. **One helper decides who is playing.** Code asks the arena or the match for its *active stations* (four in 2v2, six in 3v3) instead of testing the setting. Match-start checks, PLC stop handling, stack-light readiness, network setup, displays and game rules all go through it. New upstream code that loops over stations or names `R3`/`B3` is adapted the same way.
6. **Generalise rather than hide.** Prefer rendering from the list of active teams over CSS that hides a third control. Where a game screen has per-robot controls, the game playbook renders one per active robot.
7. **Games count real robots.** Rules such as "all robots left" use the active, non-bypassed robots. Thresholds may differ by alliance size only if the game spec says so.

## What it touches

| Area | Requirement |
|------|-------------|
| Settings | Checkbox, persisted, sent to clients in the match-load message |
| Schedule generation | `2p_*` templates; third slots zero, surrogate flags cleared; choose template by team count and matches per team |
| Arena | Load, substitute and reset leave station 3 empty and bypassed; start conditions, PLC stops and **stack-light readiness** use active stations; network config skips empty stations |
| Alliance selection and playoffs | Alliance size 2; third-round setting cannot silently make it 4; playoff lineups and post-match alliance updates never write team 0 |
| Panels | Scoring, referee (cards, fouls by team), head-ref bypass: active robots only |
| Displays | Audience, announcer, wall, alliance station, queueing, field monitor (both), FTA views, bracket: lay out two teams per alliance cleanly |
| Match review and logs | Edit form and logs show active teams only; editing does not resurrect a third team |
| Reports | Schedule and team reports drop the third columns |

## Known gaps (follow-up work, not blockers for regeneration)

- **Schedule generation is minimal**: one 14-team template, cut at 8, 10 or 12 matches per team, no generator for other team counts, no tests. `rand.Seed(0)` no longer makes it deterministic under current Go. Planned as its own PRs after the base exists.
- "Two plus a backup" alliances are not supported.
- Station-3 per-team lights need hardware that does not exist yet (see [BASE.md](BASE.md), PLC).

## Defects in the 2025 implementation that must not be carried forward

- Stack lights never turn green with the PLC enabled in 2v2: station-3 stop handling was skipped, so those stations never counted as reset, and readiness still required them. Fixed by design point 5.
- Audience final score lost the off-field team row in both modes.
- `tournament` tests did not compile after the schedule function signature changed.
- Three different stored representations of a 2v2 match (design point 4).

## Verification

Automated (all in `go test ./...`):
- Existing upstream suite unchanged and green with the setting off.
- Arena: load/substitute/reset in 2v2; start conditions with four stations; PLC readiness and stack lights reach green in 2v2.
- Schedule: every generated 2v2 match has zero third slots, no surrogates, four distinct teams.
- Alliance selection: size 2 with each playoff type; no team 0 after a playoff substitution.
- Game: worked examples from the spec in both sizes.

Manual, on one database: turn 2v2 on; generate a schedule; play a qualification match from the panels; commit; check audience, announcer, wall, queueing, field monitors, rankings, reports; run alliance selection and one playoff match; then turn 2v2 off and confirm a 3v3 match looks and scores exactly like upstream.
