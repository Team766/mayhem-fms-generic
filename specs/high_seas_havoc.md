# Game Spec: High Seas Havoc (Mechanical M-Ayhem 2025)

Status: reverse-engineered from the hand-written 2025 implementation
(`../mayhem-fms-2025` @ `3bc6890`, identical to `fms2025/main` in `mayhem-fms-generic`;
generic base = `83b987f`). This describes the game AS IMPLEMENTED, including quirks. Markers used below:

- `[CODE]` stated directly by the code or its tests.
- `[INFERRED]` my interpretation (field semantics, intent); not verifiable from the code.
- `[QUIRK]` implemented behaviour that looks unintended; a spec author must decide "keep" or "fix".
- `[AMBIGUOUS]` code, comments and prior YAML disagree.

```yaml
spec_version: 1
game:
  id: high_seas_havoc
  name: "High Seas Havoc"          # [INFERRED] the string never appears in the 2025 code; name comes from the
                                   # prior YAML (cheesy-arena cgr/7-settings-review:game/examples/high_seas_havoc.yaml)
  event: "Mechanical M-Ayhem 2025"
  theme: pirates / kraken
  go_struct: Mayhem                # Score.Mayhem; the struct name is generic-fork vocabulary, not game vocabulary
```

## 1. Overview

Pirate-themed game. One game piece, the CANNONBALL. Each alliance scores cannonballs into its TREASURE SHIP,
which has two goals, HULL (low value) and DECK (high value), during auto and teleop; and into its KRAKEN LAIR
during the endgame. Robots additionally earn per-robot points for LEAVE and MUSTER (auto) and PARK (endgame).
Field vocabulary from the rules list: SAFE HARBOR (a per-alliance perimeter, protected in auto), HUMAN PLAYER
ZONE / Human Player Station (cannonballs enter the field through holes in the station), KRAKEN LAIR (protected
for its owner in endgame), midfield line (may not be crossed in auto), one-cannonball control limit, one preload.
`[INFERRED]` physical meaning of Leave / Muster / Park is not in the code: Leave = robot left its starting
area in auto; Muster = robot reached a designated muster location in auto (the prior YAML modelled it as
none/partial(3)/full(6); the 2025 code is a plain yes/no worth 6); Park = robot parked in its endgame area.

## 2. Alliance size and robots

```yaml
alliance:
  robots_per_alliance_data_model: 3        # [CODE] every per-robot array is [3]bool
  event_format_2025: 2v2                   # [INFERRED] from commit history + schedules/2p_14_0.csv (14 teams)
  two_v_two_setting: EventSettings.TwoVsTwoMode   # runtime checkbox on Setup > Settings, default false [CODE]
```

- 2v2 is a generic-fork runtime setting, NOT part of the game. When on: station 3 on each alliance is
  auto-bypassed (`field/arena.go`), `Red3/Blue3 = 0`, and every game screen hides the third robot's
  Leave/Muster/Park control via CSS (`#scoringPanel[data-two-v-two="true"] #muster-3 {display:none}` etc.).
  A new game must add its own per-robot status ids to those hide-lists.
- No scoring rule depends on robot count. Thresholds (20 / 14 / 3) are the same in 2v2 and 3v3. `[CODE]`
- `Score.RobotsBypassed[3]` is recorded from the alliance-station bypass state at match start and compared in
  `Equals`, but is NOT consulted by any 2025 scoring or RP rule (and has no input on the edit form, see 9.7). (The generic base had "all non-bypassed robots left/parked" RPs;
  2025 removed them.) `[CODE]`

## 3. Match timing

```yaml
timing:            # game/match_timing.go — unchanged from generic base / upstream
  warmup_sec: 0
  auto_sec: 15
  pause_sec: 3
  teleop_sec: 135
  warning_remaining_sec: 20      # doubles as "endgame starts" cue [INFERRED]
  timeout_sec: 0
  teleop_grace_period_sec: 3
```

There is no separate endgame period in the FMS state machine. Kraken Lair and Park controls are enabled for the
whole match; "endgame only" is enforced by the referee via rule M2510, not by the FMS. `[CODE]`

