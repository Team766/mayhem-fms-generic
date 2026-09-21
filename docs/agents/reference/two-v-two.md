# 2v2 on top of Cheesy Arena: inventory for regeneration

Sources (all in `../mayhem-fms-generic`, read-only): fork point `aa1cd2b` (2025-05-27), `fms2025/main` (3bc6890, source of truth), generic commits `38f8664`, `83b987f`, `a2a286c`, `upstream/main` (f2f38f4, 2026-09-19), untracked `docs/TwoVTwo.md`.
Method: `git diff aa1cd2b fms2025/main -- <path>`, `git grep TwoVsTwo fms2025/main`, `git show upstream/main:<path>`. Nothing was run/compiled; statements about test compilation are from reading call sites.

Note: the `aa1cd2b..fms2025/main` diff is dominated by non-2v2 work (M-Ayhem game swap, TBA/Nexus/team-sign/Twitch removal, GreenScreen, RP rename). Only the 2v2 parts are inventoried here.

---

## 1. Design as implemented

**Setting.** `TwoVsTwoMode bool` on `model.EventSettings` (`model/event_settings.go`, placed after `SelectionShowUnpickedTeams`, next to the unrelated `GreenScreen`). Persisted with the rest of EventSettings (single JSON row in the bbolt table; no migration needed; zero value = `false` = stock 3v3). No explicit default in `GetEventSettings()`. Set by checkbox `name="twoVsTwoMode"` in `templates/setup_settings.html` (labelled "2v2 Mode", in the top event section beside "Green Screen"), read in `web/setup_settings.go: settingsPostHandler` as `r.PostFormValue("twoVsTwoMode") == "on"`. It is global to the event, not per match.

**Schema.** Unchanged. `model.Match` keeps `Red1..3/Blue1..3` and `*IsSurrogate`; `game.Score` keeps `[3]` arrays; `model.Alliance.Lineup` stays `[3]int`; `arena.AllianceStations` always has 6 entries; PLC/AP/switch APIs stay 3-per-alliance / `[6]`.

**How a 2v2 match is actually represented (three different conventions coexist, this is the main inconsistency):**

| Match source | Red3 / Blue3 stored in DB | Where |
|---|---|---|
| Qualification/practice from the 2v2 scheduler | NOT 0. Both are the real team that anonymous slot "1" maps to, with `Red3IsSurrogate = Blue3IsSurrogate = true` (every row of `schedules/2p_14_0.csv` has `...,1,1,...,1,1`). Rankings ignore them only because `CalculateRankings` skips surrogates (comment added in `tournament/qualification_rankings.go`). | `tournament/schedule.go: BuildRandomSchedule`, `schedules/2p_14_0.csv` |
| Playoff matches | NOT 0. `alliance.Lineup[2] = alliance.TeamIds[0]` (captain duplicated; so Red2 == Red3 == captain) with comment "Doesn't matter, as this is 2v2". | `web/alliance_selection.go: allianceSelectionFinalizeHandler`, consumed by `playoff/playoff_tournament.go: positionRedTeams/positionBlueTeams` |
| Any match after a manual substitution on Match Play (practice/playoff/test) | 0. JS sends `Red3: twoVsTwoMode ? 0 : ...`. | `static/js/match_play.js: substituteTeams` (commit a8040ed) |

`docs/TwoVTwo.md` says Red3/Blue3 should be 0 and surrogate flags cleared; that is NOT what fms2025/main does for scheduled or playoff matches.

