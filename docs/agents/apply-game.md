# Playbook: apply-game

Replace the game in the tree with the game described by a **game spec**.

For a coding agent (any LLM) or a careful person. It is phase 2 of two; phase 1 is
[sync-upstream.md](sync-upstream.md). The spec format is in [game-spec-format.md](game-spec-format.md);
`specs/CURRENT` names the spec of the game that is in the tree now.

## The idea

The output is ordinary Cheesy Arena code: a typed score struct, a `Summarize()` a volunteer could check against
the manual's point table, and screens that name the game's elements. No runtime config, no generator, no engine.
The spec is the durable artifact; when the tree moves to a new upstream, the same spec is applied again.

## Ground rules

1. **Stay inside the game seam** as [../DEVELOPMENT.md](../DEVELOPMENT.md) defines it. If the spec needs a base change (arena, PLC, playoffs, a new display), stop and ask, or do it as a separate pull request.
2. **The spec decides; last year's code does not.** Where the spec is silent or contradicts itself, ask. If nobody can answer, choose, record the question in the spec's questions file, and say so.
3. **Find every consumer.** This playbook names no files on purpose. Anything that enters, reads or displays the score, the summary or the rankings must be updated, including the displays nobody looks at during development. Find them with `git grep`; step 5 checks you did.
4. **Robot count.** Rules about "all robots" use the robots actually playing (see [../TwoVTwo.md](../TwoVTwo.md)), never a literal 3.
5. Upstream's style: typed fields, table-driven tests, imports alphabetical and ungrouped.
6. New branch. Do not push or open pull requests unless asked. If you think you need a destructive command, stop and ask.

## Procedure

1. **Read the whole spec and settle open questions first.** They are cheap now and expensive after forty files.
2. **Name what goes and what comes.** Outgoing: the ids and labels in the `specs/CURRENT` spec and the Go, JSON and CSS names derived from them. Incoming: the names you will use, from the new spec's ids. Put both in the pull request.
3. **Model and math, with tests in the same commit**: every scoring element in every phase; every ranking point at its threshold and one below; every tiebreak level with each side winning; every rule that mentions robots, in each alliance size; and **each worked example in the spec as a test with the spec's numbers**. If a test disagrees with the spec, work out by hand which is wrong; fix the spec if its arithmetic is wrong, never bend the code to a wrong number.
4. **Entry, then display and reporting.** Every element must be enterable live and editable after the match, by the scorer the spec assigns it to. Every screen and report shows what the spec's "Screens" section says, with third-robot controls following the 2v2 layout rule. Install the spec's assets.
5. **Completeness checks.**
   - No outgoing name is left: `git grep -n -i -E '<outgoing words>'` over Go, templates, scripts, styles and CSV, excluding `docs/` and `specs/`, prints nothing you cannot explain.
   - Every score, summary or ranking field that a script or template reads exists in the Go structs (grep the field accesses and compare).
   - `git diff --stat <branch you started from>` touches only the game seam. Explain any exception.
   - Point `specs/CURRENT` at the new spec.
6. **Verify.**
   - `go generate ./... && go fmt ./... && go vet ./... && go build ./... && go test ./...`, with no new vet warnings.
   - Run the server and play each worked example through the real screens, once per alliance size the spec declares: enter it on the scoring panels, add the fouls on the referee panel, commit, and compare the final score, ranking points and rankings with the spec. Edit the result and confirm it recomputes.
   - Open every display during a match. None may show an error, `NaN` or `undefined`.

## Report

Spec and commit; assumptions made where the spec was silent; anything touched outside the seam and why; each
worked example and whether it matched; screens checked and not checked; proposed edits to the spec or to this
playbook. Do not claim a check you did not run.
