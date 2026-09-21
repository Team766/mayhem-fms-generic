# M-Ayhem generic base: strip list against upstream Cheesy Arena v2026.1.1

Sources compared (all local, read-only):

- `arena_rel/` = upstream Team254/cheesy-arena at v2026.1.1 (`a3b623d`).
- `lite/` = cheesy-arena-lite v2.0.1; its `UPSTREAM.md` checkpoint is exactly `a3b623d`. Diffed file by file with the module path normalised.
- `upstream/main` in `../mayhem-fms-generic` = 5 commits past the tag.
- Old generic fork strip commit `0d265a4` (Sep 2025, against upstream of May 2025, REEFSCAPE era).

Counting note: after normalising the module path I get 92 common files with differences (not 58); the extra ones are mostly `_test.go` files and whitespace-only JS/HTML diffs. 10 upstream files/dirs are absent from Lite (list in the brief, plus `game/hub_test.go`, `game/rule_test.go`, `static/audio/shift_change.wav`). Binary differences: `static/img/game-logo.png`, `static/img/blinds-logo.png`.

Legend for the "Base should" column: **STRIP** = delete, **PLACEHOLDER** = replace the 2026 content with the small placeholder game (file stays, shape stays as close to upstream as possible), **KEEP** = take upstream verbatim, **NEUTRALISE** = keep file, remove only the named identifiers.

---

## 1. Game surface map

### game/

| File | Game-specific content (v2026.1.1) | What Lite did | Base should |
|---|---|---|---|
| `game/score.go` | `Score{AutoTowerStatuses, Hub, EndgameTowerStatuses, Fouls, PlayoffDq}`; vars `EnergizedBonusThreshold`, `SuperchargedBonusThreshold`, `TraversalBonusThreshold`; `TowerStatus` + `TowerNone..TowerLevel3`; all of `Summarize()` (fuel, tower points 15/10/20/30, bonus RPs, G206 check); `Equals()` | Rewrote to a points-only struct `{AutoPoints, TeleopPoints, PostMatchPoints, FoulPointsAgainst, PlayoffDq}`; dropped `Fouls []Foul`; added `matchPoints()`, `winRankingPoints()` | PLACEHOLDER. Keep `Fouls []Foul` and `PlayoffDq` and the foul loop in `Summarize()` (incl. `NumOpponentMajorFouls`); replace the three 2026 fields with the placeholder game's fields. Do NOT copy Lite's `FoulPointsAgainst` int. Remove the three threshold vars and the G206 block (or generalise to `Rule.IsRankingPoint`). |
| `game/score_summary.go` | `ScoreSummary` fields `AutoFuelPoints, AutoTowerPoints, TeleopFuelPoints, TeleopTowerPoints, NumFuel, NumFuelPostMatch, NumFuelGoal, Energized/Supercharged/TraversalBonusRankingPoint, BonusRankingPoints`; playoff tiebreakers in `DetermineMatchStatus()` ("TIEBREAK: AUTO FUEL", "TIEBREAK: TOWER POINTS") | Generic `AutoPoints/TeleopPoints/PostMatchPoints/WinRankingPoints`; removed `NumOpponentMajorFouls` and all tiebreakers except "TRUE TIE" | PLACEHOLDER. Keep `MatchPoints, PostMatchPoints, FoulPoints, Score, PlayoffDq, BonusRankingPoints, NumOpponentMajorFouls` and the "TIEBREAK: MAJOR FOULS" step (fouls are kept). Game-specific tiebreak steps become placeholder steps. `MatchStatus`/`comparePoints` are generic: KEEP. |
| `game/hub.go`, `game/hub_test.go` | Whole file: `Hub`, `Shift` enum (`ShiftAuto, ShiftTransition, Shift1..4, ShiftEndgame, ShiftPostMatch`, `ShiftCount`), `UpdateState`, `GetTeleopActiveFuelCount`, `GetShiftCount`, `GetCurrentShiftTiming`, `GetActiveShiftTiming`, `getScoringGracePeriod` | Deleted | STRIP |
| `game/foul.go` | `Foul{FoulId, IsMajor, TeamId, RuleId}` generic; `PointValue()` has 2026 values (15/5) and the G206 zero-point special case | Deleted | KEEP struct and `Rule()`; PLACEHOLDER the point values (make them named constants the yearly playbook sets) and drop the G206 literal |
| `game/rule.go`, `game/rule_test.go` | `Rule` struct + `GetRuleById`/`GetAllRules` generic; the `rules` slice is the 2026 manual (G206..G4xx, FUEL/HUB/TOWER text; header comment still says "2022 game") | Deleted | KEEP machinery, PLACEHOLDER the `rules` slice (short generic list). `rule_test.go` asserts on specific IDs: rewrite against the placeholder list. Old fork did exactly this (`game/rule.go` 57 lines changed, kept). |
| `game/match_timing.go` | `MatchTiming{TransitionShiftDurationSec, ShiftDurationSec, EndgameDurationSec}`; `GetTeleopDurationSec()` (= transition + 4*shift + endgame); consts `ScoringGracePeriodSec`, `MotorsOnExtraPeriodSec` | Restored pre-2026 shape `{AutoDurationSec, PauseDurationSec, TeleopDurationSec, WarningSoundTimeSec, TimeoutDurationSec}{20,3,140,30,0}`; removed `GetTeleopDurationSec()`; LEFT the two hub consts behind (dead code) | PLACEHOLDER with a flat `TeleopDurationSec` + warning time. Recommend keeping a `GetTeleopDurationSec()` accessor (returns the field) so the seven upstream call sites (`match_sounds.go`, `team_sign.go`, `driver_station_connection.go:308,310`, `arena.go:845`, `static/js/match_timing.js`) stay textually identical to upstream, which makes future ports cheaper. Remove `MotorsOnExtraPeriodSec`; `ScoringGracePeriodSec` only if the placeholder has no PLC-scored element. Upstream's pre-2026 field name was `WarningRemainingDurationSec` (that is what the old fork has); Lite's `WarningSoundTimeSec` is a Lite invention: pick one deliberately. |
| `game/match_sounds.go` | Four `shift_change` entries; warning time computed from `EndgameDurationSec` | Removed shift entries; warning from `WarningSoundTimeSec` clamped at 0 | PLACEHOLDER (same as Lite minus the naming caveat). Delete `static/audio/shift_change.wav`. |
| `game/ranking_fields.go` | `RankingFields{AutoFuelPoints, TowerPoints}`; `GetWinRankingPoints()` (3 RP if traversal bonus enabled, else 2); sort order in `Less()` | `AutoPoints`, `PostMatchPoints`; moved `Random` to end of struct; win RP comes from `ScoreSummary.WinRankingPoints`; added `TeleopPoints()` method | PLACEHOLDER the two tiebreak fields and their use in `AddScoreSummary`/`Less`. Keep upstream's flow (`Score` comparison + `BonusRankingPoints`), win = 2 RP constant. Do not copy Lite's `WinRankingPoints`-in-summary design or field reorder (breaks positional struct literals in upstream tests, e.g. `test_helpers.go`). |
| `game/test_helpers.go` | `TestScore1/2` (Hub, tower statuses, fouls), `TestRanking1/2` positional literals | Rewritten for points-only | PLACEHOLDER, but keep fouls in `TestScore1` (many web tests depend on foul counts) |
| `game/*_test.go` (`score_test`, `score_summary_test`, `ranking_fields_test`, `match_timing_test`, `match_sounds_test`) | 2026 numbers throughout | Rewritten | PLACEHOLDER |