## 4. Scoring elements

### 4.1 Counters (alliance-wide, non-negative integers, +/- buttons, clamped at 0)

```yaml
game_pieces:
  - id: cannonball

scoring_counts:
  - id: hull
    display_name: "Hull"
    group: ship
    phases: { auto: 4, teleop: 2 }          # AutoHullPoints = 2 * TeleopHullPoints
    fields: [AutoHullCount, TeleopHullCount]
    entered_by: scoring panel NEAR
    ws_command: hull {Autonomous: bool, Adjustment: int}
  - id: deck
    display_name: "Deck"
    group: ship
    phases: { auto: 10, teleop: 5 }         # AutoDeckPoints = 2 * TeleopDeckPoints
    fields: [AutoDeckCount, TeleopDeckCount]
    entered_by: scoring panel NEAR
    ws_command: deck {Autonomous: bool, Adjustment: int}
  - id: kraken_lair
    display_name: "Kraken Lair"             # short: "Lair"
    group: lair
    phases: { endgame: 10 }
    fields: [EndgameKrakenLairCount]
    entered_by: scoring panel FAR
    ws_command: kraken_lair {Adjustment: int}

scoring_groups:
  - id: ship   # "Ship" / Go: TreasureShipPoints = all hull + all deck points
  - id: deck   # Go: DeckPoints = auto deck + teleop deck points. NESTED inside ship; used only as a playoff tiebreaker
  - id: lair   # "Lair" / Go: KrakenLairPoints
```

Auto vs teleop is chosen by WHICH counter the scorer taps (four separate counters), not by match phase; all
counters stay enabled from auto through post-match until Commit. `[CODE]`

### 4.2 Per-robot statuses (yes/no toggles, one per robot position 1..3)

```yaml
statuses:
  - id: leave   # "Leave"   phase: auto     points: 4   field: LeaveStatuses   entered_by: scoring panel FAR   ws: leave {TeamPosition}
  - id: muster  # "Muster"  phase: auto     points: 6   field: MusterStatuses  entered_by: scoring panel FAR   ws: muster {TeamPosition}
  - id: park    # "Park"    phase: endgame  points: 3   field: ParkStatuses    entered_by: scoring panel FAR   ws: park {TeamPosition}
```

`[AMBIGUOUS]` prior YAML had Muster as multi-value none 0 / partial 3 / full 6. 2025 code: boolean, 6.

### 4.3 Derived totals (`ScoreSummary`)

| Field | Formula |
|---|---|
| LeavePoints | 4 x robots with Leave |
| MusterPoints | 6 x robots with Muster |
| AutoPoints | LeavePoints + MusterPoints + 4*AutoHull + 10*AutoDeck |
| DeckPoints | 10*AutoDeck + 5*TeleopDeck |
| TreasureShipPoints | 4*AutoHull + 2*TeleopHull + DeckPoints |
| KrakenLairPoints | 10 * EndgameKrakenLairCount |
| ParkPoints | 3 x robots with Park |
| MatchPoints | Leave + Muster + TreasureShip + KrakenLair + Park |
| FoulPoints | sum of PointValue() of the OPPONENT's fouls |
| Score | MatchPoints + FoulPoints |
| AutonRankingPoint, ScoringRankingPoint, EndgameRankingPoint, BonusRankingPoints | see section 7 |

If `PlayoffDq` is set the whole summary is zero (score 0, no RPs). `[CODE]`

## 5. Fouls

```yaml
fouls:
  minor: { points: 5,  wording: "Minor" }      # buttons labelled "Red" / "Blue"
  major: { points: 10, wording: "Tech"  }      # buttons labelled "Red Tech" / "Blue Tech"; data field stays IsMajor
  entered_by: referee panel (and Edit Match Result). Team and rule are optional.
  special_cases:
    - rule: G206, when minor -> 0 points        # game/foul.go
```

- Wording: 2025 renamed the user-visible word "Major" to "Tech" in referee panel buttons, the foul-list type
  toggle, the rule dropdown tag, the announcer foul list and the edit-result rule dropdown. NOT renamed: the Go
  field `IsMajor`, constants `MajorFoulPoints`, and the edit-result checkbox label "Is Major?". `[CODE]`
