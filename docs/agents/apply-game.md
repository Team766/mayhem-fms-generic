# Playbook: apply-game

Replace the game that is currently in the tree with the game described by a **game spec**.

This is a playbook for a coding agent (any LLM) or a careful human. It is phase 2 of two; phase 1 is
[sync-upstream.md](sync-upstream.md). The spec format is in [game-spec-format.md](game-spec-format.md).
`specs/CURRENT` names the spec of the game that is in the tree now.

## The idea

The output is ordinary Cheesy Arena code: a typed `Score` struct, a hand-readable `Summarize()`,
templates and JS that name the game's elements. There is no runtime config, no code generator and no
generic engine. The spec is the durable artifact; when the base moves to a new upstream, the same spec is
applied again.

The playbook replaces the game that is present. Normally that is last year's game, described by the spec
that `specs/CURRENT` names. When the base has just been regenerated there is no game (upstream's was
stripped) and `CURRENT` is empty.

## Inputs

- The game spec (`specs/<year>_<name>.md`) and its assets folder (`specs/<year>_<name>/`: logos, sounds, rule text).
- `docs/BASE.md` "Placeholder-game seam": the allowlist of files this playbook may touch.
- The tree's current game, as the pattern to follow for plumbing (websocket commands, template helpers, test fixtures).

## Ground rules

1. **Stay inside the seam.** If the spec needs something outside the allowlist (a new arena hook, PLC I/O, a new display), stop and ask, or do it as a clearly separate commit and say so. A yearly game must not quietly change base behaviour.
2. **The spec decides; the old code does not.** Where the spec is silent or contradicts itself, ask; do not copy last year's behaviour. Record every assumption you had to make in the report.
3. **Robot count.** Write rules against the robots that are actually playing (2v2 or 3v3, bypassed robots), using the base's active-stations helper (`docs/TwoVTwo.md`). "All robots" never means a literal 3.
4. **Every consumer of the score JSON gets updated.** The 2025 hand port missed the announcer, wall and alliance-station displays and the rankings PDF; they rendered `NaN` or failed. The consumer list in `docs/BASE.md` is a checklist, and step 6 checks it mechanically.
5. Keep upstream's style: imports alphabetical and ungrouped, table-driven tests, `go fmt`.
6. Work on a new branch. Do not push or open PRs unless asked.

## Procedure

1. **Read the spec fully. List open questions first.** Ambiguities in ranking points, tiebreak direction, what counts toward a threshold, and what each screen shows are cheap to settle now and expensive after 40 files. If the operator is available, ask; otherwise choose, and flag it.
2. **Derive two vocabularies.**
   - *Outgoing*: the ids and display names in the `specs/CURRENT` spec, plus the Go, JSON and CSS names the code derived from them (read `game/score.go` and `score_summary.go`). Nothing is kept between years; the list is rebuilt from the old spec each time.
   - *Incoming*: the names you will use, from the spec's ids. Put both in the PR description.
3. **Model and math** (`game/`): `Score`, `ScoreSummary`, `Summarize()`, `Equals()`, ranking fields and sort order, `DetermineMatchStatus` playoff tiebreakers, foul point constants, the rules list, match timing and sounds if the spec changes them. Write the tests in the same commit:
   - one table-driven test per scoring element and phase;
   - every ranking point at its boundary (threshold − 1, threshold);
   - every tiebreaker level, each side winning;
   - each **worked example** in the spec as a test with the spec's expected numbers;
   - 2v2 and bypassed-robot cases for any rule that mentions robots.
   Update the shared fixtures (`game/test_helpers.go`) last and fix the expectations they shift in `tournament/`, `web/`, `model/` and `field/`.
4. **Entry** : scoring panel (`web/scoring_panel.go` commands, template, JS, CSS, near/far split per the spec), referee panel status rows, edit-match-result form and `match_review.js`. Every spec element must be enterable live *and* editable after the match.
5. **Display and reporting**: audience overlay and final-score breakdown, wall display and `display_shared.js`, announcer score-posted template, alliance-station display, rankings display, rankings CSV and PDF columns, match review RP symbols, settings page game section (only if the spec declares tunable thresholds). Install assets from the spec's folder (`static/img/game-logo.png`, `blinds-logo.png`, sounds).
6. **Completeness checks.**
   - No outgoing vocabulary is left: `git grep -n -i -E '<word1>|<word2>|...' -- '*.go' '*.html' '*.js' '*.css' '*.csv' ':!specs' ':!docs'` prints nothing (explain any hit you keep).
   - `git grep -n 'ScoreSummary\.\|\.Score\.' -- static/js templates` : every hit names a field that exists in the new structs.
   - `git diff --stat <base>` touches only seam files (plus `specs/`). Explain any exception.
   - Set `specs/CURRENT` to the new spec's file name.
7. **Verify.**
   - `go generate ./... && go fmt ./... && go vet ./... && go build ./... && go test ./...`
   - Run the server (`go build && ./cheesy-arena -dev`) and play each worked example from the spec through the real UI: enter it on the scoring panels, add the fouls on the referee panel, commit, and compare the audience final score, RPs and rankings with the spec's expected numbers. Do it once in each alliance size the spec declares. Then edit the result and confirm the summary recomputes. Attach screenshots of every screen in the spec's "Screens" section.
   - Open the announcer, wall, alliance-station, rankings and field-monitor displays during a match; none may show `NaN`, `undefined` or a template error.

## Report

End with: spec file and commit, assumptions made where the spec was silent, files touched outside the seam and why, worked examples and whether each matched, screens checked, anything not verified, and proposed edits to the spec or to this playbook. Do not claim a check you did not run.
