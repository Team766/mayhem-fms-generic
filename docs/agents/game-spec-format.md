# Game spec format

A game spec is one Markdown file that tells [apply-game.md](apply-game.md) what this year's game is. It is
read by a person or an LLM, never parsed by the FMS, so it can say things in prose that a config schema
cannot. Small YAML blocks are used where exact names and numbers matter. Start from the spec that
`specs/CURRENT` names, or from [`specs/high_seas_havoc.md`](../../specs/high_seas_havoc.md).

Specs are named `specs/<year>_<name>.md` and are never deleted: the newest one is the game in the tree,
and `apply-game` reads it next year to know what to replace.

Put assets next to it in a folder of the same name: `game-logo.png`, `blinds-logo.png` (600x600), any
replacement sounds, the rules text.

## Sections

| # | Section | Must contain |
|---|---------|--------------|
| 1 | Overview | One paragraph a volunteer could read. Field vocabulary. `game.id` (lower_snake, becomes the code vocabulary) and display name |
| 2 | Alliance size | 2v2, 3v3, or both. Whether any rule or threshold changes with robot count or bypassed robots |
| 3 | Match timing | Auto, pause, teleop, endgame warning. Say "stock" if unchanged |
| 4 | Scoring elements | **Counters**: id, display name, points per phase, which phases, who enters it (near or far scoring panel, referee). **Per-robot statuses**: yes/no or named levels with points, phase. **Derived totals**: groups shown on screens and used by tiebreakers |
| 5 | Fouls | Point values and wording (minor/major/tech). Any foul that changes ranking points, exactly how |
| 6 | Rules list | Number, text, minor/major, flags. May point to a file in the assets folder |
| 7 | Ranking points | Win/tie values. Each bonus RP as an exact sentence: what is counted, in which phases, the threshold, the comparison (`>=`), how bypassed or absent robots are treated |
| 8 | Ranking order and tiebreakers | Qualification sort order. Playoff tiebreakers in order, and for each which side wins (more or fewer) |
| 9 | Screens | For each of: scoring panels, referee panel, audience overlay, audience final score, announcer, wall, rankings display, reports, edit result: what is shown, in what order, with what labels |
| 10 | Assets | Logos, sounds, colours |
| 11 | Worked examples | At least three complete matches with every input and the expected points, RPs and winner, computed by hand. Include a threshold boundary, a tie, and one match in each alliance size. These become tests and the UI acceptance script |
| 12 | Open questions | Anything undecided. The playbook asks about these before writing code |

## Writing rules that save a round trip

- Give every element an `id`. Ids become Go fields, JSON keys and CSS ids, so keep them short and unique.
- State comparisons and directions outright: "at least 14", "the alliance with **fewer** foul points against it wins the tiebreak".
- Say what does *not* count ("auto cannonballs do not count toward the Scoring RP").
- If last year's behaviour should be kept, say so; the playbook does not read last year's code for intent.
- Tunable numbers (RP thresholds you may change at the event) should be marked `tunable: true`; they become event settings. Everything else is a constant.

## Sample prompts

Apply a new game:
> Follow `docs/agents/apply-game.md` with the spec `specs/2027_<name>.md`. Ask me the open questions before you change code. Work on a new branch off the base and do not push.

Try a rule change during design:
> Using `docs/agents/apply-game.md`, update the game in this branch for the edits I just made to `specs/2027_<name>.md` sections 7 and 11. Show me the worked examples before and after.

Bring the base up to date:
> Follow `docs/agents/sync-upstream.md` in `port` mode up to upstream tag `v2027.0.1`. Give me the classification table before porting anything.

Re-apply this year's game after a base update:
> The base moved to a new upstream. Re-apply `specs/2027_<name>.md` with `docs/agents/apply-game.md` on a fresh branch off the base and tell me what differed from the previous application.