### field/

| File | Game-specific content | What Lite did | Base should |
|---|---|---|---|
| `field/arena.go` | imports `led`, `math`, `math/rand`; `Arena.Leds`, `matchStopTime`, `redWonAuto`; `AllianceStation.GameData`; `LoadSettings()`: `Leds.SetAddress/SetUniverseMode`, `partner.EventEndgameStart` Companion mapping, `MatchTiming.TransitionShift/Shift/Endgame`, three bonus thresholds; `LoadMatch()`: `redWonAuto=false`; `AbortMatch()`: `matchStopTime`; `SetAllianceStationDisplayMode()`: LED logo mode; `Update()`: GameData reset in PreMatch, `handleAutoWinner()` call at Pause->Teleop, `checkEndgameStart()`, `ActiveRemainingSec/ActiveDurationSec` bookkeeping, `updateHubLeds()`; funcs `checkEndgameStart`, `handleAutoWinner`, `getHubLightStates`; `handlePlcInputOutput()`: `GetHubCounts`, `Hub.UpdateState`, hub motor grace logic, `SetHubMotors`, `SetHubLights`, `Leds.SetMode(OffMode)`; `SignalVolunteers/SignalReset`: LED purple/green | Removed all of the above; also removed `NextFoulId` (foul flow); changed `positionPostMatchScoreReady("red") && ("blue")` to `("scoring")`; reflowed several multi-line `append`/`Sprintf` calls onto single lines (cosmetic) | NEUTRALISE: remove every identifier in the left column. KEEP `NextFoulId` (+ reset in `LoadMatch`), KEEP the red/blue `positionPostMatchScoreReady` pair. Do not copy Lite's cosmetic reflows. `checkEndgameStart` + `EventEndgameStart`: see Companion note in section 2. |
| `field/arena_leds.go` | Whole file: `updateHubLeds`, `updateTeleopHubLeds`, const `hubLightWarningSec` (also referenced from `arena.go:getHubLightStates`) | Deleted | STRIP |
| `field/arena_notifiers.go` | `audienceAllianceScoreFields.ActiveRemainingSec/ActiveDurationSec`; `GenerateScorePostedMessage()` uses `game.GetWinRankingPoints()` | Removed those; ALSO removed `RedFouls/BlueFouls/RulesViolated` and `getRulesViolated()`; flattened `generateScoringStatusMessage()` from `PositionStatuses map[string]positionStatus` to `ScoringReady/NumScoringPanels/NumScoringPanelsReady` | NEUTRALISE only the Active*Sec fields and the win-RP call. KEEP fouls/rules in score-posted message and KEEP `PositionStatuses` map (it is the multi-panel seam). |
| `field/realtime_score.go` | `ActiveRemainingSec`, `ActiveDurationSec` | Removed | NEUTRALISE (same as Lite) |
| `field/driver_station_connection.go` | Game-data plumbing: `udpSendPacket [1500]byte`, `SentGameData`, `newDs`, `update(arena, gameData)`, tag-32 block in `encodeControlPacket`, `checkGameData`, `sendGameDataPacketTcp`; `GetTeleopDurationSec()` uses | Removed the entire game-data mechanism (see section 3) | Judgement call. The mechanism is game-agnostic FMS capability; only the payload (`"R"/"B"` from `handleAutoWinner`) is 2026. Recommend KEEP the plumbing with `GameData` always `""` so the file stays byte-close to upstream; strip only `handleAutoWinner`. If the team prefers Lite's removal, accept a permanent merge-conflict hotspot in this file. |
| `field/team_sign.go`, `field/team_sign_test.go` | `teamSignYear = 2026`, `generateTeamSignPeriodText`, `generateTeamSignPeriodPrefix` (Shift letters), fuel/goal rear text in `generateInMatchTeamRearText` | Kept the file, removed the 2026 bits, replaced constant with `currentYearTeamSignText()` | STRIP whole file (team decision). See section 2. |
| `field/fake_plc_test.go` | `redHubCount` etc., `GetHubCounts`, `SetHubMotors`, `SetHubLights` | Removed | NEUTRALISE (must match `plc.Plc` interface) |
| `field/arena_test.go` | Hub/LED/GameData/auto-winner/endgame-Companion tests; `TestLoadTeamsFromNexus` | Removed game tests, kept Nexus test | Remove game tests AND Nexus/team-sign assertions |
| `field/scoring_panel_registry.go`, `field/display.go`, `event_status.go`, `team_match_log.go` | none (positions are free-form strings) | unchanged | KEEP (except Twitch enum, section 2) |

### plc/

