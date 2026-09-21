# Game Spec: Medieval Mayhem (Mechanical M-Ayhem 2026)

Status: written from the team's game manual (plain-text export `medieval.txt`, "Mechanical M-Ayhem 2026 /
Medieval Mayhem", event 2026-11-21). The manual is the source of truth. Anything marked **(inferred)** is my
reading, not the manual's words; every such point that could change a number is also in section 12.
Section 12 must be answered (or its defaults accepted) before code is written.

```yaml
spec_version: 1
game:
  id: medieval_mayhem
  name: "Medieval Mayhem"
  event: "Mechanical M-Ayhem 2026"
  event_date: 2026-11-21
  theme: dragon knights / treasure
  rule_prefix: MA26          # game-specific rules are MA2601..MA2624; FRC rules are Gnnn
```

## 1. Overview

Dragon knights (drivers) pilot dragons (robots) that collect TREASURES (6 inch foam cubes) from a pile in
the middle of the field or from their GOLD MINE (human loading zone) and place them at their own alliance's
SHELVES: on the FLOOR behind the gold scoring line, on the FIRST SHELF, on the TOP SHELF, or STACKED on top
of treasures already scored. Placements made in the 15 second autonomous period are worth more. One
treasure in the centre pile is the DRAGON'S CROWN; it scores double. Robots earn points in auto for leaving
their SAFE HOUSE (the lair, where they start) and for balancing on their MOUNTAIN TOP (a rocking balance
beam). At the end of the match robots score by being balanced on the mountain top or parked in their safe
house. In the last 20 seconds each alliance's human player may throw one specially marked cube into the
TREASURE CHEST (a bucket outside the field): THE TOSS.

Field vocabulary used in rules: SAFE HOUSE, SAFE ZONE (alliance-owned protected areas), HUMAN LOADING ZONE /
HUMAN PLAYER ZONE, scoring line, APRON / TILT / TOP (parts of the balance beam), control limit of 2
treasures per robot, 1 preloaded treasure per robot, 18 treasures + 1 crown in the centre, 14 human-player
treasures per side, 1 toss cube per side (never enters play; worth nothing if placed on a shelf).

Use the word **treasure** for the game piece in code and on screens, **crown** for the dragon's crown,
**floor / first / top / stacked** for the four placements, **balance** for the mountain top, **park** for
the safe house.

## 2. Alliance size and robots

- The event is **2v2** (`EventSettings.TwoVsTwoMode` on). Manual 6.4 allows the organisers to switch to 3v3
  if many teams attend, so the game must work in both sizes.
- No point value, threshold or rule changes with robot count. The Auton RP threshold (20), the Scoring RP
  threshold (12) and "one robot balanced" are the same in 2v2 and 3v3 **(inferred; the manual gives one set
  of numbers and says nothing about 3v3 values)**.
- One human player and one toss cube per alliance in either size, so the Toss is once per alliance.
- Per-robot controls (`leave`, `auto_balance`, `endgame`) are rendered once per active station (2 or 3).
- Bypassed / absent robots: no rule counts robots, so a bypassed robot has no effect on points or on any RP
  condition. Its per-robot statuses simply stay at their zero values. `RobotsBypassed` is recorded as the
  base does and is not used by any rule.
- Did-not-start (manual 6.6, 4.7): a robot that is not in the queueing area by the end of the preceding
  match is bypassed and may not enter late. The manual's sentence on what such a team receives is garbled.
  Spec behaviour **(inferred)**: a bypassed team whose drive team is present receives its alliance's match
  result and RPs like its partner. A NO-SHOW team (no drive-team member in the alliance area at match start,
  manual 6.4.1) is given a red card by the head referee and so gets 0 RP for that match; its partner is
  unaffected. Both cases are stock Cheesy Arena behaviour; no game code is needed.
- There are no surrogates (manual 6.4).

## 3. Match timing

```yaml
timing:
  warmup_sec: 0
  auto_sec: 15
  pause_sec: 3                 # (inferred) manual gives no pause length; stock value
  teleop_sec: 120              # manual "teleop" 1:30 + "endgame" 0:30 are ONE FMS teleop period
  warning_remaining_sec: 30    # endgame starts; stock warning cue moves from 20 s to 30 s remaining
  toss_remaining_sec: 20       # NEW constant: The Toss begins 10 s into endgame
  timeout_sec: 0
  teleop_grace_period_sec: 3   # stock; matches "balanced 3 seconds after the clock hits 0:00"
```

- The base has teleop 135 and warning 20; both change. Defaults live in `game/match_timing.go`; an existing
  event database keeps its old values, so the operator must also set them on Setup > Settings.
- There is no separate endgame state in the FMS. "Endgame only" and "auto only" are judged by people, not
  enforced by the FMS; all scoring controls stay enabled from match start until Commit.
- **New sound cue:** the manual requires "a musical tone" when the Toss begins. Add a match sound
  `toss` (`static/audio/toss.wav`, new asset, short musical fanfare) at
  `auto_sec + pause_sec + teleop_sec - toss_remaining_sec` = 15 + 3 + 120 - 20 = **118 s** into the match.
  Resulting sound table (seconds into match): `start` 0, `end` 15, `resume` 18, `warning_sonar` 108,
  `toss` 118, `end` 138; `abort` and `match_result` manual. The audience display loads sounds from the
  `MatchSounds` list, so adding the entry and the file is enough.
- The balance-beam "balanced" indicator light is field hardware, not FMS **(inferred)**. No PLC I/O.

## 4. Scoring elements

Official scoring is the state of the field at the end of the match (manual 4.9): a treasure that is
knocked off or descored before the end does not count, and the scorer decrements it. A treasure placed in
auto that is still in place at the end keeps its auto value **(inferred)**.