- Rule dropdown shows only rules whose `IsMajor` equals the foul's type; entry format
  `G423 [Tech Foul]: <description>`, or `[Tech Foul + RP]` when `IsRankingPoint`.
- A foul with no rule (RuleId 0) scores normally (5 / 10).
- `[QUIRK]` The scoring-panel websocket still accepts `addFoul {Alliance, IsMajor}` and
  `static/js/scoring_panel.js` still has foul-dialog code, but the 2025 template has no foul buttons; and the JS
  was half-renamed (`"tech"` key vs `localFoulCounts["red-major"]`). Dead code.
- `[QUIRK]` The G206 code comment says it "does make the alliance ineligible for some bonus RPs". Nothing
  implements that.
- `[QUIRK]` "+ RP" rules (M2501, M2502, M2503, `IsRankingPoint: true`) are display-only. No code awards or
  removes a ranking point because of them. `[INFERRED]` intent: M2501/M2502 (endgame interference) should give
  the victim the Endgame RP; M2503 (entering opponent SAFE HARBOR in auto) the Auton RP.

## 6. Rules list

`game/rule.go`: 27 entries `{Id, RuleNumber, IsMajor, IsRankingPoint, Description}`, ids 1..27 contiguous
(`rule_test.go` asserts `rules[17]` is id 18, so contiguity matters).

- 14 FRC-derived general rules: G206, G210, G401, G402, G403, G408, G422, G423, G424, G425, G429, G430, G434, G435
  (several reworded with CANNONBALL).
- 13 game-specific rules M2500..M2512 (M = M-Ayhem, 25 = year).
- Tech (IsMajor) = 10 rules: G210, G408, G423, G424, G430, G435, M2500, M2501, M2502, M2503. Minor = 17.
- IsRankingPoint = 3 rules: M2501, M2502, M2503 (all Tech).
- Do not copy the text into the spec; reference the file. Known typo in source: "perimter" (M2503).

## 7. Ranking points

```yaml
match_result_rp: { win: 3, tie: 1, loss: 0 }        # game/ranking_fields.go; also re-derived in
                                                    # field/arena_notifiers.go GenerateScorePostedMessage
bonus_rp:
  - id: auton     # Go: AutonRankingPoint    display: "Auton"
    condition: AutoPoints >= 20              # POINTS: leave + muster + auto hull + auto deck. AutonRankingPointThreshold
  - id: scoring   # Go: ScoringRankingPoint  display: "Scoring"
    condition: TeleopHullCount + TeleopDeckCount >= 14     # COUNT of cannonballs. ScoringRankingPointThreshold
  - id: endgame   # Go: EndgameRankingPoint  display: "Endgame"
    condition: EndgameKrakenLairCount >= 3                 # COUNT. EndgameRankingPointThreshold
```

- Each bonus RP is worth 1; `BonusRankingPoints` = number achieved (0..3). Max 6 RP per qualification match.
- Bypassed / absent robots: no effect on any condition. No scaling for 2v2. `[CODE]`
- `[QUIRK]` The Scoring RP comment says "during teleop+endgame" but Kraken Lair cannonballs are NOT counted;
  auto cannonballs are not counted either.
- Disqualified team (red card in quals): match counts as Played, Disqualifications+1, zero RP and zero
  tiebreaker points.
- Surrogates do not accrue (generic). In 2v2 the fake third team is treated as absent.

## 8. Ranking order and tiebreakers

```yaml
ranking_sort:            # all compared as per-match AVERAGES (value / Played) via cross-multiplication
  - RankingPoints
  - MatchPoints          # excludes foul points
  - AutoPoints
  - EndgameKrakenLairPoints      # accumulates ScoreSummary.KrakenLairPoints
  - Random               # rand.Float64 stored at last AddScoreSummary

playoff_tiebreakers:     # DetermineMatchStatus(..., applyPlayoffTiebreakers=true), applied only when Score is tied
  - foul_points          # see QUIRK
  - KrakenLairPoints     # higher wins
  - DeckPoints           # higher wins (nested sub-total of ship)
  - AutoPoints           # higher wins
  - else: TieMatch (replayed)
```

