# 2v2 mode

M-Ayhem is played 2v2. Upstream Cheesy Arena is 3v3, and we keep that intact so an event could use it if it ever
needed to. 2v2 is added on top as an event setting; with the setting off, the FMS is upstream's 3v3.

This page is the design that [agents/sync-upstream.md](agents/sync-upstream.md) re-applies onto each new upstream.

## The idea

**A 2v2 match is a match whose third team slots are 0.** Three kinds of code know about 2v2:

1. **Generators** decide how many teams go into a match, and read the setting.
2. **The arena** ignores the empty third stations in 2v2, so they can never block a match.
3. **Layouts** (panels, displays, reports) have exactly two shapes, 2v2 and 3v3, chosen by the setting.

Layouts follow the setting, never team numbers: team 0 also appears in the Test Match, in playoff matches before
their alliances are known, and in a practice match that is short a team.

## The setting

`EventSettings.TwoVsTwoMode`, default off, a checkbox on Setup > Settings. It is sent to pages with the match, and
read by the generators:

| Generator | With the setting on |
|-----------|---------------------|
| Schedule builder | Uses `schedules/2p_<teams>_<matchesPerTeam>.csv`: four teams per match, third slots 0 |
| Alliance selection | Alliances of two; no third or backup round |
| Playoff matches | Lineups of two; the third slot is 0 |

## Everywhere else

- **Arena.** One helper, `activeStations()`, decides which stations are in play: all six in 3v3; in 2v2, all but an *empty* R3 or B3 (no team and no driver station connected). Start conditions, PLC e-stop and a-stop handling and stack-light readiness use it. A station with a team is therefore always checked and always stoppable, whatever the setting; the setting only takes empty stations out of play. Driver-station packets and network configuration are upstream's and already act only on stations with a team. Nothing bypasses a station on the operator's behalf.
- **Stored data.** Empty is always 0: schedules, playoff lineups, substitutions, edited results. Alliance updates after a playoff match ignore 0.
- **Panels and displays.** With the setting on, pages get a `two-v-two` class (or a template branch) that removes the third robot's row and controls. With it off, pages are upstream's. The audience final score keeps its fourth (off-field) row in both.
- **Guard.** With the setting on, a third team cannot be entered: substitution and the edit-result form reject one, and loading a match that has one is an error.
- **Game code.** Per-robot arrays stay length 3; position 3 is unused in 2v2. Rules about "all robots" count the robots actually playing.

## Out of scope for now

- Alliances of two plus a backup.
- A 2v2 schedule generator. Today there are checked-in templates for 14 teams (8, 10 or 12 matches per team). A follow-up adds a generator and more team counts, so a schedule can be rebuilt in seconds when a team drops out, as 3v3 works upstream.
- Station-3 lights and mixed 2v2/3v3 schedules.

## Verification

Automated, in `go test ./...`:
- Upstream's tests pass unchanged with the setting off.
- Arena: in 2v2 the match starts with four robots and, with a PLC, the stack lights go green; an e-stop on an occupied station always stops it, including after the setting is switched with a 3v3 match loaded.
- Schedule: every 2v2 match has four distinct teams and zero third slots, and every team plays the same number of matches.
- Alliance selection and playoffs: alliances of two with each playoff type; no team 0 written back.
- Match review: editing a 2v2 result keeps the third slots 0.

Manual, on one database: turn the setting on, generate a schedule, play and commit a match from the panels, check every display and report, run alliance selection and a playoff match. Then turn it off and check a 3v3 match looks and scores like upstream.