| File | Game-specific content | What Lite did | Base should |
|---|---|---|---|
| `plc/plc.go` | Interface methods `GetHubCounts`, `SetHubMotors`, `SetHubLights`; inputs `redHubSensor1..4`, `blueHubSensor1..4`; registers `redHubTotal, blueHubTotal, red/blueHubCount1..4`; coils `redHubMotor, blueHubMotor, redHubLight, blueHubLight` | Removed all | NEUTRALISE, same as Lite. All 2026 enumerants are at the END of each iota block, so removing them does not shift the Modbus addresses of e-stops, a-stops, `redConnected*`, `ftaReady`, stack lights, `fieldResetLight`, `awardsModeLight`. KEEP `GetCycleState` (generic blink helper). |
| `plc/coil_string.go`, `input_string.go`, `register_string.go` | stringer output containing the hub names | Regenerated | Regenerate with `go generate ./plc` (pinned `stringer@v0.43.0`, needs network) or hand-edit; never leave stale |
| `plc/plc_test.go` | hub assertions | Removed | NEUTRALISE |
| `plc/armorblock_string.go` / `armorBlock` enum (`redDs, blueDs, redIoLink, blueIoLink`) | UNSURE. `redIoLink/blueIoLink` predate 2026 (present in the old fork's 2025 base) and Lite keeps them, but they are field-element I/O blocks. `arena.getStartMatchConditions()` refuses to start when any ArmorBlock bit in `fieldIoConnection` is 0 and the PLC is enabled. | unchanged | KEEP for upstream parity, but the playbook should call this out: a PLC program without IO-Link blocks must still set those bits, or the team must drop the two enumerants. |

### led/

| File | Content | Lite | Base |
|---|---|---|---|
| `led/color.go, controller.go, controller_test.go, fixture.go, mode.go, zone.go` | sACN/DMX controller for the 2026 Hub LED fixtures (`Controller`, `Mode` incl. `RedStartupMode`, `*PulseMode`, `*AdvantageMode`, `ModeNames`, fixture layout, universe mode) | Whole package deleted | STRIP whole package. Callers: `field/arena.go`, `field/arena_leds.go`, `web/setup_field_testing.go`, settings `LedControllerAddress`, `LedUniverseMode`. (Upstream main adds `led/mode_test.go`.) |

### model/

| File | Game-specific content | Lite | Base |
|---|---|---|---|
| `model/event_settings.go` | `LedControllerAddress`, `LedUniverseMode` (+ default `"single"`), `CompanionEndgameStartPage/Row/Column`, `TransitionShiftDurationSec`, `ShiftDurationSec`, `EndgameDurationSec`, `EnergizedBonusThreshold`, `SuperchargedBonusThreshold`, `TraversalBonusThreshold` (+ defaults in `GetEventSettings()`) | Removed; added `TeleopDurationSec`, `WarningSoundTimeSec` | NEUTRALISE as Lite, plus integration fields (section 2). BoltDB stores the struct as JSON, so removed fields are silently dropped on next save; no migration needed. |
| `model/event_settings_test.go`, `match_result_test.go` | 2026 values | Rewritten | PLACEHOLDER |
| `model/match_result.go` | none (holds `*game.Score`) | Added `EnsureInitialized()` (not stripping, section 3) | KEEP upstream |
| `model/match.go` | `TbaMatchKey`, `ShouldAllowNexusSubstitution()` | unchanged | see section 2 |

### web/

| File | Game-specific content | Lite | Base |
|---|---|---|---|
| `web/scoring_panel.go` | commands `autoTower`, `endgame` (decode into `game.TowerStatus`); `addFoul` (generic); `positionParameters{"red","blue"}` and `ScoringPosition` | Replaced with single `/panels/scoring` panel, one `score` command carrying six ints (`scoringPanelScoreMessage`); removed `addFoul` and positions | PLACEHOLDER the two game commands only. KEEP `ScoringPosition`/`positionParameters`, the `{position}` routes, `addFoul`, `commitMatch`. |
| `web/referee_panel.go` | none in Go (pure foul flow) | Removed `refereePanelFoulListHandler`, replaced `addFoul/toggleFoulType/updateFoulTeam/updateFoulRule/deleteFoul` with `foulPoints`/`foulPointsAgainst`, split `commitMatch` from `commitAndPost` | KEEP upstream verbatim |
| `web/match_review.go` | none (passes `game.GetAllRules()`) | Removed `Rules`; uses `EnsureInitialized()` | KEEP upstream |
| `web/reports.go` | rankings PDF columns "Auto Fuel", "Tower" (`ranking.AutoFuelPoints`, `TowerPoints`); rankings CSV via template | Auto/Teleop/Endgame columns | PLACEHOLDER column names |
| `web/setup_field_testing.go` | `led` import, `LedModeNames/RedLedMode/BlueLedMode`, 100 ms `ledStatus` streamer goroutine, `setLedMode` command, `fieldTestingLedModeDisabledMessage` | Removed | NEUTRALISE as Lite. KEEP PLC coil override + sounds. |
| `web/setup_settings.go` | form parsing for LED, Companion endgame, shift/endgame durations, three thresholds | Removed/replaced | NEUTRALISE (+ section 2) |
| `web/api.go` | none | ADDED `/api/scores` GET/PUT/PATCH and `RankingWithNickname.TeleopPoints` (section 3) | KEEP upstream |
| `web/*_test.go` (`scoring_panel_test`, `match_play_test`, `match_review_test`, `referee_panel_test`, `setup_field_testing_test`, `setup_settings_test`, `reports_test`, `api_test`, `audience_display_test`, `announcer_display_test`) | 2026 score JSON and expectations | Rewritten | PLACEHOLDER; keep the foul/referee test cases from upstream |

### templates/

| File | Game-specific content | Lite | Base |
|---|---|---|---|
| `audience_display.html` | `.score-fuel` numer/denominator, `#left/rightHubActive` SVG progress ring; final breakdown rows AutoFuel/AutoTower/TeleopFuel/TeleopTower + three bonus RP rows | Emptied `.score-fields`; generic rows Auto/Teleop/Endgame/Match | PLACEHOLDER |
| `wall_display.html` | same fuel + hub-active widgets | removed | PLACEHOLDER |
| `announcer_display_score_posted.html` | four point rows + three bonus RP rows; Fouls list with rules (generic) | generic rows; REMOVED fouls list | PLACEHOLDER rows; KEEP fouls/rules block |
| `scoring_panel.html` | `#tower-controls`, auto/endgame tower button templates; fouls dialog (generic) | replaced by numeric entry for both alliances | PLACEHOLDER tower controls; KEEP fouls dialog, `{{.Position.Title}}` |
| `referee_panel.html` | "Auto Tower"/"Endgame Tower" status rows (`.tower-status team-N-auto-tower` ...) | whole foul UI replaced by foul-points entry | KEEP upstream; PLACEHOLDER only the tower status rows |
| `referee_panel_foul_list.html` | none | Deleted | KEEP |
| `edit_match_result.html` | Hub `WonAuto`, `HubShiftCount0..7`, Auto/Endgame tower radios; fouls editor (generic) | points-only form | PLACEHOLDER game inputs; KEEP fouls editor |
| `match_review.html` | bonus RP symbols (Energized/Supercharged/Traversal) | removed | PLACEHOLDER |
| `rankings_display.html`, `rankings.csv` | "Auto Fuel", "Tower" columns | Auto/Teleop/Endgame | PLACEHOLDER |
| `setup_settings.html` | "Game-Specific" fieldset (transition shift, shift, endgame durations; three thresholds); "LED Lighting" fieldset; Companion "Endgame Start" row | replaced/removed; also removed Publishing tab, moved `tbaEventCode` into Team Info Download | NEUTRALISE (+ section 2) |
| `setup_field_testing.html` | "LED Lighting" fieldset (Blue Hub / Red Hub pixel previews) | removed | NEUTRALISE |
| `base.html` | none. Panel menu: Head Referee, Referee (`?hr=false`), Scoring Red/Blue | Collapsed to Head Referee + single Scoring; tab/space damage in Display menu | KEEP upstream (menu entries follow `positionParameters`) |
| `match_play.html` | none (`#redScoreStatus`, `#blueScoreStatus` badges follow positions) | single `#scoringPanelStatus`; tab damage | KEEP upstream |

### static/js

| File | Game-specific content | Lite | Base |
|---|---|---|---|
| `display_shared.js` | `createHubActiveController` and its helpers (`getActiveProgress*Offset`, `restartHubActiveAnimation`, `scheduleHubActiveReset`, `updateHubActiveIndicator`, `restartPendingHubActiveIndicators`), fuel numerator/denominator in realtime-score handler | removed (190 diff lines) | NEUTRALISE; keep `applyDisplaySides`, `getAvatarUrl`, `handleMatchLoad`, `handleMatchTime` |
| `audience_display.js`, `wall_display.js` | `hubActiveController` wiring; `FinalAutoFuelPoints`...`FinalTraversalBonusRankingPoint` | generic fields | PLACEHOLDER |
| `scoring_panel.js` | `endgameStatusNames`, `handleAutoTowerClick`, `handleEndgameClick`, `.scoring-tower-button`; `addFoul` (generic) | rewritten for numeric entry | PLACEHOLDER tower handlers; KEEP foul handlers |
| `referee_panel.js` | `towerStatusNames`, `setTowerStatus`, six `score.AutoTowerStatuses/EndgameTowerStatuses` lines | rewritten for foul points | KEEP upstream; PLACEHOLDER tower status |
| `match_review.js` | `NUM_HUB_SHIFTS`, bonus RP names, Hub/tower (de)serialisation; fouls (generic) | points-only | PLACEHOLDER; KEEP foul handling |
| `match_timing.js` | `getTeleopDurationSec()` | removed; ALSO renamed state texts (section 3, this is a Lite bug) | PLACEHOLDER the duration helper only |
| `match_play.js` | none (`PositionStatuses` loop) | single-panel status | KEEP upstream |
| `setup_field_testing.js` | `setLedMode`, `updateLedOverrideTooltips`, `handleLedStatus`, `ledContainers`, `modeSelects` | removed | NEUTRALISE |

### static/css, static/img, static/audio, docs

| File | Game-specific | Lite | Base |
|---|---|---|---|
| `css/display_overlay_shared.css` | `.score-fuel*`, `.score-hub-active*`, `.progress-guide/.active-progress/.progress-time`; `.score-fields` width | removed; width 130 -> 180 | PLACEHOLDER |
| `css/wall_display.css` | `.score-aux .score-fuel`, `.score-aux .score-hub-active` | left two duplicate `.score-aux {}` rules (sloppy) | NEUTRALISE cleanly |
| `css/scoring_panel.css`, `css/referee_panel.css` | tower button/status styles | large rewrites | PLACEHOLDER tower rules; KEEP foul styles |
| `css/audience_display.css` | `#blindsLogo` geometry tuned to the 2026 logo | changed to `top:65px;height:150px` for Lite's logo | follow whichever `blinds-logo.png` the base ships |
| `img/game-logo.png`, `img/blinds-logo.png` | REBUILT artwork | replaced with Lite artwork | replace with M-Ayhem/placeholder art (do NOT take Lite's) |
| `audio/shift_change.wav` | 2026 sound | deleted | STRIP |
| `README.md`, `.github/workflows/release.yml` | "2026 FRC game, REBUILT" release text, TBA feature bullet | rebranded "Cheesy Arena Lite" | rewrite for M-Ayhem identity |

### partner/

| File | Game-specific | Lite | Base |
|---|---|---|---|
| `partner/tba.go` | `TbaScoreBreakdown`, `tbaHub`, `createTbaScoringBreakdown()` | removed with all publishing | STRIP whole file (section 2) |
| `partner/companion.go` | `EventEndgameStart` | removed constant | see section 2 (recommend keep, generalised) |
| `partner/blackmagic.go` | none | unchanged | KEEP |

---

## 2. Integration strip list

### 2.1 The Blue Alliance

Files to delete: `partner/tba.go`, `partner/tba_test.go`.

Call sites to remove (upstream line numbers at v2026.1.1):

- `field/arena.go:65` field `TbaClient`; `:225` `partner.NewTbaClient(...)` in `LoadSettings()`.
- `web/match_play.go:492-504` async `PublishMatches`/`PublishRankings` in `commitMatchScore()`.
- `web/alliance_selection.go:240-252` `PublishAlliances` + `PublishMatches` in `allianceSelectionFinalizeHandler`.
- `web/setup_settings.go:115-119` form fields; `:382-497` handlers `settingsPublishAlliancesHandler`, `settingsPublishAwardsHandler`, `settingsPublishMatchesHandler`, `settingsPublishRankingsHandler`, `settingsPublishTeamsHandler`; `settingsTabFromRequest()` case `"publishing"`.
- `web/web.go` five routes `GET /setup/settings/publish_{alliances,awards,matches,rankings,teams}` and `GET /setup/teams/refresh`.
- `web/setup_teams.go:57` `TbaDownloadEnabled` branch in `teamsPostHandler`; `teamsRefreshHandler`; `populateOfficialTeamInfo()` (`:303-350`, uses `GetTeam`, `GetRobotName`, `GetTeamAwards`, `DownloadTeamAvatar`); then-unused imports `bytes`, `regexp`, `time`.
- `web/api.go:228,230` uses `partner.AvatarsDir` (defined at `partner/tba.go:27`). The avatar endpoint and `static/img/avatars/` must stay (displays use them): move the constant, as the old fork did (`const AvatarsDir = "static/img/avatars"` in `web/api.go`).
- `field/driver_station_connection.go:519` sends `EventSettings.TbaEventCode` to the DS as the event-name TCP packet (type 20). This is NOT TBA traffic. If `TbaEventCode` is removed, either rename the setting to a neutral `EventCode` and keep the packet, or drop the block. Recommend rename; test at `driver_station_connection_test.go:230`.
- `model/event_settings.go`: `TbaDownloadEnabled` (default `true` at `:142`), `TbaPublishingEnabled`, `TbaEventCode`, `TbaSecretId`, `TbaSecret`.
- Templates: `setup_settings.html` "Team Info Download" fieldset (`:148-152`), whole "Publishing" tab + nav button (`:465-520`), "Publishing Operations" buttons; `setup_teams.html:21-23, 31-37, 51-54` and the `#loadingFromTba` modal (`:120-130`, generic progress bar: keep, retitle); `setup_awards.html:57-60`; `setup_schedule.html:62-65`.
- Tests: `web/setup_teams_test.go:94`, `web/alliance_selection_test.go:149-150`, `web/setup_settings_test.go:288-295` (publish handler tests), `web/match_play_test.go:111-112`, `model/event_settings_test.go:28`.

KEEP (do not strip): `model.TbaMatchKey` and `Match.TbaMatchKey`. It is set by `tournament/schedule.go:75-95` and every playoff spec (`playoff/double_elimination.go:225`, `single_elimination.go:227-283`, `match_group.go:38,63`, `playoff_tournament.go:159`, `playoff/test_helpers.go:127-129`). Removing it touches the whole playoff package for no benefit; it is inert data.

What Lite did: kept `partner/tba.go` as a download-only client. `GetTeam`, `GetRobotName`, `GetTeamAwards`, `DownloadTeamAvatar`, `getEventName` are live (refactored onto a new `getJson` helper); `PublishTeams/Matches/Rankings/Alliances/Awards` and `DeletePublishedMatches` are stubs returning `nil`; `postRequest`, `TbaMatch`, `TbaScoreBreakdown`, `TbaRanking*`, `TbaPublishedAward`, `createTbaAlliance` are gone. Settings: `TbaPublishingEnabled` forced `false`, `TbaSecretId/TbaSecret` forced `""` in `settingsPostHandler`, but all five struct fields remain; `tbaEventCode` input moved into the Team Info Download fieldset; `TbaDownloadEnabled` still defaults to true. So Lite still talks to TBA for team info. The team wants that gone too.

What the old fork (`0d265a4`) did: deleted `partner/tba.go` + test outright, removed `populateOfficialTeamInfo`/`teamsRefreshHandler`/publish handlers/routes, moved `AvatarsDir`. Left behind: all `Tba*` fields in `model.EventSettings` (still defaulting `TbaDownloadEnabled: true`), partial template cleanup, blank lines where blocks were cut, and goimports-style import regrouping in every touched file (avoid, see guardrails).

Consequence to state in the playbook: with TBA download gone, team name/nickname/location/avatars are manual (CSV-free: only the edit-team form). The old fork already lived with this.

### 2.2 Nexus

Files to delete: `partner/nexus.go`, `partner/nexus_test.go`.

Call sites:

- `field/arena.go:66` field `NexusClient`; `:226` `NewNexusClient(settings.TbaEventCode, settings.NexusAutoQueueKey)`.
- `LoadMatch()` `:340-359` (`loadedByNexus`, `GetLineup`, substitute; keep the `if !loadedByNexus` body unconditionally).
- `StartMatch()` `:542-543` `MatchStarted`; `startTimeout()` `:626-627` `BreakStarted`; `Update()` `:759-760` `BreakEnded`.
- `preLoadNextMatch()` `:1039-1047` lineup fetch only. KEEP the function: it still pre-configures the network for the next match via `arena.setupNetwork(teams, true)`; only the Nexus block and the `TeamSigns.SetNextMatchTeams` line go.
- `checkForUpdatedNexusLineup()` `:1486-1513` and its call in `runPeriodicTasks()` `:1519`.
- `field/arena_notifiers.go:206-208`: `allowManualSubstitution` reduces to `arena.CurrentMatch.ShouldAllowSubstitution()`.
- `model/match.go:133-136` `ShouldAllowNexusSubstitution()`: delete once callers are gone.
- `web/match_play.go:506-511` `AutoQueue` in `commitMatchScore()`.
- `web/setup_settings.go:121-123` form fields; `:513-514` `NexusBaseUrl` template datum.
- `model/event_settings.go`: `NexusEnabled`, `NexusAutoQueueEnabled`, `NexusAutoQueueKey`.
- `templates/setup_settings.html:523-551` Nexus fieldset and `:849-858` `postMessage` listener script.
- Tests: `field/arena_test.go:510-580` `TestLoadTeamsFromNexus`; any AutoQueue assertions in `web/match_play_test.go`.

Lite: keeps Nexus fully functional (lineups and AutoQueue); its only changes are cosmetic (named constants `PostScores/MatchStart/BreakStart/BreakEnd`, `http.StatusOK`). Old fork: deleted `partner/nexus.go` and the (then smaller) call sites, left `NexusEnabled` in settings.

### 2.3 Team signs

Files to delete: `field/team_sign.go`, `field/team_sign_test.go`.

Call sites: `field/arena.go:71` field `TeamSigns`; `:141` `NewTeamSigns()`; `:180-187` eight `SetId` calls in `LoadSettings()`; `:829` `arena.TeamSigns.Update(arena)` in `Update()`; `:1059` `SetNextMatchTeams(teamIds)` in `preLoadNextMatch()`. Settings: `TeamSignRed1Id, Red2Id, Red3Id, RedTimerId, Blue1Id, Blue2Id, Blue3Id, BlueTimerId` in `model/event_settings.go:75-82`, parsed at `web/setup_settings.go:141-148`, rendered at `templates/setup_settings.html:366-425` (whole "Team Signs" fieldset). Comments in `SignalVolunteers()/SignalReset()` mention signs.

Do not confuse with the Alliance Station Display (`web/alliance_station_display.go`, `AllianceStationDisplayMode`): that is the browser display and is KEPT. `SignalVolunteers/SignalReset` still set `AllianceStationDisplayMode` and the PLC field-reset light.

Lite: keeps team signs (de-gamed). Old fork: deleted both files and the arena call sites, left the eight settings fields.

### 2.4 Twitch display

Files to delete: `web/twitch_display.go`, `web/twitch_display_test.go`, `templates/twitch_display.html`, `static/js/twitch_display.js`. Routes `web/web.go:168-169`. Enum `TwitchStreamDisplay` in `field/display.go:38,55,72` (names map and paths map). `DisplayType` values are in-memory only (display registry is keyed by URL, not persisted), so removing a mid-list iota is safe; check `field/display_test.go` and `web/setup_displays_test.go` for the constant. No `base.html` menu entry exists.

Lite: keeps Twitch. Old fork: deleted only the Go handler + test + routes; left the template, JS and the `TwitchStreamDisplay` enum, so Setup > Displays still offered a type that 404s. Do the full removal this time.

### 2.5 Companion (unknown: report + recommendation)

`partner/companion.go` (author Kyle Waremburg, 2025) is a client for Bitfocus Companion, the A/V button-deck automation tool. On arena events it fires a configured button press (page/row/column) over the network: `EventMatchPreview, EventShowOverlay, EventMatchStart, EventTeleopStart, EventEndgameStart, EventMatchEnd, EventShowFinalScore, EventAllianceSelection, EventMatchAbort`. Settings: `CompanionAddress`, `CompanionPort`, plus Page/Row/Column per event; UI is the "Automation" tab. It is inert when `CompanionAddress` is empty, has no external service dependency, and is game-agnostic except `EventEndgameStart`, which upstream derives from `MatchTiming.EndgameDurationSec` in `arena.checkEndgameStart()`.

Recommendation: KEEP (same category as Blackmagic). For the endgame event either (a) drop it like Lite (`EventEndgameStart`, three settings fields, settings row, `checkEndgameStart`), or (b) keep it and trigger at the same instant as the "warning" sound (`WarningSoundTimeSec`/`WarningRemainingDurationSec`). (b) keeps more upstream text intact and is useful for A/V; I would pick (b).

### 2.6 Reverse list: what Lite stripped that the base must KEEP

Take these from UPSTREAM, never from Lite:

- `game/foul.go`, `game/rule.go`, `game/rule_test.go`.
- `Score.Fouls`, `ScoreSummary.NumOpponentMajorFouls`, major-foul playoff tiebreaker.
- `Arena.NextFoulId` and its reset in `LoadMatch()`.
- `web/referee_panel.go` in full: `refereePanelFoulListHandler`, messages `addFoul`, `toggleFoulType`, `updateFoulTeam`, `updateFoulRule`, `deleteFoul`, `card`, `signalVolunteers`, `signalReset`, `commitAndPost`; route `GET /panels/referee/foul_list`; `templates/referee_panel_foul_list.html`; `templates/referee_panel.html`; `static/js/referee_panel.js`; `static/css/referee_panel.css`; the `?hr=false` assistant-referee menu entry.
- `RedFouls/BlueFouls/RulesViolated` + `getRulesViolated()` in `GenerateScorePostedMessage()`; announcer fouls block.
- Fouls editor in `edit_match_result.html` / `match_review.js`; `Rules` datum in `web/match_review.go`.
- Scoring panel `addFoul` command and fouls dialog; per-position scoring panels (`positionParameters`, `/panels/scoring/{position}`), `PositionStatuses` in scoring status, red/blue badges in `match_play.html`/`match_play.js`.

---

## 3. Lite differences that are NOT stripping (do not copy)

1. **Driver station game data removal** (`field/driver_station_connection.go`): Lite removed `udpSendPacket [1500]byte`, `SentGameData`, `newDs` (field and `newDriverStationConnection` parameter), the tag-32 game-data TLV in `encodeControlPacket` (return type changed from `[]byte` to `[22]byte`), `checkGameData`, `sendGameDataPacketTcp` (TCP packet type 28), and `AllianceStation.GameData`. Also renamed `packenLen` -> `packetLen`, `0xFF` -> `0xff`, dropped blank lines and zero-assignments. Port behaviour is NOT different: both trees have `driverStationRoboRioUdpPort = 1121`, `driverStationRoboRioUdpPortLite = 1120` and the `UseLiteUdpPort` setting ("Lite" there means NI's FMS Lite, nothing to do with cheesy-arena-lite).
2. **`/api/scores` endpoint** (`web/api.go`, routes in `web/web.go`): GET/PUT/PATCH of `apiScore{Red,Blue}{auto,teleop,postMatch,foulPointsAgainst}`, with `currentApiScore()`, `applyScorePatch()`. Unauthenticated write access to the realtime score. Lite-only feature.
3. **Points-only scoring model**: `Score.FoulPointsAgainst`, `ScoreSummary.WinRankingPoints`, `winRankingPoints()`, `RankingFields.TeleopPoints()`, `RankingWithNickname.TeleopPoints`, single `"scoring"` registry position, `scoringPanelScoreMessage`, referee `foulPoints`/`foulPointsAgainst`/separate `commitMatch`.
4. **`model.MatchResult.EnsureInitialized()`** and its calls in `RedScoreSummary()`, `BlueScoreSummary()`, `CorrectPlayoffScore()`, `web/match_review.go:normalizeMatchResult`. Defensive refactor, not upstream.
5. **Match-state text renames in `static/js/match_timing.js`**: "PRE-MATCH"->"START", "AUTONOMOUS"->"AUTO", "TELEOPERATED"->"TELEOP", "POST-MATCH"->"POSTMATCH", plus a new `PAUSE_PERIOD` countdown case. This is a latent Lite BUG: `static/js/field_monitor_display.js:168` and `static/js/fms_field_monitor_display.js:38,180,228,301,332,341` still compare against the old strings, so in-match detection, min-battery reset and auto/teleop state on the field monitors never fire in Lite. Never copy.
6. **Naming/identity**: module path `github.com/Team254/cheesy-arena-lite`, binary `cheesy-arena-lite`, release names/zip names in `.github/workflows/release.yml`, README rebrand (and README wrongly drops the `-dev` flag paragraph although `main.go` still has the flag), `AGENTS.md` porting section, `UPSTREAM.md`.
7. **Setting rename** `WarningSoundTimeSec` (upstream historically `WarningRemainingDurationSec`).
8. **`TbaClient.getJson`** helper and TBA stubs; Nexus event-name constants.
9. **Copyright-year churn**: `game/score.go` 2023->2020, `web/field_monitor_display.go` 2026->2018 (Lite kept older headers). Take upstream headers.
10. **Whitespace/format noise**: tabs injected into `templates/base.html` and `templates/match_play.html`, `match_play.js` indentation, trailing-space cleanups in `fms_field_monitor_display.js`, `alliance_selection.js`, `unpicked_display.js`; single-line reflow of `append`/`Sprintf` calls and `isFinals` in `field/arena.go`; grouped imports in `web/scoring_panel.go` (violates its own AGENTS.md import rule).
11. **Artwork/CSS**: Lite's `game-logo.png`, `blinds-logo.png`, `#blindsLogo` geometry, `.score-fields` width 180px.
12. **Dead leftovers**: `ScoringGracePeriodSec`/`MotorsOnExtraPeriodSec` still in Lite's `game/match_timing.go`; duplicate `.score-aux` rules in `wall_display.css`.

Net guidance: use Lite as a checklist of WHERE 2026 content lives (it is reliable for that: every hub/LED/shift/threshold site in section 1 was found by its diff), not as a source of replacement code.

---

## 4. Upstream commits after v2026.1.1 (`a3b623d..upstream/main`)

Tag `v2026.1.1` exists in `mayhem-fms-generic` and resolves to `a3b623d`.

| Commit | Date | Subject | Files | Class |
|---|---|---|---|---|
| `e86332e` | 2026-09-05 | Update playoff timing to match announced official changes for 2027 | `playoff/double_elimination.go` (break specs `{9,360}`->`{9,300}`, `{11,360}`->`{11,600}`), two playoff tests | Game-agnostic: PORT. Policy note: it encodes FIRST's 2027 break lengths; an off-season event may prefer its own, but take it for parity and because tests move with it. |
| `375bb9f` | 2026-09-17 | Send match start time and score post time to TBA (#310) | `partner/tba.go`, `partner/tba_test.go` | Integration-only: SKIP (files are stripped). Verify it touches nothing in `model/` (stat says it does not). |
| `820b7d1` | 2026-09-19 | Hide traversal bonus status on displays when disabled | `field/arena_notifiers.go` (adds `TraversalBonusEnabled` to score-posted struct, re-aligns the whole struct literal), `audience_display.js/.html`, `announcer_display_score_posted.html`, two web tests | Game-specific: SKIP. Caution: it re-indents the `GenerateScorePostedMessage` struct, so later ports touching that struct will conflict textually; resolve by keeping the base's field list. |
| `069a11e` | 2026-09-19 | Increase post-match hub motor clearing time | `game/match_timing.go` (`MotorsOnExtraPeriodSec` 2->7) | Game-specific: SKIP (constant is stripped). |
| `f2f38f4` | 2026-09-19 | Fix red hub testing dropdown labels (#311) | `led/mode.go`, new `led/mode_test.go`, `templates/setup_field_testing.html`, `web/setup_field_testing.go` | Game-specific (LED): SKIP. Make sure the new `led/mode_test.go` is not resurrected by a directory copy. |

Result: 1 of 5 to port. If the base is regenerated from `upstream/main` directly rather than from the tag, the strip list is unchanged except `led/mode_test.go` joins the delete list and `TraversalBonusEnabled` joins the identifiers to remove from `arena_notifiers.go`, the two display templates and `audience_display.js`. Record `f2f38f4` as the checkpoint.

---

## 5. Guardrails for the playbook

Identity and mechanics

- Preserve the base repo's module path, remotes, binary name and README identity; upstream is read from a sibling checkout/ref, never fetched unless asked. Decide the module path once (the old fork kept `github.com/Team254/cheesy-arena`, which makes upstream diffs apply without a sed pass; Lite renamed and pays for it on every port).
- Keep an `UPSTREAM.md` with the last REVIEWED upstream commit (not last ported). Review range = `<checkpoint>..upstream/main`; classify every commit as in section 4 and record skips with a reason.
- Minimise textual distance from upstream. Remove whole lines/blocks; do not reflow, regroup imports, rename variables, fix typos or change copyright headers in files you are only stripping. No goimports (upstream rule: imports alphabetical, ungrouped). The old fork's `0d265a4` shows the cost: import regrouping and stray blank lines in ~15 files turned every later upstream diff into a conflict.
- After any enum change run `go generate ./...` (stringer pinned at `v0.43.0` in `plc/plc.go` for `input`, `register`, `coil`, `armorBlock`, and in `model/match.go` for `MatchType`); never leave `*_string.go` stale. Then `go fmt ./...`, `go build`, `go test ./...`.

Never reintroduce

- Season game: `game/hub.go`, `Hub`, `Shift*`, `TowerStatus`, `AutoTowerStatuses`, `EndgameTowerStatuses`, `WonAuto`, `handleAutoWinner`, `redWonAuto`, `matchStopTime`, `getHubLightStates`, `GetHubCounts/SetHubMotors/SetHubLights`, hub PLC enumerants, `Energized/Supercharged/TraversalBonus*`, `GetWinRankingPoints`, `TransitionShiftDurationSec/ShiftDurationSec/EndgameDurationSec`, `ActiveRemainingSec/ActiveDurationSec`, `createHubActiveController`, `shift_change.wav`, G206 literals, REBUILT artwork/text.
- LED/DMX: package `led/`, `field/arena_leds.go`, `Arena.Leds`, `LedControllerAddress`, `LedUniverseMode`, field-testing LED UI (`setLedMode`, `ledStatus`).
- Integrations: `partner/tba.go`, `partner/nexus.go`, `TbaClient`, `NexusClient`, all `Tba*`/`Nexus*` settings except a neutral event code if retained, publish handlers/routes, `/setup/teams/refresh`, `populateOfficialTeamInfo`, `checkForUpdatedNexusLineup`, `ShouldAllowNexusSubstitution`; `field/team_sign.go`, `TeamSigns`, `TeamSign*Id`; Twitch display (handler, template, JS, routes, `TwitchStreamDisplay`).
- Lite-isms: `/api/scores`, `FoulPointsAgainst`, `WinRankingPoints`, single `"scoring"` panel position, `foulPoints` referee message, `EnsureInitialized`, renamed match-state strings in `match_timing.js`, `cheesy-arena-lite` naming.

Always preserve

- Foul/rule/referee flow exactly as upstream (section 2.6), cards, `PlayoffDq`, `NumOpponentMajorFouls`.
- Per-position scoring panels and `PositionStatuses`; `ScoringPanelRegistry` untouched.
- PLC generic surface and enumerant ORDER: `fieldEStop`, per-station E/A-stops, `redConnected1..blueConnected3`, `ftaReady`, `fieldIoConnection`, `heartbeat`, `matchReset`, stack lights + buzzer, `fieldResetLight`, `awardsModeLight`, `GetCycleState`, ArmorBlock statuses, field-testing coil override. Game I/O for a season goes strictly AFTER these so the generic Modbus map never moves.
- `model.TbaMatchKey` on `Match` and in `playoff/` (inert, deeply threaded).
- Avatars: `static/img/avatars/`, `GET /api/teams/{teamId}/avatar`, relocated `AvatarsDir` constant.
- Companion and Blackmagic clients; all displays (audience, wall, announcer, alliance station, bracket, field monitor, FMS field monitor, logo, queueing, rankings, unpicked, webpage, placeholder); reports; playoffs; alliance selection; match logs; judging/awards/breaks/lower thirds/sponsor slides; network (AP, switch, SCC); `UseLiteUdpPort`.
- DS game-data plumbing (if the recommendation in section 1 is accepted) with an empty payload.

Placeholder-game seam (what the yearly playbook is allowed to touch)

- `game/score.go`, `score_summary.go`, `ranking_fields.go`, `match_timing.go`, `match_sounds.go`, `rule.go` (list only), `foul.go` (point constants only), `test_helpers.go`, their tests.
- The game rows/widgets in: `scoring_panel.{html,js,css}` + `web/scoring_panel.go` commands and `positionParameters`; `referee_panel.html/js` status rows; `audience_display.{html,js}`, `wall_display.{html,js}`, `display_shared.js`, `display_overlay_shared.css`; `announcer_display_score_posted.html`; `edit_match_result.html` + `match_review.{html,js}`; `rankings_display.html`, `rankings.csv`, `web/reports.go` ranking columns; "Game-Specific" fieldset in `setup_settings.html` + matching `EventSettings` fields and `LoadSettings()` lines; PLC game I/O appended after the generic block + `handlePlcInputOutput()` tail; `game-logo.png`, `blinds-logo.png`.
- Anything outside this seam changed by a yearly playbook is a defect; the base-regeneration playbook should be able to assert that with a path allowlist on the diff.

Verification gates worth scripting

- `grep -rniE 'hub|fuel|tower|shift_change|energized|supercharged|traversal|led\.|twitch|nexus|tbaclient|teamsign' --include='*.go' --include='*.html' --include='*.js' --include='*.css'` (excluding `github`, `static/**/lib`, `TbaMatchKey`) returns nothing.
- `go vet ./...` clean (catches orphaned imports such as `math`, `math/rand`, `bytes`, `regexp`, `time` after removals).
- Field monitor pages still see "PRE-MATCH"/"AUTONOMOUS"/"TELEOPERATED"/"POST-MATCH" from `match_timing.js`.
- Settings page round-trip test posts no removed form fields; `settingsTabFromRequest` has no `"publishing"` case.

## Open points / where I am unsure

- ArmorBlock `redIoLink/blueIoLink`: generic or field-element hardware? Affects match-start preconditions when the PLC is enabled (section 1, plc/).
- Whether to keep DS game-data plumbing (recommended) or follow Lite.
- Whether to retain a neutral `EventCode` setting for the DS event-name packet.
- I did not diff every `_test.go` line by line; test rewrites are listed by file, not by case.