`[QUIRK]` First playoff tiebreaker is coded `comparePoints(blue.FoulPoints, red.FoulPoints)`. `FoulPoints` on a
summary is the points that alliance RECEIVED from opponent fouls, so red wins when blue received more, i.e. the
alliance that COMMITTED MORE penalty points wins. FRC convention (and the generic base, which compared
`NumOpponentMajorFouls`) is the opposite. `score_summary_test.go` asserts the as-coded behaviour
(red.FoulPoints=10, blue.FoulPoints=11 -> RedWonMatch). Likely the author read `FoulPoints` as "fouls committed".
Decide before using this spec as a test oracle.

## 9. Screens

### 9.1 Scoring panel (`/panels/scoring/{red,blue}_{near,far}`)

Header: `"<Red|Blue> <Near|Far> - <match long name>"`. Sections are `<h2>` + content, in this order:

| Position | Sections |
|---|---|
| NEAR (`ScoresTreasureShip`) | "Hull - Auto" [counter "Auto Hull"], "Deck - Auto" ["Auto Deck"], "Hull - Teleop" ["Teleop Hull"], "Deck - Teleop" ["Teleop Deck"], Commit |
| FAR (`ScoresAuto`, `ScoresKrakenLair`, `ScoresEndgame`) | "Autonomous" [row of 3 team buttons "Leave"; row of 3 team buttons "Muster"], "Kraken Lair" [counter "Kraken Lair"], "Endgame" [row of 3 team buttons "Park"], Commit |

Team buttons show the team number + label and toggle `data-selected`. Counters are `- value +`. Element ids:
`leave-N`, `muster-N`, `park-N`, `auto_hull`, `auto_deck`, `teleop_hull`, `teleop_deck`, `kraken_lair`. All four
positions must Commit after the match. No foul entry on the scoring panel.

### 9.2 Referee panel (`/panels/referee`)

- Foul buttons: "Blue", "Blue Tech", "Red", "Red Tech". Foul list row: index, type toggle ("Tech Foul" /
  "Minor Foul"), team buttons (2 in 2v2, 3 otherwise), rule `<select>` (50vw wide, 1.0vw font so long
  descriptions fit), Delete.
- Head-ref score summary per alliance (grid, team numbers across the top): rows "Leave", "Park", "Muster"
  (check / cross per robot, in that order), then wide rows "Auto" `hull / deck`, "Teleop" `hull / deck`,
  "Kraken Lair" `count`. CSS classes: `.team-N-leave`, `.team-N-park`, `.team-N-muster`, `.auto-hull`,
  `.auto-deck`, `.teleop-hull`, `.teleop-deck`, `.kraken-lair`.
- Cards, scoring-status lights (Red/Blue Near/Far), Signal Count / Signal Reset / Commit Match: generic.

### 9.3 Audience display

- In-match overlay: two live fields per alliance next to the score: "Ship" = COUNT of cannonballs
  (auto+teleop hull+deck) and "Lair" = Kraken Lair COUNT. Ids `leftShip/rightShip`, `leftLair/rightLair`.
- Final score breakdown rows (centre labels, values left/right, all POINTS): "Leave", "Muster", "Ship"
  (TreasureShipPoints), "Lair" (KrakenLairPoints), "Park", "Foul"; then, hidden in playoffs, check/cross rows
  "Auton", "Scoring", "Endgame". The generic "Ranking Points" total row was removed in 2025. Ids
  `{left,right}Final{Leave,Muster,Ship,Lair,Park,Foul}Points`, `{left,right}Final{Auton,Scoring,Endgame}RankingPoint`.
- Logo is larger than stock: `#logo` 150px, shrinks to 80px during a match (`gameplayScale`, `maxLogo` in JS);
  `#blindsLogo` 310px; row height/line-height tweaks for the 9-row breakdown.

### 9.4 Announcer display

