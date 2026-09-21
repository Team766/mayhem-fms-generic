# Game surface inventory: High Seas Havoc in mayhem-fms-2025

Repo `../mayhem-fms-2025` @ `3bc6890` (= `fms2025/main` in `mayhem-fms-generic`), compared with
generic base `83b987f`. Totals: 57 files changed, +1043 / -996 (`git diff --numstat 83b987f HEAD`). Roughly 60% of
that is game content; the rest is 2v2, scheduling, green-screen, reformatting (section C).

Line counts are `+added / -removed` vs `83b987f`. "G" = game content, "N" = non-game, "G+N" = mixed.

## A. Files containing game-specific content

### A1. Go: game package (the model)

| File | +/- | Kind | Game content |
|---|---|---|---|
| `game/mayhem.go` | 21/19 | G | `type Mayhem struct` fields `AutoHullCount, TeleopHullCount, AutoDeckCount, TeleopDeckCount, EndgameKrakenLairCount, LeaveStatuses, MusterStatuses, ParkStatuses`; consts `LeavePoints=4, MusterPoints=6, ParkPoints=3, AutoHullPoints, TeleopHullPoints=2, AutoDeckPoints, TeleopDeckPoints=5, EndgameKrakenLairPoints=10, AutonRankingPointThreshold=20, ScoringRankingPointThreshold=14, EndgameRankingPointThreshold=3, MinorFoulPoints=5, MajorFoulPoints=10` |
| `game/score.go` | 32/56 | G | `Summarize` (all point math, 3 bonus-RP conditions), `Equals` (field-by-field list; must list every Mayhem field) |
| `game/score_summary.go` | 20/18 | G | `ScoreSummary` fields `LeavePoints, MusterPoints, AutoPoints, DeckPoints, TreasureShipPoints, KrakenLairPoints, ParkPoints, MatchPoints, FoulPoints, Score, AutonRankingPoint, ScoringRankingPoint, EndgameRankingPoint, BonusRankingPoints`; `DetermineMatchStatus` playoff tiebreakers (FoulPoints, KrakenLairPoints, DeckPoints, AutoPoints) |
| `game/ranking_fields.go` | 13/13 | G | `RankingFields.EndgameKrakenLairPoints` (3rd tiebreaker), `AddScoreSummary` (win 3 / tie 1), `Less` sort order |
| `game/rule.go` | 25/16 | G | 27 rules; G-rules reworded with CANNONBALL; M2500..M2512; vocabulary KRAKEN LAIR, SAFE HARBOR, HUMAN PLAYER ZONE |
| `game/foul.go` | 0/0 | G (unchanged, inherited) | `PointValue`: G206 minor = 0 special case (upstream 2025 leftover that the game kept) |
| `game/match_timing.go` | 0/0 | G (unchanged) | 0 / 15 / 3 / 135 / warning 20 |
| `game/match_sounds.go` | 0/0 | G (unchanged) | start, end, resume, warning_sonar, end, abort, match_result |
| `game/test_helpers.go` | 24/26 | G | `TestScore1`, `TestScore2` (fixtures used repo-wide), `TestRanking1/2` with `EndgameKrakenLairPoints` |
| `game/score_test.go` | 42/62 | G | expected 144 / 231, `TestAutonRankingPoint`, `TestScoringRankingPoint`, `TestEndgameRankingPoint` |
| `game/score_summary_test.go` | 20/15 | G | tiebreaker order assertions incl. the inverted foul comparison |
| `game/ranking_fields_test.go` | 23/24 | G | `KrakenLairPoints`, `EndgameKrakenLairPoints` |
| `game/rule_test.go` | 0/0 | G (unchanged) | asserts `rules[17]` has id 18 (needs >= 18 contiguous rules) |

### A2. Go: web / field / model / tournament