**Runtime invariants that do hold in 2v2 (enforced in `field/arena.go`):**
- After `LoadMatch` and `SubstituteTeams`: `AllianceStations["R3"/"B3"]` have `Team = nil`, `DsConn` closed and nil, `Bypass = true` (regardless of what Red3/Blue3 hold in the match record).
- `ResetMatch` sets `R3/B3.Bypass = TwoVsTwoMode` (others false).
- `checkCanStartMatch` only checks `R1,R2,B1,B2`.
- `preLoadNextMatch` and `setupNetwork` nil out `teams[2]` and `teams[5]` before `accessPoint.ConfigureTeamWifi` / switch config, so no SSID/VLAN for the third stations.
- `handlePlcInputOutput` does not call `handleTeamStop` for R3/B3, so their E-stop/A-stop inputs are ignored.
- `GenerateMatchLoadMessage` (`field/arena_notifiers.go`) carries `TwoVsTwoMode` to clients (used by `match_play.js`).
- Scoring: no robot-count logic at all. M-Ayhem 2025 `game.Score.Summarize` just loops the `[3]bool` status arrays (third stays false) and RP thresholds are absolute counts. The third-robot inputs are hidden in the UI only.

UI convention: server-side `{{if not .EventSettings.TwoVsTwoMode}}` around third-slot markup, or `data-two-v-two="true"` attribute on the panel root + CSS `display:none` for `.team-3*`, or body class `two-vs-two` for display CSS. Two handlers instead pass `NumTeamsPerAlliance` (2|3) and templates use `{{range $i := seq .NumTeamsPerAlliance}}`.

---

## 2. Touchpoints (fms2025/main vs aa1cd2b; 2v2-relevant parts only)

### Model / settings
| File | Change | Why |
|---|---|---|
| `model/event_settings.go` | `TwoVsTwoMode bool` field | the switch |
| `web/setup_settings.go` `settingsPostHandler` | parse `twoVsTwoMode` checkbox | |
| `templates/setup_settings.html` | "2v2 Mode" checkbox | |
| `field/arena_notifiers.go` `GenerateMatchLoadMessage` | add `TwoVsTwoMode` to MatchLoad payload | JS needs it (match_play.js) |