`[QUIRK]` `templates/announcer_display_score_posted.html` was NOT ported: it still references
`.summary.Gamepiece1Points`, `Gamepiece2Points`, `LeaveBonusRankingPoint`, `Gamepiece1BonusRankingPoint`,
`ParkBonusRankingPoint`, which no longer exist, so the Final Results modal fails mid-render after "Auto Leave
Points" (not run; derived from Go template semantics; the existing test only checks for the match name which is
emitted before the failure). Only the Tech/Minor wording was updated. Intended content `[INFERRED]`: Auto Leave,
Muster, Ship, Lair, Park, Foul points; Auton / Scoring / Endgame RP Yes/No; Final Score; Ranking Points; fouls
(type, team, rule number with description tooltip); cards; rankings.
Also unported from the generic base: `static/js/announcer_display.js`, `alliance_station_display.js`,
`wall_display.js` compute `Score - ScoreSummary.BargePoints` (undefined -> NaN), and `wall_display.html/js` still
show Coral/Algae.

### 9.5 Rankings

- Rankings display: Rank, Team, Name, RP, Match, Auto, Lair, W-L-T (2025 dropped Coop, DQ, Played). Values are
  totals, not averages, although sorting is by average.
- `rankings.csv`: `Rank,TeamId,RankingPoints,MatchPoints,AutoPoints,EndgameKrakenLairPoints,Wins,Losses,Ties,Disqualifications,Played`
- Rankings PDF (`web/reports.go`): Rank, Team, RP, Match, Auto, Lair, W-L-T, DQ, Played. `[QUIRK]` the header cell
  text is still "Barge" and the width key `"Lair"` is missing from `colWidths` (which still has "Coop", "Barge"),
  so the Lair column has width 0.

### 9.6 Other reports

No other report has game columns. (Schedule / team / cycle-time reports are generic; 2v2 tweaks only.)

### 9.7 Edit Match Result (`/match_review/{id}/edit`)