### 4.1 Counters (per alliance, non-negative integers, +/- buttons, clamped at 0)

```yaml
game_pieces:
  - id: treasure

counters:                      # all entered by the NEAR scoring panel
  - { id: auto_floor,     label: "Auto Floor", phase: auto,   points: 4 }
  - { id: auto_first,     label: "Auto First", phase: auto,   points: 8 }
  - { id: auto_top,       label: "Auto Top",   phase: auto,   points: 12 }
  - { id: teleop_floor,   label: "Floor",      phase: teleop, points: 2 }
  - { id: teleop_first,   label: "First",      phase: teleop, points: 5 }
  - { id: teleop_top,     label: "Top",        phase: teleop, points: 10 }
  - { id: teleop_stacked, label: "Stacked",    phase: teleop, points: 8 }
```

- "teleop" counters cover the manual's teleop AND endgame periods. Treasures placed during the endgame
  period score at the **teleop** values; there are no separate endgame placement values (manual 4.1.3
  "Robots may continue to score normally during the endgame").
- Auto vs teleop is decided by which counter the scorer taps, not by the match clock.
- **Floor** = entire treasure behind the scoring line, resting on the floor. **First / Top** = resting on
  that shelf. **Stacked** = resting only on other scored treasures, touching no floor or shelf below it,
  behind the scoring line. A treasure touching both a shelf and a treasure on the level below counts as
  stacked (manual 4.3 note). A stacked treasure is worth 8 whatever level it sits over, including over the
  top shelf (where it is worth less than the 10 of a top-shelf placement).