| File | +/- | Kind | Game content |
|---|---|---|---|
| `web/scoring_panel.go` | 78/76 | G | `ScoringPosition{ScoresAuto, ScoresEndgame, ScoresTreasureShip, ScoresKrakenLair}`; near/far map; websocket commands `leave`, `muster`, `park`, `hull`, `deck`, `kraken_lair` (+ dead `addFoul`) |
| `web/scoring_panel_test.go` | 111/182 | G | drives those commands; asserts Mayhem fields |
| `web/reports.go` | 41/19 | G+N | G (~3 lines): rankings PDF column `EndgameKrakenLairPoints`, `colWidths["Lair"]` (key missing; header text still "Barge"; stale "Coop"/"Barge" widths). N: 2v2 schedule PDF |
| `web/reports_test.go` | 1/1 | G | CSV header `...AutoPoints,EndgameKrakenLairPoints,...` |
| `web/match_play_test.go` | 10/10 | G | `game.Mayhem{TeleopHullCount: 2}`, `TeleopDeckCount`, `LeaveStatuses` |
| `web/match_review_test.go` | 10/10 | G | JSON `"Mayhem":{"TeleopHullCount":10}`, `ParkStatuses`, `LeaveStatuses`, RuleId 1 / 4 |
| `web/audience_display_test.go` | 1/1 | N | greenscreen background default |
| `web/referee_panel.go` | 27/9 | N | 2v2 only (`NumTeamsPerAlliance`, `TwoVsTwoMode`); passes `game.GetAllRules()` (generic) |
| `web/match_review.go` | 16/7 | N | 2v2 only |
| `tournament/qualification_rankings_test.go` | 26/8 | G (indirect) | expected rank orders changed because `TestScore1/2` values changed |
| `field/arena_notifiers.go` | 2/0 | N (+generic G hook) | `GreenScreen`; unchanged `GenerateScorePostedMessage` hard-codes win 3 / tie 1 RP (duplicate of ranking_fields.go) |
| unchanged, fixture consumers | 0/0 | G (indirect) | `model/test_helpers.go` (`TestScore1/2`), `model/match_result_test.go` (`Mayhem.ParkStatuses`), `model/ranking_test.go`, `web/api_test.go` (`TestRanking1/2`) |

### A3. Templates

| File | +/- | Kind | Game content |
|---|---|---|---|
| `templates/scoring_panel.html` | 22/14 | G (+1 line 2v2) | sections "Autonomous", "Hull - Auto", "Deck - Auto", "Hull - Teleop", "Deck - Teleop", "Kraken Lair", "Endgame"; labels "Leave", "Muster", "Park", "Auto Hull", "Auto Deck", "Teleop Hull", "Teleop Deck"; ids `leave-N muster-N park-N auto_hull auto_deck teleop_hull teleop_deck kraken_lair`; `.Position.ScoresTreasureShip/.ScoresKrakenLair`; 2v2 hide-list `#muster-3` |
| `templates/referee_panel.html` | 15/16 | G+N | score summary rows "Leave", "Park", "Muster", "Auto", "Teleop", "Kraken Lair"; classes `muster-symbol team-N-muster auto-hull auto-deck teleop-hull teleop-deck kraken-lair`; buttons "Blue Tech"/"Red Tech"; 2v2 hide-list `.team-3-muster`. N: `seq .NumTeamsPerAlliance` |
| `templates/referee_panel_foul_list.html` | 9/5 | G+N | "Tech" wording (3 places). N: 2v2 team buttons |
| `templates/audience_display.html` | 54/28 | G+N | labels "Ship", "Lair", "Leave", "Muster", "Park", "Foul", "Auton", "Scoring", "Endgame"; ids `leftShip rightShip leftLair rightLair`, `{left,right}Final{Muster,Ship,Lair}Points`, `{left,right}Final{Auton,Scoring,Endgame}RankingPoint`; removed "Ranking Points" row. N: final team rows `seq 2` + conditional team 3 |
| `templates/edit_match_result.html` | 39/73 | G+N | legends "Autonomous/Teleop/Endgame"; "Hull Count", "Deck Count", "Leave Status", "Muster Status", "Park Status", "Kraken Lair Count"; input names `{alliance}AutoHullCount AutoDeckCount TeleopHullCount TeleopDeckCount EndgameKrakenLairCount LeaveStatusesN MusterStatusesN ParkStatusesN`; "Tech" wording in rule select; stale `reefPipe`/`reefBranch` defines. N: `NumTeamsPerAlliance`, `TwoVsTwoMode` |
| `templates/announcer_display_score_posted.html` | 2/2 | G (NOT ported) | only "Tech" wording changed; still uses base-game `Gamepiece1Points`, `Gamepiece2Points`, `LeaveBonusRankingPoint`, `Gamepiece1BonusRankingPoint`, `ParkBonusRankingPoint` -> broken at runtime |
| `templates/rankings_display.html` | 2/8 | G | column "Lair" / `this.EndgameKrakenLairPoints`; dropped Coop, DQ, Played columns |
| `templates/rankings.csv` | 2/2 | G | `EndgameKrakenLairPoints` |
| `templates/setup_settings.html` | 7/0 | N | Green Screen checkbox (game-specific settings were already removed in base) |

### A4. JavaScript