Per alliance card: fieldset "Autonomous" [number "Hull Count" `AutoHullCount`, "Deck Count" `AutoDeckCount`;
"Leave Status" checkboxes per team; "Muster Status" checkboxes per team]; fieldset "Teleop" ["Hull Count"
`TeleopHullCount`, "Deck Count" `TeleopDeckCount`]; fieldset "Endgame" ["Park Status" checkboxes; "Kraken Lair
Count" `EndgameKrakenLairCount`]; "Fouls" [Is Major?, Team radio, Rule select with Tech/Minor tag]; "Cards".
Team loops use `NumTeamsPerAlliance` (2 or 3). `[QUIRK]` there are no `RobotsBypassed` inputs although
`match_review.js` reads/writes them (so saving an edit resets them to false); dead `reefPipe` / `reefBranch`
template blocks remain at the bottom of the file.

## 10. Assets

```yaml
logos:       # all replaced in 2025
  game-logo.png, blinds-logo.png, alliance-station-logo.png: same 2700x2700 file; round black badge, red gear
      with "2025", red kraken holding two skull-and-crossbones flags, text "Mechanical M-Ayhem"
  lower-third-logo.png: same artwork, 959x720
  endofmatch-bg.png: 1687x1002 pink / salmon / maroon hexagon-and-square tile pattern (final score background)
colours: no CSS palette change; brand colours are the logo's brick red (~#9b2d22) on near-black, pink final-score bg
sounds:      # all stock; game/match_sounds.go unchanged
  start.wav @0, end.wav @auto end, resume.wav @teleop start, warning_sonar.wav @20 s remaining,
  end.wav @match end, abort.wav, match_result.wav (manual)
leftover_stock_assets: static/img/coral.png, algae.png (unused by 2025 game)
```

## 11. Worked examples (become tests)

### Example A: `TestScore1` (red) vs `TestScore2` (blue); asserted in `game/score_test.go`

| | Red | Blue |
|---|---|---|
| AutoHull / TeleopHull | 2 / 10 | 3 / 5 |
| AutoDeck / TeleopDeck | 3 / 4 | 2 / 10 |
| Kraken Lair | 3 | 4 |
| Leave / Muster / Park | TTT / TTT / TTF | TTT / TTT / TTT |
| RobotsBypassed | F F T | F F F |
| Fouls committed | 1 Tech rule16 (M2501), 2 Minor rule13 (G434), 4 Tech rule15 (M2500) | none |
| LeavePoints / MusterPoints | 12 / 18 | 12 / 18 |
| AutoPoints | 12+18+8+30 = **68** | 12+18+12+20 = **62** |
| DeckPoints | 30+20 = 50 | 20+50 = 70 |
| TreasureShipPoints | 8+20+50 = 78 | 12+10+70 = 92 |
| KrakenLairPoints / ParkPoints | 30 / 6 | 40 / 9 |
| MatchPoints | **144** | **171** |
| FoulPoints (received) | 0 | 10 + 2*5 + 4*10 = **60** |
| Score | **144** | **231** |
| Auton / Scoring / Endgame RP | Y (68) / Y (14) / Y (3) | Y (62) / Y (15) / Y (4) |
| Qualification RP | 0 + 3 = **3** | 3 + 3 = **6** |

Result: Blue wins. (Derived fields not asserted by the existing test: Muster, Deck, TreasureShip points, total RP.)

### Example B: 2v2 qualification match, thresholds at the boundary, G206 and rule-less fouls

| | Red | Blue |
|---|---|---|
| Leave / Muster / Park | TTF / TFF / TFF | TFF / TTF / TTF |
| AutoHull / AutoDeck | 1 / 0 | 0 / 1 |
| TeleopHull / TeleopDeck | 8 / 6 | 3 / 4 |
| Kraken Lair | 2 | 3 |
| RobotsBypassed | F F T | F F T |
| Fouls committed | Minor G206 (rule 1), Minor M2504 (rule 19), Tech no rule (RuleId 0) | none |
| AutoPoints | 8+6+4+0 = 18 | 4+12+0+10 = 26 |
| DeckPoints / TreasureShipPoints | 30 / 4+16+30 = 50 | 10+20 = 30 / 6+30 = 36 |
| KrakenLairPoints / ParkPoints | 20 / 3 | 30 / 6 |
| MatchPoints | 8+6+50+20+3 = **87** | 4+12+36+30+6 = **88** |
| FoulPoints | 0 | 0 + 5 + 10 = **15** |
| Score | **87** | **103** |
| Auton RP | N (18 < 20) | Y (26) |
| Scoring RP | Y (8+6 = 14) | N (7) |
| Endgame RP | N (2) | Y (3) |
| Qualification RP | 0 + 1 = **1** | 3 + 2 = **5** |

Audience live fields: Red Ship 15 / Lair 2; Blue Ship 8 / Lair 3.

### Example C: tied score, qualification vs playoff

| | Red | Blue |
|---|---|---|
| Inputs | AutoDeck 2, TeleopHull 5, Lair 1 | TeleopDeck 4, Lair 2 |
| AutoPoints / DeckPoints / KrakenLairPoints | 20 / 20 / 10 | 0 / 20 / 20 |
| MatchPoints = Score (no fouls) | **40** | **40** |
| Bonus RPs | Auton only | none |

- Qualification: TieMatch. Red 1 + 1 = **2 RP**, Blue 1 + 0 = **1 RP**.
- Playoff: foul points equal (0 = 0) -> KrakenLairPoints 10 < 20 -> **Blue wins**.
- C2 (foul quirk): Red MatchPoints 40 + FoulPoints 5 = 45; Blue MatchPoints 35 + FoulPoints 10 = 45. Playoff
  result AS CODED: **Red wins** (blue received more foul points, i.e. red committed more). Under FRC convention
  Blue would win. See section 8 QUIRK.

## 12. Open questions for the spec author

1. Playoff foul tiebreaker direction (section 8).
2. Should "+ RP" rules and G206 affect ranking points (section 5)?
3. Muster: boolean 6, or none/partial/full (prior YAML)?
4. Should the Scoring RP include Kraken Lair and/or auto cannonballs? (comment vs code)
5. Are RP thresholds meant to differ between 2v2 and 3v3? (code: no)
6. Announcer / wall / alliance-station displays were never ported; what should they show?
