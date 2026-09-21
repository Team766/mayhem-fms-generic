# Game questions: Medieval Mayhem (M-Ayhem 2026)

Questions for the game authors and event organizers. The spec (`2026_medieval_mayhem.md`) and the FMS use the **default** until a question is answered. When one is answered: fill in the answer and date, change the spec (and its worked examples) if the answer differs from the default, and re-run `docs/agents/apply-game.md` for the affected sections.

Status: `open`, `answered` (spec updated), `n/a`.

| # | Question | Default in use | Answer | Status |
|---|----------|----------------|--------|--------|
| 1 | Crown: does "x2" double only the crown's own placement value, or the alliance's whole score? | only its own placement |  | open |
| 2 | Crown: is there exactly one crown on the field? 4.1.2 says "one of each mound" but also "19 treasures including the crown". | one. |  | open |
| 3 | Crown: may it be stacked (16), and does it count once toward the 12-treasure Scoring RP? | yes, once. |  | open |
| 4 | A treasure stacked during auto "is scored as regular points": is that the auto value of the level beneath it (4/8/12), the flat stacked 8, or the teleop level value? | auto value of the level beneath it (4/8/12) |  | open |
| 5 | Does an auto placement have to stay in place until the end of the match to score, and does it then keep its auto value? | yes to both. |  | open |
| 6 | 4.3 says endgame scoring is "as in auton or teleop". Do endgame-period placements score at teleop values? | yes. |  | open |
| 7 | Does a treasure stacked on a FLOOR treasure count toward the Scoring RP ("on first and top shelves ... including stacked")? | yes, all stacked treasures count (one Stacked counter). |  | open |
| 8 | Are auto balance (12) and endgame balance (12) per robot, so two robots earn 24? | per robot. |  | open |
| 9 | Safe-house park (2): what exactly qualifies (fully inside own safe house at the end?), is it per robot, and is it exclusive with balance? 2.3.3 omits it and the table cites 4.1.3, which does not define it. | per robot, fully inside at the end, exclusive with balance. |  | open |
| 10 | Endgame RP says "parked on balance beam": must the robot meet the full 12-point balanced condition? | yes, the RP needs one robot scored as Balance. |  | open |
| 11 | MA2610 (more than one preload): do that robot's auto placements score nothing, or teleop values if still in place at the end? | teleop values, entered in the teleop counters (so first/top ones count toward the Scoring RP). |  | open |
| 12 | Does setting Auto Balance imply Leave (16 total), or can a robot balance without being credited a leave? | independent toggles; the scorer sets both. |  | open |
| 13 | The Toss: is there one shared Treasure Chest or one per alliance, where does it stand, and who judges it? | one per-alliance yes/no entered by the FAR scorer. |  | open |
| 14 | "Throwing the treasure early is a foul": minor or major, and what rule number? Unsafe throw (minor) and the optional MA2621 foul also have no foul number or severity. | all minor, entered without a rule. |  | open |
| 15 | "Initiating damaging contact" and "reaching onto the field" are major fouls with no rule number. Assign numbers? | entered as rule-less major fouls. |  | open |
| 16 | 2.1 says points "can be taken away via penalties"; 4.5 says fouls ADD 5/10 to the other alliance. | add to the other alliance, never subtract. |  | open |
| 17 | Auton RP "score 20 points": is it at least 20, counting leave, auto balance, auto placements and the crown's doubling? | yes. |  | open |
| 18 | If the event switches to 3v3, do the thresholds (20 auton points, 12 treasures, one robot balanced) change? | no. |  | open |
| 19 | Tiebreaker 1 "average match score": does it include foul points received (literal reading), or is it match points without fouls (what the earlier FMS attempt used)? | includes foul points (literal reading of the manual) |  | open |
| 20 | Ranking Score "rounded to 2 decimal places": should rounding be able to create ties? | no, compare exactly and round only for display. |  | open |
| 21 | Red card in a qualification match: does only the carded team lose its RP and points (stock FMS), or does the whole alliance score 0? | only the carded team (upstream behaviour) |  | open |
| 22 | Did-not-start (6.6) is garbled. Does a bypassed team whose drive team is present receive its alliance's RPs? | yes; only a no-show (red card) gets 0. |  | open |
| 23 | Tied playoff match: the manual gives no procedure, and 6.2 says no replays. | fewer major fouls committed, then more auton points, then more match points without fouls, then replay. |  | open |
| 24 | Playoff bracket: single-game semifinals and a best-of-three final need a base change; the base plays best of three in every round. | use the base bracket unchanged with 4 alliances. |  | open |
| 25 | Length of the pause between auto and teleop? | 3 s. |  | open |
| 26 | The manual describes air horns at auto start, teleop start, endgame start and match end. Replace the stock sounds? | keep stock sounds, add only `toss.wav`. Who supplies `toss.wav`? |  | open |
| 27 | Is the balance-beam indicator light standalone hardware, or should the FMS read it? | standalone hardware; the FMS does not read it |  | open |
| 28 | Comment [a] says "the balance beam is no longer in the safe zone", but 3.6 and 3.8 say it is. Which is right? (Affects the wording of MA2603/MA2606 only.) | no default; affects only the wording of MA2603/MA2606 |  | open |
| 29 | Control limit: text shows "three two" and "3 2". | 2. |  | open |
| 30 | Is this export the final post-kickoff manual? It still says values may change before the September kickoff. | numbers as exported on 2026-09-20 |  | open |
| 31 | Logos (`game-logo.png`, `blinds-logo.png`) are not supplied. | keep the existing logos until supplied |  | open |
| 32 | G401's text is cut off in the manual ("during AUTO and"). | the shortened text in section 6. |  | open |
| 33 | MA2604 (hoarding) is a major foul that repeats every 10 seconds. A long hoard could swing a match by 30 points or more. Is that intended, or should repeats be capped or be minor? | as written in the manual: major (10), a separate foul for each repeat |  | open |
| 34 | The Toss cue plays 20 s before the end and needs a sound. Who supplies `toss.wav`? | placeholder: a copy of the stock resume sound |  | open |

## Event format (not part of the game spec)

| # | Question | Default in use | Answer | Status |
|---|----------|----------------|--------|--------|
| F1 | Manual 6.5.2: single elimination, final best-of-three "if there is still time". What does the FMS run? | Single-game rounds before the final; final best-of-three. | Decided 2026-09-20 (Debajit): single game except for the final; nothing else needs support. A single-game final, if time runs out, is handled by hand. If a game in a single-game round ends in a tie, the next game of that series is played. | answered |