| File | +/- | Kind | Game content |
|---|---|---|---|
| `static/js/scoring_panel.js` | 25/24 | G | `handleCounterClick` ids/commands, `handleMusterClick`, `score.Mayhem.*` reads, `#muster-N`; dead foul code with half-renamed `"tech"` key |
| `static/js/referee_panel.js` | 12/6 | G | `.team-N-muster`, `.auto-hull .auto-deck .teleop-hull .teleop-deck .kraken-lair`, `score.Mayhem.*` |
| `static/js/audience_display.js` | 41/38 | G+N | `redShipCount redLairCount blueShipCount blueLairCount`, `#…Ship #…Lair`, `Final{Muster,Ship,Lair}Points`, `TreasureShipPoints KrakenLairPoints MusterPoints`, `AutonRankingPoint ScoringRankingPoint EndgameRankingPoint`. N/branding: logo sizing `gameplayScale`, `maxLogo`, `logoUp` |
| `static/js/match_review.js` | 15/33 | G | form <-> JSON mapping for all Mayhem fields (base still had REEFSCAPE `Reef`, `BargeAlgae`, `EndgameStatuses`) |
| unchanged, NOT ported | 0/0 | G (stale) | `static/js/announcer_display.js`, `static/js/alliance_station_display.js`, `static/js/wall_display.js` use `ScoreSummary.BargePoints` (NaN), `NumCoral`, `NumAlgae`; `templates/wall_display.html` ids `leftCoral leftAlgae …` |

### A5. CSS, images, audio, docs