- **Stacking in auto** earns no stacking value. A treasure stacked during auto is entered in the auto
  counter of the level of the treasure that supports it (floor 4, first 8, top 12) **(inferred from "scored
  as regular points")**. There is no `auto_stacked` counter. It must still be stacked at the end of the
  match to score.
- The toss cube never counts in any counter. A robot's treasures beyond the control limit, treasures added
  by a human player during endgame (MA2613) and treasures thrown onto a shelf from outside the safe house
  (MA2617) are not counted; the scorer leaves them out.
- A robot that started with more than one preloaded treasure (MA2610) cannot earn auto placement points.
  Its auto placements are not entered in the auto counters; if still in place at the end they are entered in
  the teleop counters **(inferred; see 12)**.

### 4.2 The dragon's crown

```yaml
crown:
  id: crown
  type: single-select per alliance
  values: [none, auto_floor, auto_first, auto_top, teleop_floor, teleop_first, teleop_top, teleop_stacked]
  default: none
  entered_by: NEAR scoring panel
  bonus_points: { none: 0, auto_floor: 4, auto_first: 8, auto_top: 12,
                  teleop_floor: 2, teleop_first: 5, teleop_top: 10, teleop_stacked: 8 }
```

- There is ONE crown on the field (19 centre treasures = 18 + the crown), so at most one alliance can score
  it. It is not per alliance. Each alliance's score still carries its own `crown` field; the FMS does not
  stop both alliances from setting it (the head referee fixes that in review).
- What is doubled: only the crown's own placement value, in the phase and at the location where the crown
  ends the match. Nothing else is doubled (not the alliance score, not other treasures, not robot points)
  **(inferred from "worth x2 points"; see 12)**.
- How it is recorded: the scorer counts the crown in the normal counter for its location like any other
  treasure, AND sets `crown` to that same location. The counter gives the base value; `crown` adds a bonus
  equal to the base value again. So a crown on the top shelf in teleop scores 10 (counter) + 10 (bonus) =
  20; on the first shelf in auto 8 + 8 = 16; stacked in teleop 8 + 8 = 16. The counters therefore always
  equal the number of cubes physically scored.
- The crown may be stacked (16 in teleop) **(inferred; nothing forbids it)**. A crown stacked in auto is
  recorded as `auto_<level beneath>` like any auto-stacked treasure. A treasure stacked on top of the crown
  is an ordinary stacked treasure (8, not doubled).
- No validation: the bonus is added whenever `crown != none`, even if the matching counter is 0.
- The crown cannot be used for the Toss; the Toss uses its own marked cube.
- The crown is ONE treasure: it counts once toward the Scoring RP (through its counter), never twice.

### 4.3 Per-robot statuses (one per active robot)

```yaml
statuses:
  - id: leave
    label: "Leave"
    phase: auto
    type: bool
    points: 4                  # robot left its safe house during auto
    entered_by: FAR scoring panel
  - id: auto_balance
    label: "Auto Balance"
    phase: auto
    type: bool
    points: 12                 # robot stopped and balanced on its mountain top at the end of auto
    entered_by: FAR scoring panel
  - id: endgame
    label: "Endgame"
    phase: endgame
    type: enum
    values:
      - { id: none,    label: "None",    points: 0 }
      - { id: park,    label: "Park",    points: 2 }    # parked in own safe house at the end of the match
      - { id: balance, label: "Balance", points: 12 }   # balanced on own mountain top at the end of the match
    entered_by: FAR scoring panel
```

- `leave` and `auto_balance` are independent toggles and stack: a robot with both earns 4 + 12 = 16. The
  FMS does not set `leave` automatically when `auto_balance` is set.
- Both auto statuses and the endgame status are per robot: two balanced robots earn 24 **(inferred from
  "Robots that end the auton period balanced ... will receive 12" and "Park a robot ... for 12")**.
- `endgame` is one three-way choice per robot: park and balance are mutually exclusive, because a balanced
  robot must rest entirely on the beam **(inferred)**.
- Balance scores only if the robot rests entirely on the beam with nothing on the ground, is not moving at
  the end of play, and is still balanced 3 seconds after 0:00 (judged by people).

### 4.4 The Toss

```yaml
toss:
  id: toss
  label: "Toss"
  type: bool per alliance      # once per alliance, not per robot
  phase: endgame
  points: 2
  entered_by: FAR scoring panel
```

Set when ALL of these hold (judged by people): the alliance's human player threw the alliance's marked toss
cube by hand, after the Toss cue (20 s remaining) and before the end of the match, with one foot in the
Human Player Zone (otherwise MA2622: no score, no penalty); the cube is in the Treasure Chest and not moving
at the end of the match; the chest is vertical at the end of the match. One cube only; maximum 2 points per
alliance per match. The Toss cube is not a shelf treasure and counts toward nothing else.

### 4.5 Derived totals (`ScoreSummary`)

| id | Formula |
|---|---|
| `leave_points` | 4 x robots with `leave` |
| `auto_balance_points` | 12 x robots with `auto_balance` |
| `crown_bonus_points` | `bonus_points[crown]` (section 4.2) |
| `auto_treasure_points` | 4*auto_floor + 8*auto_first + 12*auto_top + (crown bonus if `crown` is an `auto_*` value) |
| `auton_points` | leave_points + auto_balance_points + auto_treasure_points |
| `teleop_treasure_points` | 2*teleop_floor + 5*teleop_first + 10*teleop_top + 8*teleop_stacked + (crown bonus if `crown` is a `teleop_*` value) |
| `endgame_points` | 2 x robots with `endgame = park` + 12 x robots with `endgame = balance` |
| `toss_points` | 2 if `toss` else 0 |
| `match_points` | auton_points + teleop_treasure_points + endgame_points + toss_points |
| `foul_points` | sum of point values of the OPPONENT's fouls |
| `score` | match_points + foul_points |
| `treasure_count` | sum of all seven counters (live display) |
| `shelf_treasure_count` | teleop_first + teleop_top + teleop_stacked (Scoring RP, live display) |
| `num_opponent_major_fouls` | number of major fouls the OPPONENT committed (playoff tiebreak) |
| `auton_rp`, `scoring_rp`, `endgame_rp`, `bonus_rp` | section 7 |
| `auton_rp_by_foul`, `endgame_rp_by_foul` | true when that RP's opponent-violation alternative is satisfied (display only) |

`crown_bonus_points` is informational; it is already inside `auto_treasure_points` or
`teleop_treasure_points` and must not be added to `match_points` a second time. If `PlayoffDq` is set the
whole summary is zero (stock).

## 5. Fouls and cards

```yaml
fouls:
  minor: { points: 5,  wording: "Minor" }
  major: { points: 10, wording: "Major" }     # NOT "Tech": rename the placeholder game's user-visible "Tech" back to "Major"
  awarded_to: the opposing alliance (added to its score)
  entered_by: referee panel (and Edit Match Result). Team and rule are optional.
  special_cases: none                         # remove the placeholder's "G206 minor = 0 points" special case
```

- A foul with no rule selected scores normally (5 / 10) and never affects a ranking point.
- **Fouls that change ranking points.** Three major fouls are flagged `rp: true`:
  - **MA2601** (contacting the opposing alliance's balance beam during endgame) and **MA2602** (contacting
    an opposing robot at all while it is on its own balance beam during endgame): the OTHER alliance gets
    the Endgame RP "no matter the scoring".
  - **MA2603** (entering the opposing alliance's safe house or safe zone during auto): the OTHER alliance
    gets the Auton RP "no matter the scoring".
  - How the FMS knows: the referee enters a foul against the offending alliance and selects that rule in
    the foul's rule dropdown. The RP condition reads the opponent's foul list for a foul whose rule number
    is MA2601 / MA2602 / MA2603. The foul also scores its normal 10 points. A foul entered without the rule
    gives points but no RP; the head referee can add the rule later on the referee panel or in Edit Match
    Result and the RP recomputes. MA2601 and MA2602 may both be assessed for one incident (20 points); the
    Endgame RP is still awarded once. Each RP is 0 or 1; earning it both ways does not give 2.
- Repeating fouls (MA2604 every 10 s, MA2606 every 10 s) are entered by the referee as separate fouls.
- "No-score" violation MA2622 is not a foul: the scorer just leaves `toss` off.
- **Cards** (stock FMS behaviour, no game code): yellow card = warning, no points; a second yellow card
  becomes a red card; yellow cards are cleared when alliance selection begins. Red card in a qualification
  match: that team is disqualified from the match (0 RP, nothing added to its tiebreaker sums, match counts
  as played). Red card in a playoff match: the whole alliance scores 0 for that match.
  - Red-card violations in the manual: egregious or continued unsafe behaviour; G102 robot leaving the
    field; initiating disabling contact; egregious damage to the field; G101/G209 battery contacting the
    field; G246 colluding to shut down match play; intentionally sabotaging the match (given after the
    match). Any yellow-card behaviour may be escalated to red.
  - Yellow-card violations: G101 humans on or reaching over the field; initiating disabling contact; G423
    damaging or impairing an opponent robot; G206 colluding to violate rules; MA2605 crossing barricades
    with the robot; MA2615 ignoring referee-area signs; MA2618 taking or dislodging a piece inside another
    robot's bounding box; MA2619 purposely pushing a robot into a human loading zone, safe house or safe
    zone; MA2620 contacting the opposing shelf; MA2621 a thrown toss cube contacting a treasure, shelf,
    robot or human (head referee may add a foul); MA2623 no safety glasses; MA2624 not behind the barriers.
  - G423 and MA2615 may instead be assessed as a minor foul, a major foul or a red card (referee's choice).
- Scorer A-stops and E-stops (manual 4.6) are not fouls or cards and have no scoring effect.

## 6. Rules list

`game/rule.go`, ids contiguous from 1. `major: false` means minor. Texts are shortened from the manual.
Rules that can be assessed at either severity (G423, MA2615) appear twice because the rule dropdown filters
by the foul's type. Dropdown entry format: `MA2603 [Major + RP]: <text>`, `G425 [Minor]: <text>`.

```yaml
rules:
  - { id: 1,  rule: G210,   major: true,  rp: false, text: "Do not force an opponent robot to commit a foul." }
  - { id: 2,  rule: G401,   major: true,  rp: false, text: "Drive team members must stay behind the lines during auto." }
  - { id: 3,  rule: G402,   major: true,  rp: false, text: "Drive team members must not control their robot during auto, except to press A-stop or E-stop." }
  - { id: 4,  rule: G403,   major: false, rp: false, text: "A robot may not cross the midfield line into the opponent's half during auto." }
  - { id: 5,  rule: G408,   major: true,  rp: false, text: "No human or human player may damage a treasure." }
  - { id: 6,  rule: G415,   major: true,  rp: false, text: "Contacting another robot inside its perimeter. Can stack with MA2618." }
  - { id: 7,  rule: G423,   major: false, rp: false, text: "Damaging or impairing an opponent robot (referee's choice of severity)." }
  - { id: 8,  rule: G423,   major: true,  rp: false, text: "Damaging or impairing an opponent robot (referee's choice of severity)." }
  - { id: 9,  rule: G424,   major: true,  rp: false, text: "Intentionally attaching to, entangling or tipping another robot." }
  - { id: 10, rule: G425,   major: false, rp: false, text: "Pinning a robot for longer than 3 seconds." }
  - { id: 11, rule: G429,   major: true,  rp: false, text: "Drive team members must stay in their designated areas." }
  - { id: 12, rule: G430,   major: true,  rp: false, text: "A robot may be operated only by its drivers." }
  - { id: 13, rule: G434,   major: true,  rp: false, text: "Coaches may not touch treasures, unless for safety." }
  - { id: 14, rule: MA2601, major: true,  rp: true,  text: "Contacting the opposing alliance's balance beam during endgame. Opponent gets the Endgame RP. Can stack with MA2602." }
  - { id: 15, rule: MA2602, major: true,  rp: true,  text: "Contacting an opposing robot at all while it is on its own balance beam during endgame. Opponent gets the Endgame RP. Can stack with MA2601." }
  - { id: 16, rule: MA2603, major: true,  rp: true,  text: "Entering the opposing alliance's safe house or safe zone during auto. Opponent gets the Auton RP." }
  - { id: 17, rule: MA2604, major: true,  rp: false, text: "Hoarding: more than 6 non-scoring treasures in your safe house or safe zone. Repeats every 10 seconds." }
  - { id: 18, rule: MA2606, major: false, rp: false, text: "Entering the opposing human loading zone, safe house or safe zone, or touching their balance beam before endgame. Repeats every 10 seconds. The apron facing midfield is allowed." }
  - { id: 19, rule: MA2607, major: false, rp: false, text: "Putting a treasure into the field other than through the human loading holes (not during the Toss)." }
  - { id: 20, rule: MA2608, major: false, rp: false, text: "Loaded treasure does not first touch an own-alliance robot or the human loading zone floor." }
  - { id: 21, rule: MA2609, major: false, rp: false, text: "Contacting own shelf repeatedly or in an unsafe manner." }
  - { id: 22, rule: MA2610, major: false, rp: false, text: "Starting auto with more than 1 treasure. That robot cannot earn auto placement points." }
  - { id: 23, rule: MA2611, major: false, rp: false, text: "Descoring the other alliance's treasures. Can stack with MA2620." }
  - { id: 24, rule: MA2612, major: false, rp: false, text: "Robot deliberately destroying a treasure." }
  - { id: 25, rule: MA2613, major: false, rp: false, text: "Adding treasures to the field during endgame. The treasure does not count." }
  - { id: 26, rule: MA2614, major: false, rp: false, text: "Placing or shooting treasures into the opposing alliance's zones." }
  - { id: 27, rule: MA2615, major: false, rp: false, text: "Failure to obey referee-area signs (referee's choice of severity)." }
  - { id: 28, rule: MA2615, major: true,  rp: false, text: "Failure to obey referee-area signs (referee's choice of severity)." }
  - { id: 29, rule: MA2616, major: false, rp: false, text: "Controlling more than 2 treasures for more than a moment. Extra treasures do not count." }
  - { id: 30, rule: MA2617, major: false, rp: false, text: "Throwing a treasure onto the shelf from outside own safe house. The treasure does not count." }
```

Totals: 30 entries; 15 major, 15 minor; 3 with `rp: true` (ids 14, 15, 16). Card-only rules (G101, G102,
G206, G209, G246, MA2605, MA2618..MA2621, MA2623, MA2624) and the no-score rule MA2622 are not in the list
because they carry no foul points. The manual also names fouls with no rule number (initiating damaging
contact: major; reaching onto the field: major; throwing the toss cube early: "a foul"; throwing unsafely:
minor); the referee enters these as rule-less fouls unless the designers number them (section 12).

## 7. Ranking points

```yaml
match_result_rp: { win: 3, tie: 1, loss: 0 }
bonus_rp:                      # each worth 1; max 6 RP per qualification match
  - { id: auton,   label: "Auton RP",   threshold: 20, tunable: true, setting: AutonRpThreshold }
  - { id: scoring, label: "Scoring RP", threshold: 12, tunable: true, setting: ScoringRpThreshold }
  - { id: endgame, label: "Endgame RP" }
```

- **Auton RP.** The alliance earns 1 RP if its `auton_points` is **at least 20** (`>=`), OR the opposing
  alliance has at least one foul with rule MA2603. `auton_points` counts leave, auto balance, auto
  placements and the crown bonus when the crown was placed in auto. It does not count foul points or
  anything from teleop. Note every auto value is a multiple of 4, so 19 is unreachable; the boundary cases
  are 16 (no) and 20 (yes).
- **Scoring RP.** The alliance earns 1 RP if `shelf_treasure_count` = `teleop_first + teleop_top +
  teleop_stacked` is **at least 12** (`>=`). This counts treasures placed during the teleop and endgame
  periods that are on the first shelf, on the top shelf, or stacked (stacked on any level) at the end of the
  match. It does NOT count: floor treasures (`teleop_floor`), any auto placement (`auto_floor`,
  `auto_first`, `auto_top`, including treasures stacked during auto), or the Toss cube. The crown counts
  ONCE (it is one treasure in one counter); the `crown` selector adds points only, never a count. There is
  no opponent-violation alternative for this RP.
- **Endgame RP.** The alliance earns 1 RP if **at least one** of its robots has `endgame = balance`, OR the
  opposing alliance has at least one foul with rule MA2601 or MA2602. `park` does not count. `auto_balance`
  does not count.
- `bonus_rp` = number of the three achieved (0..3). Qualification RP = match result RP + `bonus_rp`.
- Bypassed or absent robots have no effect on any condition; thresholds are not scaled by alliance size.
- A team disqualified from a qualification match (red card, including a no-show) gets 0 RP for it.
- Playoff matches award no RPs; RP rows are hidden in playoffs.

## 8. Ranking order and tiebreakers

```yaml
ranking_sort:          # each compared as a per-match AVERAGE (total / Played), higher first
  - RankingPoints      # manual "Ranking Score" = RP / matches
  - Score              # manual "average match score": the alliance's final score INCLUDING foul points (see 12)
  - AutonPoints        # manual "average auton score": auton_points
  - Random             # manual "coin flip" / "arbitrary FMS determination (no appeals)"
ranking_fields: [RankingPoints, ScorePoints, AutonPoints, Random, Wins, Losses, Ties, Disqualifications, Played]
```

- The manual rounds the Ranking Score to 2 decimals. Keep the base's exact cross-multiplied comparison; the
  2-decimal value is for display only **(inferred)**.
- Averaging per match played also covers manual 6.4.2 (a shortened schedule is pro-rated per game).
- A disqualified team's match adds 1 to Played and Disqualifications and 0 to everything else (stock).

**Playoffs.** The manual defines NO tiebreak procedure for a tied playoff match. It says only that teams
"advance based on winning, losing, or tying a match" (6.5) and that "there are no match replays" (6.2,
written about disputes). Until the designers answer (section 12), implement this proposed order
**(inferred: first level and third level from the team's earlier FMS attempt, second level from manual
4.1.1 "auton points are ... tallied separately for tiebreaking")**:

```yaml
playoff_tiebreakers:   # applied only when score is tied
  - num_opponent_major_fouls   # the alliance that COMMITTED FEWER major fouls wins     "TIEBREAK: MAJOR FOULS"
  - auton_points               # MORE wins                                              "TIEBREAK: AUTON POINTS"
  - match_points               # MORE wins (score excluding foul points)                "TIEBREAK: MATCH POINTS"
  - else: TieMatch             # the base bracket replays the match
```

Playoff format (manual: 4 alliances of 2, single elimination, final best of three if time allows) is an
event setting and base behaviour, not game code. The base's single-elimination bracket is best of three in
every round; single-game semifinals would need a base change (outside the game seam).

## 9. Screens

Labels are final; keep them this short. "Per robot" means one control per active station (2 or 3), labelled
with the team number.

### 9.1 Scoring panels (`/panels/scoring/{red,blue}_{near,far}`)

Header: `<Red|Blue> <Near|Far> - <match name>`. All controls enabled from match start until Commit. All four
positions must commit. No foul entry on scoring panels.

**NEAR = shelf scorer.**

| Section heading | Controls, in order |
|---|---|
| "Auto" | counter `auto_floor` "Auto Floor", counter `auto_first` "Auto First", counter `auto_top` "Auto Top" |
| "Teleop + Endgame" | counter `teleop_floor` "Floor", counter `teleop_first` "First", counter `teleop_top` "Top", counter `teleop_stacked` "Stacked" |
| (on each counter row) | a "Crown" toggle button to the right of the `- value +` control. The seven crown buttons form one single-select group bound to `crown`; tapping the selected one again sets `crown = none`. Element ids `crown-auto_floor` ... `crown-teleop_stacked` |
| | Commit |

**FAR = robot scorer.**

| Section heading | Controls, in order |
|---|---|
| "Auto" | per robot toggle "Leave" (`leave-N`); per robot toggle "Auto Balance" (`auto_balance-N`) |
| "Endgame" | per robot three-way button group "None" / "Park" / "Balance" (`endgame-N`) |
| "The Toss" | one toggle "Toss" (`toss`) |
| | Commit |

Websocket commands **(inferred naming)**: `treasure {Counter: string, Adjustment: int}`, `crown {Value:
string}`, `leave {TeamPosition}`, `auto_balance {TeamPosition}`, `endgame {TeamPosition, Value}`, `toss {}`.

### 9.2 Referee panel (`/panels/referee`)

- Foul buttons "Blue Minor", "Blue Major", "Red Minor", "Red Major". Foul list row: index, type toggle
  ("Major Foul" / "Minor Foul"), team buttons (active teams only), rule select filtered by type with the
  `[Minor]` / `[Major]` / `[Major + RP]` tag, Delete.
- Head-referee score summary per alliance, team numbers across the top:
  per-robot rows "Leave" (check/cross), "Auto Bal" (check/cross), "Endgame" (text None / Park / Balance);
  wide rows "Auto" `floor / first / top`, "Teleop" `floor / first / top / stacked`, "Crown" (location text
  such as "Teleop Top", or "-"), "Toss" (check/cross), "Shelf" `shelf_treasure_count / 12`.
- Cards, scoring-status lights (Red/Blue Near/Far), signal and commit controls: stock.

### 9.3 Audience display

- In-match overlay: two live values per alliance beside the score: "Treasure" = `treasure_count`, and
  "Shelf" = `shelf_treasure_count` shown as `n/12` in qualification matches (12 = the Scoring RP threshold
  setting) and as `n` in playoffs. Ids `leftTreasure/rightTreasure`, `leftShelf/rightShelf`. The live score
  is `score`; nothing is subtracted or hidden during the match.
- Final score breakdown rows, in order, centre label with values left and right, all POINTS:
  1. "Leave" (`leave_points`)
  2. "Auto Balance" (`auto_balance_points`)
  3. "Auto Treasure" (`auto_treasure_points`)
  4. "Teleop Treasure" (`teleop_treasure_points`)
  5. "Endgame" (`endgame_points`)
  6. "Toss" (`toss_points`)
  7. "Foul" (`foul_points`)
  then, hidden in playoffs, check/cross rows 8. "Auton RP", 9. "Scoring RP", 10. "Endgame RP".
  Ten rows (the placeholder has nine): adjust row height. Ids
  `{left,right}Final{Leave,AutoBalance,AutoTreasure,TeleopTreasure,Endgame,Toss,Foul}Points`,
  `{left,right}Final{Auton,Scoring,Endgame}RankingPoint`.

### 9.4 Announcer display (score posted)

Per alliance: Leave, Auto Balance, Auto Treasure, Teleop Treasure, Crown (location text and `+N`, or "-"),
Endgame, Toss, Foul points; "Auton RP", "Scoring RP", "Endgame RP" as Yes/No with "(foul)" appended when
`auton_rp_by_foul` / `endgame_rp_by_foul` is true; Final Score; total Ranking Points (qualification only);
foul list (Minor/Major, team, rule number with text as tooltip); cards; updated rankings. The in-match
announcer view shows `score` with nothing subtracted.

### 9.5 Wall display and alliance station display

- Wall: match timer, both scores (`score`, nothing subtracted) and the same two live values as the audience
  overlay ("Treasure", "Shelf"). No other game fields.
- Alliance station: stock (team number, timer, alliance `score`). No game fields. Neither display may
  reference a field of the outgoing game.

### 9.6 Rankings display, CSV, PDF

- Rankings display columns: "Rank", "Team", "Name", "RS" (RankingPoints / Played, 2 decimals), "Score"
  (average, 2 decimals), "Auton" (average, 2 decimals), "W-L-T", "Played".
- `rankings.csv`: `Rank,TeamId,RankingPoints,ScorePoints,AutonPoints,Wins,Losses,Ties,Disqualifications,Played`
  (totals).
- Rankings PDF: Rank, Team, RP, RS, Score, Auton, W-L-T, DQ, Played; every column needs a width entry.
- Match review RP symbols: one each for Auton, Scoring, Endgame.

### 9.7 Edit Match Result (`/match_review/{id}/edit`)

Per alliance: fieldset "Autonomous" [numbers "Floor" `auto_floor`, "First" `auto_first`, "Top" `auto_top`;
"Leave" checkbox per team; "Auto Balance" checkbox per team]; fieldset "Teleop + Endgame" [numbers "Floor",
"First", "Top", "Stacked"]; fieldset "Crown" [select with the 8 `crown` values, labels "None", "Auto Floor",
"Auto First", "Auto Top", "Floor", "First", "Top", "Stacked"]; fieldset "Endgame" [per team select None /
Park / Balance; checkbox "Toss"]; "Fouls" [Is Major?, team, rule select with tag]; "Cards". Team loops use
the active team count. Saving recomputes points and RPs, including the opponent-violation RPs.

### 9.8 Other

Settings page "Game-Specific" fieldset: "Auton RP threshold" (20), "Scoring RP threshold" (12). No other
report has game columns.

## 10. Assets

```yaml
assets_folder: specs/2026_medieval_mayhem/
logos:   game-logo.png, blinds-logo.png (600x600)    # NOT YET SUPPLIED (section 12)
sounds:
  new:   toss.wav          # musical tone, about 2 s, played 118 s into the match (20 s remaining). NOT YET SUPPLIED
  stock: start.wav, end.wav, resume.wav, warning_sonar.wav (now at 30 s remaining), abort.wav, match_result.wav
colours: no palette change requested
rules_text: specs/2026_medieval_mayhem/medieval.txt   # copy of the manual export
```

The manual describes every period change as an air horn. Stock sounds are kept unless the designers ask
otherwise (section 12).

- **Logo style: round badge.** `game-logo.png` and `blinds-logo.png` are square (600x600; currently the M-Ayhem placeholder). The badge fills the whole centre circle of the audience and wall overlays and the blinds circle, and the match timer is overlaid on the lower part of the badge on a near-opaque white pill so it stays legible. Implemented by `static/css/game_logo.css`, loaded by `templates/audience_display.html` and `templates/wall_display.html`. A banner-shaped logo would not need that stylesheet.

## 11. Worked examples (become tests and the UI acceptance script)

Values used: auto floor 4 / first 8 / top 12; teleop floor 2 / first 5 / top 10 / stacked 8; leave 4; auto
balance 12; park 2; balance 12; toss 2; minor 5; major 10; crown bonus = base value of its location.
Per-robot lists are in station order. Examples A to D are 2v2; E is 3v3.

### Example A: ordinary qualification match (2v2)

| Input | Red | Blue |
|---|---|---|
| leave | Y, Y | Y, N |
| auto_balance | N, N | N, N |
| auto floor / first / top | 0 / 1 / 0 | 1 / 0 / 0 |
| teleop floor / first / top / stacked | 3 / 4 / 2 / 1 | 2 / 3 / 1 / 0 |
| crown | none | teleop_first (one of Blue's 3 first-shelf treasures) |
| endgame | balance, park | park, none |
| toss | Y | N |
| fouls committed | 1 minor MA2616 | none |

| Computed | Red | Blue |
|---|---|---|
| leave_points | 4+4 = 8 | 4 |
| auto_balance_points | 0 | 0 |
| auto_treasure_points | 1x8 = 8 | 1x4 = 4 |
| **auton_points** | 8+0+8 = **16** | 4+0+4 = **8** |
| teleop_treasure_points | 3x2 + 4x5 + 2x10 + 1x8 = 6+20+20+8 = **54** | 2x2 + 3x5 + 1x10 = 4+15+10 = 29, + crown 5 = **34** |
| endgame_points | 12+2 = 14 | 2 |
| toss_points | 2 | 0 |
| **match_points** | 16+54+14+2 = **86** | 8+34+2+0 = **44** |
| foul_points (received) | 0 | 5 |
| **score** | **86** | **49** |
| Auton RP | N (16 < 20; Blue has no MA2603) | N (8) |
| Scoring RP | N (4+2+1 = 7) | N (3+1+0 = 4) |
| Endgame RP | Y (one robot balanced) | N (no balance; Red has no MA2601/MA2602) |
| Qualification RP | win 3 + 1 = **4** | loss 0 + 0 = **0** |

Red wins. Overlay at the end: Red Treasure 1+3+4+2+1 = 11, Shelf 7/12; Blue Treasure 1+2+3+1 = 7, Shelf 4/12.

### Example B: thresholds exactly at the boundary, plus the crown (2v2, qualification)

| Input | Red | Blue |
|---|---|---|
| leave | Y, Y | Y, Y |
| auto_balance | Y, N | N, N |
| auto floor / first / top | 0 / 0 / 0 | 0 / 1 / 0 |
| teleop floor / first / top / stacked | 2 / 6 / 4 / 2 | 1 / 5 / 4 / 2 |
| crown | none | teleop_top (one of Blue's 4 top-shelf treasures) |
| endgame | none, park | balance, balance |
| toss | N | Y |
| fouls committed | none | none |

| Computed | Red | Blue |
|---|---|---|
| leave_points | 8 | 8 |
| auto_balance_points | 12 | 0 |
| auto_treasure_points | 0 | 1x8 = 8 |
| **auton_points** | 8+12+0 = **20** | 8+0+8 = **16** |
| teleop_treasure_points | 2x2 + 6x5 + 4x10 + 2x8 = 4+30+40+16 = **90** | 1x2 + 5x5 + 4x10 + 2x8 = 2+25+40+16 = 83, + crown 10 = **93** |
| endgame_points | 0+2 = 2 | 12+12 = 24 |
| toss_points | 0 | 2 |
| **match_points** | 20+90+2+0 = **112** | 16+93+24+2 = **135** |
| foul_points | 0 | 0 |
| **score** | **112** | **135** |
| Auton RP | **Y** (20 >= 20, exactly at threshold) | **N** (16) |
| Scoring RP | **Y** (6+4+2 = 12, exactly at threshold; the 2 floor treasures are not counted) | **N** (5+4+2 = 11. The crown counts once: counting it twice would wrongly give 12. The auto first-shelf treasure is not counted: counting it would wrongly give 12) |
| Endgame RP | N (park only) | Y |
| Qualification RP | loss 0 + 2 = **2** | win 3 + 1 = **4** |

Blue wins. Overlay: Red Treasure 14, Shelf 12/12; Blue Treasure 1+1+5+4+2 = 13, Shelf 11/12.

### Example C: RPs earned through opponent violations, with fouls (2v2, qualification)

| Input | Red | Blue |
|---|---|---|
| leave | Y, Y | Y, N |
| auto_balance | N, N | N, N |
| auto floor / first / top | 0 / 1 / 0 | 0 / 0 / 0 |
| teleop floor / first / top / stacked | 0 / 2 / 3 / 0 | 4 / 1 / 1 / 0 |
| crown | auto_first (Red's one auto first-shelf treasure is the crown) | none |
| endgame | park, park | none, none |
| toss | Y | N |
| fouls committed | major MA2603 (auto); major MA2601 and major MA2602 (endgame, same incident) | minor MA2606; major MA2604 twice |

| Computed | Red | Blue |
|---|---|---|
| leave_points | 8 | 4 |
| auto_treasure_points | 1x8 = 8, + crown 8 = 16 | 0 |
| **auton_points** | 8+0+16 = **24** | 4+0+0 = **4** |
| teleop_treasure_points | 2x5 + 3x10 = 10+30 = **40** | 4x2 + 1x5 + 1x10 = 8+5+10 = **23** |
| endgame_points | 2+2 = 4 | 0 |
| toss_points | 2 | 0 |
| **match_points** | 24+40+4+2 = **70** | 4+23+0+0 = **27** |
| foul_points (received) | 5 + 10 + 10 = **25** | 10 + 10 + 10 = **30** |
| **score** | 70+25 = **95** | 27+30 = **57** |
| Auton RP | Y by points (24 >= 20; it would be 16 without the crown bonus) | **Y by violation** (Red committed MA2603; Blue's own auton is only 4) |
| Scoring RP | N (2+3 = 5) | N (1+1 = 2) |
| Endgame RP | N (no balance; Blue's MA2604 and MA2606 give no RP) | **Y by violation** (Red committed MA2601 and MA2602; awarded once) |
| Qualification RP | win 3 + 1 = **4** | loss 0 + 2 = **2** |

Red wins. `auton_rp_by_foul` and `endgame_rp_by_foul` are true for Blue, false for Red.
Variant C2: the same match, but Red's auto foul is entered as a major foul with no rule selected. Blue still
receives 30 foul points and score 57, but loses the Auton RP: Blue RP = 0 + 1 = **1**. Selecting MA2603 in
Edit Match Result restores 2.

### Example D: tied playoff match and the (proposed) tiebreak (2v2)

D1, no fouls:

| Input | Red | Blue |
|---|---|---|
| leave | Y, Y | Y, Y |
| auto_balance | Y, N | N, N |
| auto floor / first / top | 0 / 1 / 0 | 0 / 1 / 0 |
| teleop floor / first / top / stacked | 1 / 2 / 1 / 0 | 1 / 2 / 2 / 1 |
| crown | none | none |
| endgame | balance, none | park, park |
| toss | N | Y |

| Computed | Red | Blue |
|---|---|---|
| **auton_points** | 8+12+8 = **28** | 8+0+8 = **16** |
| teleop_treasure_points | 2+10+10 = 22 | 2+10+20+8 = 40 |
| endgame_points / toss_points | 12 / 0 | 4 / 2 |
| **match_points = score** | 28+22+12+0 = **62** | 16+40+4+2 = **62** |

- As a playoff match: score tied 62 to 62. Level 1, major fouls committed: 0 and 0, tied. Level 2,
  auton_points: 28 > 16. **Red wins**, "TIEBREAK: AUTON POINTS".
- The same inputs as a qualification match: TieMatch. Red 1 + Auton Y (28) + Scoring N (2+1 = 3) + Endgame Y
  = **3 RP**. Blue 1 + Auton N (16) + Scoring N (2+2+1 = 5) + Endgame N = **1 RP**.

D2, decided by fouls: as D1 except Blue's teleop top is 1 (not 2) and Red commits one major G415.
Blue teleop = 2+10+10+8 = 30; Blue match_points = 16+30+4+2 = 52; Blue foul_points = 10; Blue score = 62.
Red score = 62 (match 62 + 0). Tied. Level 1: Red committed 1 major foul, Blue committed 0. **Blue wins**,
"TIEBREAK: MAJOR FOULS", even though Red has more auton points.

D3, decided by match points: Red: leave Y,Y; auto first 1; teleop floor 2, top 4; nothing else.
Red auton = 8+8 = 16; teleop = 4+40 = 44; match_points = 60. Blue: leave Y,Y; auto first 1; teleop floor 2,
first 1, top 4; nothing else; Blue commits one minor MA2607. Blue auton = 16; teleop = 4+5+40 = 49;
match_points = 65. Red score = 60+5 = 65; Blue score = 65+0 = 65. Tied. Level 1: 0 and 0 major fouls.
Level 2: 16 and 16. Level 3: match_points 60 < 65. **Blue wins**, "TIEBREAK: MATCH POINTS".

D4: if every level is equal (for example both alliances enter identical inputs), the result is TieMatch and
the base bracket schedules a replay.

### Example E: 3v3 fallback with a bypassed robot (qualification)

| Input | Red (3 robots) | Blue (station 3 bypassed, did not start) |
|---|---|---|
| leave | Y, Y, Y | Y, Y, N |
| auto_balance | N, N, Y | N, N, N |
| auto floor / first / top | 0 / 0 / 0 | 0 / 0 / 1 |
| teleop floor / first / top / stacked | 0 / 3 / 2 / 0 | 0 / 8 / 3 / 1 |
| crown | none | none |
| endgame | balance, balance, park | balance, none, none |
| toss | Y | N |
| fouls | none | none |

| Computed | Red | Blue |
|---|---|---|
| **auton_points** | 12+12+0 = **24** | 8+0+12 = **20** |
| teleop_treasure_points | 15+20 = 35 | 40+30+8 = 78 |
| endgame_points / toss_points | 12+12+2 = 26 / 2 | 12 / 0 |
| **match_points = score** | 24+35+26+2 = **87** | 20+78+12+0 = **110** |
| Auton / Scoring / Endgame RP | Y (24) / N (3+2 = 5) / Y | Y (20, same threshold with a robot missing) / Y (8+3+1 = 12) / Y |
| Qualification RP | loss 0 + 2 = **2** | win 3 + 3 = **6** |

Blue wins. All three Blue teams, including the bypassed one, receive 6 RP; if the head referee red-cards
the bypassed team as a no-show, that team receives 0 and its partners still receive 6.

### Ranking check from A, B and C

One qualification match each; teams A-Red 101/102, A-Blue 103/104, B-Red 105/106, B-Blue 107/108,
C-Red 109/110, C-Blue 111/112. Sort by RP, then Score, then Auton:

| Rank | Teams | RP | Score | Auton |
|---|---|---|---|---|
| 1-2 | 107, 108 | 4 | 135 | 16 |
| 3-4 | 109, 110 | 4 | 95 | 24 |
| 5-6 | 101, 102 | 4 | 86 | 16 |
| 7-8 | 105, 106 | 2 | 112 | 20 |
| 9-10 | 111, 112 | 2 | 57 | 4 |
| 11-12 | 103, 104 | 0 | 49 | 8 |

Partners tie on every criterion and are ordered by the random value. If the designers decide the score
tiebreaker excludes foul points (section 12, Q19), C-Red drops to 70 and ranks below A-Red (86).

## 12. Open questions

Tracked in [GAME_QUESTIONS.md](GAME_QUESTIONS.md), one row per question with the default this spec uses. Event-format questions (playoff round length) are listed there too but are not part of the game.
