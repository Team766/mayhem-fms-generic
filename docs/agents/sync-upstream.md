# Playbook: sync-upstream

Regenerate or update the **M-Ayhem base** from upstream [Team254/cheesy-arena](https://github.com/Team254/cheesy-arena).

This is a playbook for a coding agent (any LLM) or a careful human. It is phase 1 of two; phase 2 is
[apply-game.md](apply-game.md). Read [../DEVELOPMENT.md](../DEVELOPMENT.md) first: it defines what the base is. This
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
| `update` | Routine catch-up during a season | Review `checkpoint..upstream`, bring over the game-agnostic commits, skip the rest |

## Inputs

- Upstream ref to sync to. Default: the latest upstream release tag; `upstream/main` if a needed fix is not released yet. Never fetch unless the operator asked you to; use the refs that are already there and say which commit you used.
- `UPSTREAM.md` (the checkpoint), `docs/DEVELOPMENT.md` (strip list, keep list, feature list), `docs/TwoVTwo.md`, `docs/agents/reference/*` (detailed inventories, dated; trust the code over them when they disagree, and update them).
- The current base branch, for the feature code you are carrying forward.

## Ground rules

1. **Stay textually close to upstream.** In files you are only stripping, delete whole lines or blocks. Do not reflow, regroup imports, rename, fix typos or touch copyright headers. No `goimports`; upstream keeps imports alphabetical and ungrouped.
2. **Module path stays `github.com/Team254/cheesy-arena`.** It keeps upstream diffs applicable without a rewrite pass.
3. **Strip completely.** A removed integration leaves no settings fields, template sections, routes, JS, tests or enum values behind.
4. **Never reintroduce** anything on the "never reintroduce" list in `docs/DEVELOPMENT.md`, and do not copy Lite-specific behaviour (list in `docs/agents/reference/strip-list.md` section 3). Lite is a map of *where* game content lives, not a source of code.
5. **PLC signal order is an interface.** The generic inputs, registers and coils keep upstream's order. Season I/O and M-Ayhem I/O go after the generic block. After any enum change run `go generate ./...`.
6. **The game is not your job here.** In `regenerate` mode you stop with upstream's season game stripped to the neutral seam described in `docs/DEVELOPMENT.md`; `apply-game` then installs the current year's spec. In `update` mode you leave the game files alone.
7. Work on a new branch. Do not push, open PRs, or delete branches unless asked.

## Procedure: `regenerate`

1. **Branch.** Create the work branch from the base branch. Replace the tree's tracked content with the chosen upstream tree, keeping everything listed under "Files the base owns" in `docs/DEVELOPMENT.md` (docs, specs, playbooks, `UPSTREAM.md`, `AGENTS.md` additions, M-Ayhem assets). Commit this as one mechanical commit: `Import upstream <tag> (<sha>)`. Nothing else goes in that commit, so reviewers can skip it.
2. **Strip integrations**, one commit each within a single strip pull request, in this order: TBA, Nexus, team signs, Twitch. Use the call-site lists in `reference/strip-list.md` section 2, but re-derive them with `git grep` because upstream moves. Build and run tests after each.
3. **Strip the season game and LEDs** down to the neutral seam. One commit, same pull request; the PR should be almost entirely deletions. The seam file allowlist is in `docs/DEVELOPMENT.md`; everything game-specific outside it (arena hooks, PLC game I/O, LED package, settings fields, sounds) goes.
4. **Apply the base features**, one pull request each, from `docs/DEVELOPMENT.md` "Feature list":
   1. 2v2 mode, following `docs/TwoVTwo.md` (invariants first, then touchpoints; new upstream screens need handling too).
   2. M-Ayhem Arduino PLC compatibility (`docs/DEVELOPMENT.md`, "PLC"): remove the FTA-ready start condition and badge, keep the signal-address guard test passing.
   3. Any other listed feature.
   Carry the previous base's implementation forward where it still fits; re-implement against the new upstream code where it does not. Port the feature's tests with it.
5. **Hand off to `apply-game`** with the current year's spec. Separate pull request.
6. **Verify** (next section), then update `UPSTREAM.md`: checkpoint sha, date, upstream tag, mode, and the skip list.

## Procedure: `update`

1. List the range: `git log --reverse --date=short --format='%n%h %ad %s' --name-only <checkpoint>..<upstream ref>`, with the checkpoint from `UPSTREAM.md`.
2. Classify every commit: **port** (game-agnostic), **skip-game** (touches only the season game, LEDs, or a stripped integration), **partial** (mixed: bring over the generic half). Write the table into the PR description and the skips into `UPSTREAM.md`. A skipped generic fix is silent, so skips get extra checks:
   - **Judge by hunk, not by file.** A commit is `skip-game` only if every hunk is game-only code. `field/arena.go`, the scoring and referee panel scripts, the display scripts and the settings page mix game and generic code; a generic fix can ride inside a commit that looks like a game commit.
   - **Cross-check against Lite.** Team 254's cheesy-arena-lite ports the same upstream range (its checkpoint is in its `UPSTREAM.md`). If Lite brought over a commit you skipped, look again.
   - **Challenge pass.** Have a second, cheaper model or a person read the skip list and argue why each skipped commit might be generic. Re-check anything it flags.
   - The yearly `regenerate` starts from a clean upstream tree, so a generic fix that was wrongly skipped is lost for one season at most.
3. Port in upstream order, one commit per upstream commit where practical, message `Port upstream <sha>: <subject>`. For each ported commit ask: does it add a screen, a loop over alliance stations, or a PLC signal? If so it needs 2v2 handling (`docs/TwoVTwo.md`), or a check against the Arduino's fixed addresses (`plc/mayhem_arduino_test.go`).
4. Verify, then move the checkpoint to the last commit you **reviewed** (not the last you ported).

If classification shows upstream's season game was swapped or more than about a third of the commits are `partial`, stop and switch to `regenerate`.

## Verification gates

All must pass before the checkpoint moves. Paste the results into the PR.

- `go generate ./... && go fmt ./... && go vet ./... && go build ./... && go test ./...`
- Nothing stripped is left: `git grep -n -i -E 'twitch|nexus|tbaclient|tbapublish|teamsign|team_sign|ledcontroller|/api/scores|FoulPointsAgainst|cheesy-arena-lite|<this season game words>' -- '*.go' '*.html' '*.js' '*.css' ':!docs' ':!specs' ':!static/js/lib' ':!static/css/lib'` prints nothing except `TbaMatchKey` (the vendored icon font has a `.bi-twitch` glyph, hence the last exclusion). Take the season's game words from upstream's `game/score.go` for the release you synced to.
- Diff review against upstream: `git diff <upstream sha> -- . ':!docs' ':!specs'` contains only strip-list deletions, feature-list additions and placeholder-game seam files. Anything else is a defect.
- 2v2 checklist and 3v3 flip-back check from `docs/TwoVTwo.md`.
- PLC unit tests, including the Arduino signal-address guard test. Bench test with the Arduino when hardware is available; say so if it was not run.
- Run the server (`go build && ./cheesy-arena -dev`), play one placeholder-game match end to end in 3v3 and one in 2v2: scoring panels, referee foul, commit, audience final score, rankings, edit result. Attach screenshots.

## Report

End with: upstream commit used, mode, commits ported/skipped/partial with reasons, features re-implemented rather than carried, gates run and their results, anything not verified, and questions for the maintainers. Do not claim a gate you did not run.