### Schedule generation
| File | Change | Why |
|---|---|---|
| `tournament/schedule.go` `BuildRandomSchedule(teams, blocks, matchType, twoVsTwo bool)` | new 4th param; `localTeamsPerMatch` 4 vs 6 for matchesPerTeam/numMatches math; early return on 0 matches; 2v2 opens `schedules/2p_<numTeams>_0.csv` (matches-per-team NOT in the filename); 2v2 accepts a file with `>= numMatches` rows and truncates (3v3 still requires exact); `rand.Seed(0)` before `rand.Perm` and re-seed with time afterwards (deterministic team shuffle "for emergencies"); error message now includes filename. Still parses 12 columns and still assigns Red3/Blue3 from the CSV. | reuse the 3v3 loader with a single long template that can be cut at 8/10/12 matches per team |
| `schedules/2p_14_0.csv` | the only 2v2 template: 14 teams, 42 rows, 12 columns, slots 3 and 6 always `1,1` (team #1 as surrogate placeholder). Commit c1ac66f: "can be stopped at 8, 10, or 12 games per team". | |
| `web/setup_schedule.go` `scheduleGeneratePostHandler` | passes `web.arena.EventSettings.TwoVsTwoMode` | |
| `tournament/qualification_rankings.go` `CalculateRankings` | comment only | documents reliance on surrogate flag |

### Match play / arena
| File | Change | Why |
|---|---|---|
| `field/arena.go` `LoadMatch` | after the six `assignTeam` calls: for R3/B3 close DsConn, `Team=nil`, `Bypass=true` | third stations inert |
| `field/arena.go` `SubstituteTeams` | same block after assignTeam | |
| `field/arena.go` `ResetMatch` | `R3/B3.Bypass = TwoVsTwoMode` | keep bypass across resets |
| `field/arena.go` `preLoadNextMatch`, `setupNetwork` | nil `teams[2]`, `teams[5]` | no AP/switch config for third stations |
| `field/arena.go` `checkCanStartMatch` | station list `R1,R2,B1,B2` in 2v2 | ready check |
| `field/arena.go` `handlePlcInputOutput` | skip `handleTeamStop("R3"/"B3")` | ignore unwired E/A-stops. Stack-light readiness (`checkAllianceStationsReady("R1","R2","R3")`) was NOT changed, see gaps |
| `plc/plc.go` | no 2v2 change (diff is game cleanup) | |
| `templates/match_play.html` | `id="mainPanel" data-two-v-two=...`; R3/B3 `matchPlayTeam` rows wrapped in `{{if not ...}}` | hides rows and bypass toggles |
| `static/js/match_play.js` `substituteTeams`, `handleMatchLoad` | send `Red3/Blue3 = 0` in 2v2; read `data.TwoVsTwoMode` | hidden inputs do not exist |
| `field/arena_test.go` | `TestTwoVsTwo_CheckCanStartMatch`, `_LoadMatchBypassThirdStations`, `_SubstituteTeamsBypassThirdStations`, `_HandlePlcSkipsThirdStations` | |

### Alliance selection / playoffs
| File | Change | Why |
|---|---|---|
| `web/alliance_selection.go` `allianceSelectionStartHandler` | `teamsPerAlliance := 3; if TwoVsTwoMode {2}; if SelectionRound3Order != "" {4}` | alliance size 2 |
| same, `allianceSelectionFinalizeHandler` | `Lineup[2] = TeamIds[0]` in 2v2 (avoids index out of range on a 2-element TeamIds) | |
| same, `determineNextCell` | skip the "third column" (TeamIds[2]) block in 2v2 | |
| same, `allianceSelectionGetHandler` | unrelated fix: render a message for non-admins | |
| `templates/alliance_selection.html` | hide `<th>Pick 2</th>` in 2v2 (cells already range over TeamIds) | |
| `playoff/*`, `model/alliance.go` | untouched | backups: no 2v2-specific handling; `UpdateAllianceFromMatch` and `GetOffFieldTeamIds` still take 3 ids |

### Scoring / RP logic
None in Go. `game/score.go`, `game/mayhem.go` have no robot-count-dependent logic (thresholds are absolute: `AutonRankingPointThreshold=20`, `ScoringRankingPointThreshold=14`, `EndgameRankingPointThreshold=3`). 2v2 only hides inputs.

### Referee / scoring panels
| File | Change |
|---|---|
| `web/referee_panel.go` `refereePanelHandler` | passes `NumTeamsPerAlliance` |
| `web/referee_panel.go` `refereePanelFoulListHandler` | passes `NumTeamsPerAlliance`, `TwoVsTwoMode` |
| `templates/referee_panel.html` | `data-two-v-two` on `#refereePanel`; card loops `seq .NumTeamsPerAlliance`; inline `<style>` hiding `#redTeam3Card, #blueTeam3Card, .team-3, .team-3-leave/-muster/-park` |
| `templates/referee_panel_foul_list.html` | `twoVsTwoMode` passed into the `foul` sub-template; third `teamButton` wrapped (commit 01da882) |
| `static/css/referee_panel.css` | `[data-two-v-two="true"]` rules: hide team 3, bigger `.team-card`, 3-column `#scoreSummary` grid (rest of the diff is reformatting/game) |
| `static/js/referee_panel.js` | no 2v2 logic (still writes `.team-3-*`, hidden by CSS) |
| `templates/scoring_panel.html` | `data-two-v-two` on `#scoringPanel`; inline CSS hides `.team-3, #leave-3, #muster-3, #park-3`; loops remain `seq 3` |
| `static/js/scoring_panel.js`, `web/scoring_panel.go` | no 2v2 logic |
| `web/referee_panel_test.go` `TestRefereePanelTwoVsTwoAttribute`, `web/scoring_panel_test.go` `TestScoringPanelTwoVsTwoAttribute` | assert the data attribute |

### Displays
| File | Change |
|---|---|
| `templates/audience_display.html` | realtime overlay: `left/rightTeam3` and `...Team3Avatar` wrapped; final score: team rows changed from `seq 4` to `seq 2` + conditional row 3 (commit dc9e716). Row 4 (off-field team) dropped in BOTH modes. |
| `static/js/audience_display.js`, `static/css/audience_display.css` | no 2v2 logic (jQuery no-ops on missing elements); CSS diff is game/greenscreen polish |
| `web/audience_display_test.go` `TestAudienceDisplayTwoVsTwoTemplate` | |
| `templates/announcer_display_match_load.html` | R3/B3 `team` rows wrapped in `{{if not .TwoVsTwoMode}}` (field comes from the MatchLoad message struct) |
| `web/announcer_display_test.go` `TestAnnouncerDisplayMatchLoadTwoVsTwo` | |
| `web/queueing_display.go` `queueingDisplayMatchLoadHandler` | adds `TwoVsTwoMode` to data |
| `templates/queueing_display_match_load.html` | Red3/Blue3 numbers and avatars wrapped |
| `templates/queueing_display.html` | `<body class="two-vs-two">` (no CSS uses it); `static/css/queueing_display.css` font 42->36px unconditionally |
| `web/field_monitor_display.go` `fieldMonitorDisplayHandler` | adds `TwoVsTwoMode` to data |
| `templates/field_monitor_display.html` | body class `two-vs-two`; rows become `(left1,right2)`, `(left2,right1)` instead of three rows |
| `static/css/field_monitor_display.css` | ~150 lines of `.two-vs-two` sizing (rows 40vh, larger team id/boxes, compact stats) (commit 83b987f) |
| `static/js/field_monitor_display.js` | also adds the class if URL contains `twoVsTwoMode=true` (redundant with the template) |
| rankings, bracket, alliance station, wall, twitch displays | untouched |

### Match review / logs
| File | Change |
|---|---|
| `web/match_review.go` `matchReviewEditGetHandler` | passes `TwoVsTwoMode`, `NumTeamsPerAlliance` |
| `templates/edit_match_result.html` | per-team rows use `seq .NumTeamsPerAlliance`; third foul-team radio wrapped |
| `static/js/match_review.js` | no 2v2 logic (loops `i<3`, missing inputs read as false) |
| `templates/match_review.html`, `templates/match_logs.html` | third team / third log link wrapped (commit 7eee95d) |

### Reports
| File | Change |
|---|---|
| `web/reports.go` `schedulePdfReportHandler` | `teamsPerMatch` 4 for the matches-per-team header; drops "Red 3"/"Blue 3" header, team and surrogate cells; `twov2eol` makes the Blue 2 cell the line-ending cell; surrogate detection ignores slot 3 in 2v2 (necessary because slot 3 is always a surrogate) |
| `templates/schedule.csv`, team/cycle/alliance/bracket reports, `web/api.go` | untouched |

### Tests
Added: 4 arena tests, 4 web template tests (above). NOT added: any test of 2v2 schedule generation, alliance selection, reports, match review. `tournament/schedule_test.go` and `tournament/judging_schedule_test.go` in fms2025/main still call `BuildRandomSchedule` with 3 args while the signature takes 4, so the `tournament` package tests cannot compile there (by reading; not executed).

### What the generic fork has vs fms2025
`origin/main` of generic has only 38f8664 + 83b987f (setting, arena, match play, audience/announcer/referee/scoring attrs, FTA display). `a2a286c` (branch `cleanup/merge-fms-2025-into-generic`, unmerged) backports alliance selection, referee foul list, match review/logs, queueing, reports, match_play.js; it explicitly defers schedule generation (no `2p_` file, 3-arg `BuildRandomSchedule`).

---

## 3. Mapping onto current upstream (`upstream/main`)

Files with NO upstream change since the fork (2v2 hunks should re-apply as-is): `templates/announcer_display_match_load.html`, `templates/queueing_display_match_load.html`, `templates/queueing_display.html`, `templates/field_monitor_display.html`, `templates/match_logs.html`, `templates/referee_panel_foul_list.html`, `web/setup_schedule.go`. Near-trivial drift: `tournament/schedule.go` (+3/-1), `web/queueing_display.go`, `static/js/field_monitor_display.js`.

| Touchpoint | Upstream status | Notes for re-application |
|---|---|---|
| `model/event_settings.go` | exists, heavily extended (2026 game, DMX, SCC, Nexus auto-queue key) | add field; `settingsPostHandler` now rejects changes during a match (good for this setting); settings page is tabbed, pick a tab |
| `field/arena.go` `LoadMatch` | exists; still has the Nexus branch that fms2025 deleted: `SubstituteTeams(lineup[0..5])` | 2v2 block must run on both paths (via SubstituteTeams and the normal path); force lineup[2]/[5] to 0 in 2v2 |
| `field/arena.go` `checkForUpdatedNexusLineup` (NEW, 8153844) + `model.Match.IsLineupEqual` | new | compares all 6 slots then calls SubstituteTeams; zero slots 3/6 before compare or it will re-substitute forever if Nexus returns a third team |
| `field/arena.go` `preLoadNextMatch` | exists; also calls `TeamSigns.SetNextMatchTeams(teamIds)` now taking `[6]int` | zero indices 2 and 5 in `teamIds` (covers Nexus lineup + team signs + network) |
| `checkCanStartMatch` | REFACTORED: now `getStartMatchConditions()` -> `getAllianceStationStartConditions(stations...)` returning human-readable strings (shown as tooltip on Match Play, 6ee45e7); also requires `Plc.IsFtaReady()` | put the station list in one helper (e.g. `arena.activeStations()`), use it in `getStartMatchConditions` AND in `handlePlcInputOutput`'s `redAllianceReady/blueAllianceReady` |
| `handlePlcInputOutput` | exists, same six `handleTeamStop` calls + ethernet | same skip; see aStopReset gap below |
| `ToggleBypass(station)` (now an Arena method, called from match play AND the head-ref panel, e80b7b3) | new location | should refuse R3/B3 in 2v2, otherwise a head ref/scorekeeper can un-bypass an empty station and block match start |
| `field/team_sign.go` | exists upstream (fms2025 deleted it); `Red3/Blue3` signs | sign with id 0 is a no-op, so leaving ids unset works; optional: blank them in 2v2 |
| `field/arena_notifiers.go` `GenerateMatchLoadMessage` | exists | add field; consider adding it to ArenaStatus too |
| `partner/tba.go` `createTbaAlliance` | exists (fms2025 deleted TBA) | already skips `teamId == 0`. With the fms2025 "surrogate team #1"/"duplicate captain" representation TBA would receive bogus/duplicate teams, so upstream 2v2 must store 0 |
| `partner/nexus.go` AutoQueue (b573da1) | new | not station dependent; no change |
| `tournament/schedule.go` | same structure | see section 4 for what to build |
| `tournament/judging_schedule.go` (expanded) | indexes `teamMatches[match.Red3/Blue3]` | harmless with 0 (extra key 0), wrong with surrogate placeholders |
| `web/alliance_selection.go` | same 3 functions, same logic at lines ~119, ~209, ~436-475; bigger file (timer, unpicked) | same edits; also fix Round3 interaction |
| `templates/alliance_selection.html` | header still hard-codes Pick 2 | same edit |
| `web/unpicked_display.go`, `templates/unpicked_display.html`, `static/js/unpicked_display.js` (NEW 1d6558a) | ranked-team list only | no 2v2 work needed |
| `web/field_monitor_display.go` | now also `fmsFieldMonitorDisplayHandler` | |
| `templates/fms_field_monitor_display.html` + `static/js/fms_field_monitor_display.js` + `static/css/fms_field_monitor_display.css` (NEW) | hard-codes 3 `team` blocks per side (`position/station 1..3`); JS iterates `data.AllianceStations` so it is data-driven | NEEDS 2v2: drop position 3 blocks and resize |
| `static/js/display_shared.js` (NEW, shared by audience + wall) | sets `#..Team3`, `#..Team3Avatar` | no-ops if elements are absent; avatar request for team 0 otherwise |
| `templates/wall_display.html`, `static/js/wall_display.js` | `left/rightTeam3` markup | NEEDS the same wrap as audience display (never handled in fms2025) |
| `templates/audience_display.html` | same ids; final rows still `seq 4`; adds winner/tiebreaker indication | wrap Team3 only; keep row 4 |
| `web/referee_panel.go` | same two handlers; "card" playoff branch writes `cards[strconv.Itoa(Red3)]`; new `toggleBypass`, `commitAndPost` messages | guard `Red3/Blue3 == 0` in the playoff card branch |
| `templates/referee_panel.html` | card ids RENAMED to `red3Card`/`blue3Card` with `data-station` (were `redTeam3Card`); cards double as bypass buttons | fms2025 CSS selectors `#redTeam3Card` no longer match; use `seq .NumTeamsPerAlliance` (that already removes them) |
| `static/js/referee_panel.js` | `setTeamCard(...,3,...)`, `.team-3-auto-tower`, `setTeamBypassedStatus("red3")` | hide via CSS or guard |
| `templates/scoring_panel.html` / `static/js/scoring_panel.js` / `web/scoring_panel.go` | rewritten for 2026 (single `red`/`blue` position, manifests renamed); still `seq 3`, `.team-3` | same data-attribute/CSS or `seq N` approach |
| `web/match_review.go` `matchReviewEditGetHandler` | REFACTORED: builds `[]MatchReviewEditAlliance{Teams: []int{Red1,Red2,Red3}}`; template ranges over `.Teams`; JS uses `NUM_ROBOTS` | cleanest: build a 2-element `Teams` slice in 2v2 and make JS tolerate missing index 2; new `matchReviewSummaryPostHandler` live summary needs no change |
| `templates/match_review.html`, `web/match_review.go` list (`RedTeams` 3-slice) | exists | same wrap |
| `web/match_logs.go` (expanded), `templates/match_logs.html`, `templates/view_match_log.html` | R3/B3 switch cases, 3-slices | same wrap |
| `web/reports.go` `schedulePdfReportHandler` | exists; PDFs now go through new `web/report_pdf.go` wrapper (`*reportPdf`, Unicode fix) | same logic, adapt to wrapper type |
| `templates/schedule.csv` | unchanged | emits Red3/Blue3 columns (0 in a clean 2v2) |
| `web/match_play.go` | `substituteTeams` args, `UpdateAllianceFromMatch([3]int{...})` at commit | see gap on team 0 |
| `model/alliance.go` `UpdateAllianceFromMatch` | unchanged upstream | appends any lineup id not in `TeamIds`, including 0 |
| `playoff/double_elimination.go` | now also 4-alliance bracket (741f792) | alliance-size independent |
| `game/score.go` (2026) | `AutoTowerStatuses [3]`, `EndgameTowerStatuses [3]`; auto tower capped at 2 robots; `TraversalBonusThreshold=50`, fuel thresholds | no structural robot-count logic, but thresholds are tuned for 3 robots; all three are already event settings. `RobotsBypassed` no longer exists in `Score` |
| `go.mod` | `go 1.26.0` (fms2025: 1.22) | `rand.Seed` is a no-op from Go 1.24 (GODEBUG `randseednop`); the fms2025 deterministic-shuffle trick will silently stop working. Use a local `rand.New(rand.NewSource(seed))` |

---

## 4. Known gaps / rough edges

**Schedule generation (what exists):** one hand-made template, `schedules/2p_14_0.csv`, for exactly 14 teams, 42 matches (12 per team), truncatable to 28 or 35 matches (8 or 10 per team). `BuildRandomSchedule` picks the file by team count only, takes the first `numMatches` rows, shuffles teams with a fixed seed.

**What is missing / wrong:**
1. Any team count other than 14 fails with "No schedule template exists". No generator exists for `2p_*` files (the memory note "scheduling generator still to be written" matches). Upstream's 3v3 templates come from an external tool; nothing equivalent for 2v2.
2. Truncation is only balanced at the designed cut points (28/35/42 rows). Other block sizes give uneven matches per team; nothing validates this. `matchesPerTeam` is computed but not used to choose the file.
3. Third slots are filled with a real team flagged surrogate instead of 0. Consequences: that team appears as Red3 and Blue3 of every match (sometimes while also playing in slot 1/2, e.g. row 3 `1,0,2,0,1,1,...`); the 12-column parse is preserved only by this trick; schedule CSV report, judging scheduler, TBA publishing, `GetOffFieldTeamIds`, team match logs and any "matches played" style query see phantom appearances; the schedule PDF needed a special case to stop marking every row as "surrogate".
4. `rand.Seed(0)` on the global source: deprecated, no-op on the upstream toolchain, and resets global RNG state for the whole process.
5. `tournament` tests not updated for the 4-arg signature; zero tests for the 2v2 path.
6. Practice vs qualification both use the same fixed seed, so they get the same team mapping.
7. The generic fork has none of this (a2a286c defers it).

**Arena / field:**
8. Stack lights / alliance-ready: `handlePlcInputOutput` still evaluates `checkAllianceStationsReady("R1","R2","R3")`. `assignTeam` sets `aStopReset = !Plc.IsEnabled()` and only `handleTeamStop` sets it back to true; since 2v2 skips `handleTeamStop` for R3/B3, with a PLC enabled R3/B3 never become "ready", so the alliance-ready stack lights would never go green. (Match start itself is fine because it uses the 4-station list.) Not covered by `TestTwoVsTwo_HandlePlcSkipsThirdStations`.
9. Toggling the setting does not touch the currently loaded match; R3/B3 bypass only updates on the next `LoadMatch`/`ResetMatch`.
10. Bypass for R3/B3 can still be toggled via websocket (`toggleBypass`); only the UI row is hidden.
11. The 2v2 block is copy-pasted in `LoadMatch` and `SubstituteTeams`; `setupNetwork` and `preLoadNextMatch` both nil the teams (redundant).

**Alliance selection / playoffs:**
12. `SelectionRound3Order != ""` overrides 2v2 and makes 4-team alliances while `determineNextCell` skips column index 2, so 2v2 + round 3 is incoherent. A "2 + 1 backup" alliance (TeamIds len 3, only 2 on the field) is not supported.
13. Playoff lineup duplicates the captain into slot 3. After any playoff substitution the JS sends 0, the match stores 0, and `UpdateAllianceFromMatch` would set `Lineup[2]=0` and append team 0 to `TeamIds` (by reading `model/alliance.go`; not reproduced).
14. Head-ref playoff cards are written for `Red3/Blue3` too (duplicate captain or "0" key).
15. No 2v2 tests for alliance selection.

**UI:**
16. Audience final score lost the 4th (off-field) team row in both modes (`seq 4` -> `seq 2` + conditional 3): a 3v3 regression.
17. Hiding is done three different ways (template `if`, data attribute + CSS duplicated inline and in the css file, body class); `NumTeamsPerAlliance` computed in three handlers by copy-paste; `queueing_display.html` gets a `two-vs-two` class no CSS uses; queueing font shrunk for everyone; `field_monitor_display.js` URL-param check is dead code.
18. Never handled: wall display, alliance station display for R3/B3, bracket/rankings (fine), team signs, TBA, Nexus, `templates/schedule.csv`, API consumers, and (new upstream) the FMS-style field monitor.

**docs/TwoVTwo.md items not implemented:** `Red3 = Blue3 = 0` with cleared surrogate flags in generated schedules; "bundle templates for common team sizes" (only 14; doc's example name `2p_12_0.csv` does not exist); "generalize pick columns dynamically" (it is an `if`); `.two-vs-two` CSS class on the audience display container (not done; only on field monitor); "2v2 schedule tests" (none); "generalization over scattered `if TwoVsTwoMode`" (the implementation is mostly scattered ifs: about 60 references). The doc does not mention arena/PLC/network handling, alliance selection Round 3/backups, Nexus, TBA, or head-ref bypass at all.

**Recommendation implied by the above (for the playbook):** adopt the doc's representation (third slots = 0, not surrogate) everywhere, one helper for team count/active stations (e.g. `EventSettings.TeamsPerAlliance()` and `arena.activeStations()`), guard `teamId == 0` in `UpdateAllianceFromMatch` and the playoff card branch, and give the 2v2 scheduler its own 8-column (or 12-column with zeros) parse path.

---

## 5. Verification checklist

**Automated (must exist after regeneration):**
- `go build ./... && go test ./...` passes with the setting off, with no golden/expectation changes in stock tests (3v3 unchanged). Specifically `tournament` compiles (all `BuildRandomSchedule` callers updated).
- Settings: POST `/setup/settings` with/without `twoVsTwoMode=on` round-trips; fresh DB defaults to false.
- Arena (port the four fms2025 tests): start with only R1,R2,B1,B2 linked; `LoadMatch`/`SubstituteTeams` leave R3/B3 `Team=nil`, `DsConn=nil`, `Bypass=true`; `ResetMatch` keeps them bypassed; R3/B3 PLC E/A-stops ignored.
- New arena tests: with `FakePlc` enabled, alliance-ready (stack lights green) becomes true in 2v2 with 2 robots per side; `ToggleBypass("R3")` rejected in 2v2; `setupNetwork`/`preLoadNextMatch` pass nil for indices 2 and 5 even if the match record has a nonzero Red3; Nexus lineup with a third team is truncated and `checkForUpdatedNexusLineup` does not loop.
- Schedule: for each shipped `2p_N` template and each supported matches-per-team: every match has 4 distinct nonzero teams, `Red3 == Blue3 == 0`, no surrogate flags on slot 3, each team plays exactly the expected count, deterministic for a fixed seed, error for unsupported team counts; 3v3 path byte-identical behavior (`TestScheduleTeams`, surrogates test unchanged).
- Rankings: `CalculateRankings` on 2v2 matches creates no ranking row for team 0 and counts `Played` correctly.
- Alliance selection: start creates 2-slot alliances; `determineNextCell` walks captain -> pick 1 only; finalize creates alliances with `Lineup = [pick1, captain, 0]` and playoff matches with `Red3 = Blue3 = 0`; defined behavior when `SelectionRound3Order` is set; after a playoff commit `alliance.TeamIds` contains no 0.
- Template tests (port): audience, announcer, referee, scoring render third-slot markup when off and omit it when on; add the same for wall display, FMS field monitor, field monitor (2 rows), queueing, match review list, match logs, edit match result, foul list.
- Reports: schedule PDF/CSV handlers return 200 in both modes; 2v2 PDF has 4 team columns and correct "matches per team" header.
- TBA (if kept): `createTbaAlliance` output has 2 teams per alliance and no surrogates in 2v2.

**Manual, 2v2 on (preview on :8080, 14-team test event):**
1. Generate practice and qualification schedules; inspect `/setup/schedule`, schedule PDF, `/api/matches/qualification`.
2. Match Play: only 2 rows per alliance; load match; start is blocked until 4 stations are linked/bypassed and the tooltip lists only R1/R2/B1/B2; run a full match with bypass; substitute teams in a practice match.
3. Scoring panels and referee/head-ref panel: no third team buttons/cards/foul team button; head-ref bypass works for 1 and 2; commit and post from head ref.
4. Displays: audience (intro, in-match, final score incl. off-field row in playoffs), announcer, queueing, field monitor, FMS field monitor (new), wall, alliance station R3/B3 (blank), rankings, bracket.
5. Match Review: list shows 2 teams; edit and save a result; rankings recompute; match logs list shows 2 links per alliance.
6. Alliance selection end to end (8 alliances x 2), unpicked display, finalize, run a playoff match including a substitution and a timeout, confirm bracket advances and alliances report is correct.
7. With PLC simulated/enabled: stack lights go green with 4 robots ready; R3/B3 E-stops do nothing.
8. Flip the setting off on the same DB copy and repeat 2, 3, 4 quickly: three rows/cards everywhere, six stations required to start, off-field 4th team row present on the final score screen.
