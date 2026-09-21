# 2v2 mode

M-Ayhem is played 2v2 some years and 3v3 others. The base supports both, chosen per event. This page is the
design that [agents/sync-upstream.md](agents/sync-upstream.md) re-applies onto each new upstream. What the
2025 code did, file by file, is in [agents/reference/two-v-two.md](agents/reference/two-v-two.md); this
design deliberately differs from it.

## The idea

**A 2v2 match is a match whose third team slots are 0.** Upstream already copes with an empty slot (a
practice match with five teams): the station has no team, and the operator bypasses it. So match play itself
never needs to know about a "mode". Three kinds of code change:

1. **Generators** decide how many teams go into a match. They read the event setting.
2. **The arena** looks at the match in front of it: a station is *empty* when its team is 0, whatever the mode.
3. **Layouts** (panels, displays, reports) have exactly two shapes, 2v2 and 3v3, chosen by the event setting. They never infer a shape from which teams happen to be 0.

Team 0 already means other things upstream, which is why layouts must not key off it: the Test Match has six
zeros, playoff matches have zeros until their alliances are known, a practice match can be short a team, and
the off-field fourth playoff team is 0 when there is none. Treating those as "2v2" would reshape screens at
the wrong moments, and a 2v3 match would get a lopsided layout.

## The setting

`EventSettings.TwoVsTwoMode bool`, default false, a checkbox on Setup > Settings. An event setting rather
than a build flag so one binary and one test run cover both sizes. It is sent to clients in the match-load
message for layout, and read by three generators:

| Generator | With the setting on |
|-----------|---------------------|
| Schedule builder (`tournament/schedule.go`) | Uses `schedules/2p_<teams>_<matchesPerTeam>.csv`, four teams per match, `Red3 = Blue3 = 0`, no surrogate flags on the empty slots |
| Alliance selection (`web/alliance_selection.go`) | Alliances of two; no third or backup round |
| Playoff match creation | Lineups of two; the third lineup slot is 0 |

## Everywhere else

- **Arena.** Loading or substituting a match leaves a station with team 0 empty and **bypassed automatically**. One helper (`AllianceStation.IsEmpty()`, or an `activeStations()` list) is used by the start conditions, PLC e-stop and a-stop handling, stack-light readiness and network configuration, so an empty station can never block a match or keep the lights from going green. This also fixes the five-team practice match in 3v3.
- **Stored data.** Empty is always 0: schedules, playoff lineups, substitutions, edited results. No filler surrogate team, no duplicated captain. Alliance updates after a playoff match ignore 0.
- **Panels and displays.** Two fixed layouts. With the setting on, every page gets a `two-v-two` class (or template branch) that removes the third robot's row and controls and rebalances the spacing; with it off, pages are upstream's. One mechanism everywhere, not a mix of CSS, template and JS checks. The audience final score keeps upstream's fourth (off-field) row in both.
- **Guard.** With the setting on, a third team cannot be entered: substitution and the edit-result form reject a non-zero third slot, and loading a match that has one reports an error instead of showing a broken layout.
- **Game code.** Per-robot arrays stay length 3; position 3 is unused in a 2v2 match. Rules that say "all robots" count the robots present and not bypassed. A game spec may give different thresholds per alliance size; otherwise they are the same.

## Out of scope

- Alliances of two plus a backup.
- A schedule generator. Templates are pre-generated files; today only `2p_14_*.csv` exists, so **a template for the event's team count must be added before the event**. Better 2v2 scheduling is follow-up work.
- Station-3 lights and mixed 2v2/3v3 schedules (the design does not prevent the latter, it is just untested).

## Verification

Automated, in `go test ./...`:
- Upstream's suite passes unchanged with the setting off.
- Arena: loading a match with empty third slots bypasses them; start conditions pass with four robots; with a PLC, readiness reaches green and an e-stop on an empty station is ignored.
- Schedule: every generated 2v2 match has four distinct teams, zero third slots, no surrogates on them.
- Alliance selection and playoffs: alliances of two for each playoff type; no team 0 written back after a playoff match.
- Match review: editing a 2v2 result keeps the third slots 0.

Manual, on one database: turn the setting on, generate a schedule, play and commit a qualification match from the panels, check every display and report, run alliance selection and one playoff match. Then turn it off and confirm a 3v3 match looks and scores exactly like upstream.
