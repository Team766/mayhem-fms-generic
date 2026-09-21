# 2v2 mode

M-Ayhem is played 2v2 some years and 3v3 others. The base supports both, chosen per event. This page is the
design that [agents/sync-upstream.md](agents/sync-upstream.md) re-applies onto each new upstream. It deliberately
differs from the 2025 implementation (`mayhem-fms-2025`), which tested the mode in about sixty places and stored a
2v2 match three different ways.

## The idea

**A 2v2 match is a match whose third team slots are 0.** Upstream already copes with an empty slot (a
practice match with five teams): the station has no team, and the operator bypasses it. So match play itself
never needs to know about a "mode". Three kinds of code change:

1. **Generators** decide how many teams go into a match. They read the event setting.
2. **The arena** has four active stations when the setting is on (R3 and B3 are inactive) and six when it is off. With the setting off it is exactly upstream, including how empty stations are handled.
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

- **Arena.** One helper, `activeStations()`, returns four stations when the setting is on and all six otherwise. Start conditions, PLC e-stop and a-stop handling, stack-light readiness, driver-station enabling and network configuration iterate over it, so R3 and B3 can never block a 2v2 match or keep the lights from going green. Nothing sets `Bypass` on the operator's behalf, and with the setting off the arena is unchanged from upstream (an empty station still needs a manual bypass, as today).
- **Stored data.** Empty is always 0: schedules, playoff lineups, substitutions, edited results. No filler surrogate team, no duplicated captain. Alliance updates after a playoff match ignore 0.
- **Panels and displays.** Two fixed layouts. With the setting on, every page gets a `two-v-two` class (or template branch) that removes the third robot's row and controls and rebalances the spacing; with it off, pages are upstream's. One mechanism everywhere, not a mix of CSS, template and JS checks. The audience final score keeps upstream's fourth (off-field) row in both.
- **Guard.** With the setting on, a third team cannot be entered: substitution and the edit-result form reject a non-zero third slot, and loading a match that has one reports an error instead of showing a broken layout.
- **Game code.** Per-robot arrays stay length 3; position 3 is unused in a 2v2 match. Rules that say "all robots" count the robots present and not bypassed. A game spec may give different thresholds per alliance size; otherwise they are the same.

## Out of scope

- Alliances of two plus a backup.
- Schedule *templates*. The 2v2 PR makes the schedule builder use `2p_` templates; a follow-up PR adds a template generator and checked-in templates for a range of team counts and matches per team, so a 2v2 schedule can be regenerated from the tool in seconds when a team drops out, the way 3v3 works upstream.
- Station-3 lights and mixed 2v2/3v3 schedules (the design does not prevent the latter, it is just untested).

## Verification

Automated, in `go test ./...`:
- Upstream's suite passes unchanged with the setting off.
- Arena: loading a match with empty third slots bypasses them; start conditions pass with four robots; with a PLC, readiness reaches green and an e-stop on an empty station is ignored.
- Schedule: every generated 2v2 match has four distinct teams, zero third slots, no surrogates on them.
- Alliance selection and playoffs: alliances of two for each playoff type; no team 0 written back after a playoff match.
- Match review: editing a 2v2 result keeps the third slots 0.

Manual, on one database: turn the setting on, generate a schedule, play and commit a qualification match from the panels, check every display and report, run alliance selection and one playoff match. Then turn it off and confirm a 3v3 match looks and scores exactly like upstream.