| File | +/- | Kind | Game content |
|---|---|---|---|
| `static/css/audience_display.css` | 42/8 | G (branding/layout) | `#logo` 150px, `#blindsLogo` 310px, `.score-value`, final-breakdown row height 38px for 9 rows, `.score-field` padding |
| `static/css/referee_panel.css` | 54/7 | mostly N | reformat (blank lines); G-ish: `.rule-select` 50vw / 1.0vw font for long rule text; stale comment "GP1 and GP2 rows"; 2v2 hide-list has `.team-3-leave .team-3-park` (muster is hidden from the template's inline style instead) |
| `static/img/game-logo.png`, `blinds-logo.png`, `alliance-station-logo.png` | binary | G (branding) | identical 2700x2700 M-Ayhem 2025 kraken badge |
| `static/img/lower-third-logo.png` | binary | G (branding) | 959x720 same artwork |
| `static/img/endofmatch-bg.png` | binary | G (branding) | pink hex tile background |
| `static/audio/*` | 0/0 | unchanged | stock sounds |
| `README.md` | 5/122 | G (light) | "Mechanical M-Ayhem FMS 2025"; no game rules |

## B. Vocabulary list (leftover check)

After replacing High Seas Havoc with another game, none of these should remain (case-insensitive grep over
`*.go *.html *.js *.css *.csv *.md`, excluding `static/lib`, and excluding the false positives listed at the end).

**Go identifiers**
`AutoHullCount`, `TeleopHullCount`, `AutoDeckCount`, `TeleopDeckCount`, `EndgameKrakenLairCount`, `MusterStatuses`,
`MusterPoints`, `AutoHullPoints`, `TeleopHullPoints`, `AutoDeckPoints`, `TeleopDeckPoints`, `EndgameKrakenLairPoints`,
`AutonRankingPointThreshold`, `ScoringRankingPointThreshold`, `EndgameRankingPointThreshold`, `DeckPoints`,
`TreasureShipPoints`, `KrakenLairPoints`, `AutonRankingPoint`, `ScoringRankingPoint`, `EndgameRankingPoint`,
`ScoresTreasureShip`, `ScoresKrakenLair`, `TestAutonRankingPoint`, `TestScoringRankingPoint`, `TestEndgameRankingPoint`

**JSON / websocket field and command names** (same as the Go fields, plus)
`"hull"`, `"deck"`, `"kraken_lair"`, `"muster"` (commands); `Mayhem.AutoHullCount` … in match-result JSON

**HTML ids / form names / CSS classes**
`auto_hull`, `auto_deck`, `teleop_hull`, `teleop_deck`, `kraken_lair`, `muster-1..3`, `handleMusterClick`,
`.muster-symbol`, `.team-N-muster`, `.auto-hull`, `.auto-deck`, `.teleop-hull`, `.teleop-deck`, `.kraken-lair`,
`leftShip`, `rightShip`, `leftLair`, `rightLair`, `FinalMusterPoints`, `FinalShipPoints`, `FinalLairPoints`,
`FinalAutonRankingPoint`, `FinalScoringRankingPoint`, `FinalEndgameRankingPoint`, `redShipCount`, `redLairCount`,
`blueShipCount`, `blueLairCount`, `MusterStatuses1..3` (form names)

**Display strings**
"Hull", "Deck", "Hull - Auto", "Deck - Auto", "Hull - Teleop", "Deck - Teleop", "Auto Hull", "Auto Deck",
"Teleop Hull", "Teleop Deck", "Hull Count", "Deck Count", "Kraken Lair", "Kraken Lair Count", "Lair", "Ship",
"Muster", "Muster Status", "Auton", "Scoring" (as RP label), "Endgame" (as RP label),
rule text: "CANNONBALL", "KRAKEN LAIR", "SAFE HARBOR", "HUMAN PLAYER ZONE", "Human Player Station", rule numbers `M25xx`

**Suggested regex**
`hull|deck|kraken|lair|muster|treasure|cannonball|safe harbor|M25[0-9]{2}|Auton(RankingPoint|&nbsp;)|ScoringRankingPoint|EndgameRankingPoint|\bShip\b`

**Tokens that are NOT leftovers (shared with the generic fork / every game)**
`Mayhem` (struct and `Score.Mayhem` JSON key), `LeaveStatuses`, `ParkStatuses`, `LeavePoints`, `ParkPoints`,
`AutoPoints`, `MatchPoints`, `FoulPoints`, `BonusRankingPoints`, `RobotsBypassed`, `MinorFoulPoints`,
`MajorFoulPoints`, `IsMajor`, `IsRankingPoint`, `ScoresAuto`, `ScoresEndgame`, `leave-N`, `park-N`, "Leave",
"Park", "Tech" (wording choice; decide per game), `red_near/red_far/blue_near/blue_far`. A new game that has no
Leave/Park would have to remove these too, so treat them as "game vocabulary of the placeholder", just not unique
to High Seas Havoc.

**Known false positives for the regex**: `partner/blackmagic.go` and `templates/setup_settings.html` ("HyperDeck"),
`templates/queueing_display_match_load.html` ("On Deck"), `templates/bracket.svg` (base64 font data).

**Absent on purpose**: the strings "High Seas Havoc", "pirate", "havoc" appear nowhere in the 2025 repo; the game
name is not displayed by any screen.

**Older leftovers already present (not High Seas Havoc, but a leftover check will hit them)**: `Gamepiece1Points`,
`Gamepiece2Points`, `LeaveBonusRankingPoint`, `Gamepiece1BonusRankingPoint`, `ParkBonusRankingPoint`
(announcer template); `BargePoints`, `NumCoral`, `NumAlgae`, `Coral`, `Algae`, `Reef*`, "Barge", "Coop"
(`web/reports.go`, `wall_display.*`, `announcer_display.js`, `alliance_station_display.js`,
`edit_match_result.html`, `model/event_settings.go` fields `AutoBonusCoralThreshold`,
`CoralBonusPerLevelThreshold`, `CoralBonusCoopEnabled`, `BargeBonusPointThreshold`, `static/img/coral.png`,
`algae.png`); CSS comment "GP1 and GP2 rows".

## C. 2025 changes that are NOT game content

| Area | Files | Notes |
|---|---|---|
| 2v2 mode follow-ups (setting itself pre-existed in base) | `web/alliance_selection.go` (23/13), `templates/alliance_selection.html`, `web/referee_panel.go`, `web/match_review.go`, `templates/match_review.html`, `templates/match_logs.html`, `templates/queueing_display*.html`, `web/queueing_display.go`, `static/css/queueing_display.css`, `static/js/match_play.js` (Red3/Blue3 = 0), parts of `audience_display.html` (final team rows), `edit_match_result.html` (`NumTeamsPerAlliance`), `referee_panel_foul_list.html` (team buttons), `web/reports.go` schedule PDF, `tournament/qualification_rankings.go` (comment) | alliance selection: 2 teams per alliance, skip pick-2 column; admin-less GET now renders a message |
| 2v2 scheduling | `tournament/schedule.go` (53/9), `schedules/2p_14_0.csv` (new, 42 matches, 14 teams), `web/setup_schedule.go` | loads a single 2-player schedule file; truncates to needed matches |
| Green-screen toggle | `model/event_settings.go`, `web/setup_settings.go`, `templates/setup_settings.html`, `web/audience_display.go`, `web/audience_display_test.go`, `field/arena_notifiers.go` | background `#0f0` vs `#333` |
| UI tweaks / formatting | `static/css/referee_panel.css` (mostly whitespace), import re-ordering in several `web/*.go` | |
| Wording | "Major" -> "Tech" in 5 templates + 1 JS string | borderline: it is game-manual wording, carried in the spec under fouls |
| README | `README.md` | trimmed upstream docs |

## D. Observations useful for the playbook

1. The same fact lives in many places: adding one per-robot status (Muster) touched 11 files (struct, Summarize,
   Equals, summary, ws handler, scoring template + 2v2 hide-list, scoring JS, referee template + hide-list,
   referee JS, edit form, match_review JS x2 directions, audience template + JS, tests).
2. Win/tie RP values are duplicated in `game/ranking_fields.go` and `field/arena_notifiers.go`.
3. `Score.Equals` silently needs every new field; nothing enforces it.
4. Screens the 2025 port missed (so the ground truth is incomplete for them): announcer score-posted, announcer /
   alliance-station / wall realtime score, rankings PDF header and width.
5. `TestScore1/TestScore2` are repo-wide fixtures; changing them shifts expectations in
   `tournament/qualification_rankings_test.go`, `web/*_test.go`, `model/*_test.go`.
