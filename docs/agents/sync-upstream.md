# Playbook: sync-upstream

Regenerate or update the **M-Ayhem base** from upstream [Team254/cheesy-arena](https://github.com/Team254/cheesy-arena).

This is a playbook for a coding agent (any LLM) or a careful human. It is phase 1 of two; phase 2 is
[apply-game.md](apply-game.md). Read [../BASE.md](../BASE.md) first: it defines what the base is. This
file says how to produce it.

## The idea

We do not merge upstream. Upstream rewrites the same ~60 game files every January, and a fork that also
rewrote them conflicts in all of them (the 2026 attempt stopped at 39 conflicted files). Instead the base
is **upstream, minus a documented strip list, plus a documented feature list**, re-applied as intent.
`UPSTREAM.md` records the last upstream commit that was reviewed.

There are two modes. Pick by distance from the checkpoint.

| Mode | When | What happens |
|------|------|--------------|
| `regenerate` | First time, after upstream's January game swap, or when the checkpoint is more than ~40 commits or one release behind | Start from a clean upstream tree and re-apply the strip list and the feature list |
| `port` | Routine catch-up during a season | Review `checkpoint..upstream`, port the game-agnostic commits, skip the rest |

## Inputs

- Upstream ref to sync to. Default: the latest upstream release tag; `upstream/main` if a needed fix is not released yet. Never fetch unless the operator asked you to; use the refs that are already there and say which commit you used.
- `UPSTREAM.md` (the checkpoint), `docs/BASE.md` (strip list, keep list, feature list), `docs/TwoVTwo.md`, `docs/agents/reference/*` (detailed inventories, dated; trust the code over them when they disagree, and update them).
- The current base branch, for the feature code you are carrying forward.

## Ground rules

1. **Stay textually close to upstream.** In files you are only stripping, delete whole lines or blocks. Do not reflow, regroup imports, rename, fix typos or touch copyright headers. No `goimports`; upstream keeps imports alphabetical and ungrouped.
2. **Module path stays `github.com/Team254/cheesy-arena`.** It keeps upstream diffs applicable without a rewrite pass.
3. **Strip completely.** A removed integration leaves no settings fields, template sections, routes, JS, tests or enum values behind.
4. **Never reintroduce** anything on the "never reintroduce" list in `docs/BASE.md`, and do not copy Lite-specific behaviour (list in `docs/agents/reference/strip-list.md` section 3). Lite is a map of *where* game content lives, not a source of code.
5. **PLC signal order is an interface.** The generic inputs, registers and coils keep upstream's order. Season I/O and M-Ayhem I/O go after the generic block. After any enum change run `go generate ./...`.
6. **The placeholder game is not your job here.** In `regenerate` mode you stop with the upstream season game stripped to the neutral seam described in `docs/BASE.md`; `apply-game` then installs the placeholder from `specs/high_seas_havoc.md`. In `port` mode you leave the game files alone.
7. Work on a new branch. Do not push, open PRs, or delete branches unless asked.

## Procedure: `regenerate`

1. **Branch.** Create the work branch from the base branch. Replace the tree's tracked content with the chosen upstream tree, keeping everything listed under "Files the base owns" in `docs/BASE.md` (docs, specs, playbooks, scripts, `UPSTREAM.md`, `AGENTS.md` additions, M-Ayhem assets). Commit this as one mechanical commit: `Import upstream <tag> (<sha>)`. Nothing else goes in that commit, so reviewers can skip it.
2. **Strip integrations**, one commit each, in this order: TBA, Nexus, team signs, Twitch. Use the call-site lists in `reference/strip-list.md` section 2, but re-derive them with `git grep` because upstream moves. Build and run tests after each.
3. **Strip the season game and LEDs** down to the neutral seam. One commit. The seam file allowlist is in `docs/BASE.md`; everything game-specific outside it (arena hooks, PLC game I/O, LED package, settings fields, sounds) goes.
4. **Apply the base features**, one commit each, from `docs/BASE.md` "Feature list":
   1. 2v2 mode, following `docs/TwoVTwo.md` (invariants first, then touchpoints; new upstream screens need handling too).
   2. M-Ayhem PLC wire map.
   3. Any other listed feature.
   Carry the previous base's implementation forward where it still fits; re-implement against the new upstream code where it does not. Port the feature's tests with it.
5. **Hand off to `apply-game`** with `specs/high_seas_havoc.md` to install the placeholder game. Separate commit(s).
6. **Verify** (next section), then update `UPSTREAM.md`: checkpoint sha, date, upstream tag, mode, and the skip list.

## Procedure: `port`

1. List the range: `scripts/agents/upstream-range.sh` prints `checkpoint..<upstream ref>` with touched paths.
2. Classify every commit: **port** (game-agnostic), **skip-game** (touches only the season game, LEDs, or a stripped integration), **partial** (mixed: port the generic half). Write the table into the PR description and the skips into `UPSTREAM.md`.
3. Port in upstream order, one commit per upstream commit where practical, message `Port upstream <sha>: <subject>`. For each ported commit ask: does it add a screen, a loop over alliance stations, or a PLC signal? If so it needs 2v2 handling (`docs/TwoVTwo.md`) or a wire-map entry.
4. Verify, then move the checkpoint to the last commit you **reviewed** (not the last you ported).

If classification shows the season game was swapped or more than about a third of the commits are `partial`, stop and switch to `regenerate`.

## Verification gates

All must pass before the checkpoint moves. Paste the results into the PR.

- `go generate ./... && go fmt ./... && go vet ./... && go build ./... && go test ./...`
- `scripts/agents/check-leftovers.sh base` returns nothing (season-game, LED, TBA, Nexus, team-sign, Twitch and Lite vocabulary).
- Diff review against upstream: `git diff <upstream sha> -- . ':!docs' ':!specs' ':!scripts'` contains only strip-list deletions, feature-list additions and placeholder-game seam files. Anything else is a defect.
- 2v2 checklist and 3v3 flip-back check from `docs/TwoVTwo.md`.
- PLC unit tests, including the wire-map guard test. Bench test with the Arduino when hardware is available; say so if it was not run.
- Run the server (`go build && ./cheesy-arena -dev`), play one placeholder-game match end to end in 3v3 and one in 2v2: scoring panels, referee foul, commit, audience final score, rankings, edit result. Attach screenshots.

## Report

End with: upstream commit used, mode, commits ported/skipped/partial with reasons, features re-implemented rather than carried, gates run and their results, anything not verified, and questions for the maintainers. Do not claim a gate you did not run.
