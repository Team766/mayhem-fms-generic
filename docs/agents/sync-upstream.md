# Playbook: sync-upstream

Bring this repository up to date with upstream [Team254/cheesy-arena](https://github.com/Team254/cheesy-arena).

For a coding agent (any LLM) or a careful person. It is phase 1 of two; phase 2 is [apply-game.md](apply-game.md).
Read [../DEVELOPMENT.md](../DEVELOPMENT.md) first: it says what we remove, keep and add, and why. This file says
how to get there and how to prove you did.

## The idea

We never merge upstream. Upstream rewrites its game code every January, and a fork that also rewrote it
conflicts everywhere. Instead the tree is upstream, minus what `DEVELOPMENT.md` removes, plus what it adds,
re-applied as intent. `UPSTREAM.md` records the last upstream commit that was reviewed.

| Mode | When | What happens |
|------|------|--------------|
| `update` | Routine catch-up | Review `checkpoint..upstream`, bring over the game-agnostic commits, skip the rest |
| `regenerate` | After upstream's January game swap, or when `update` would be mostly conflicts | Start from a clean upstream tree and re-apply our removals and additions |

## Ground rules

1. **Find, don't assume.** This playbook names no files on purpose. Locate what to change with `git grep` and by reading the code, every time.
2. **Stay textually close to upstream.** Where you are only removing, delete whole lines or blocks. Do not reflow, rename, regroup imports or reformat. Imports stay alphabetical and ungrouped.
3. **Remove completely**, as `DEVELOPMENT.md` defines it, and keep what it says must survive.
4. **Keep the module path** `github.com/Team254/cheesy-arena`, so upstream diffs stay comparable.
5. **PLC signal order is an interface** with fixed-address hardware. Never reorder the generic signals; run `go generate ./...` after any change to a signal list.
6. **The game is not your job.** In `update` mode leave game code alone. In `regenerate` mode stop at the no-game state; `apply-game` follows as its own pull request.
7. New branch. Use the upstream refs already fetched unless asked to fetch, and say which commit you used. Do not push or open pull requests unless asked. If you think you need a destructive command, stop and ask.

## `update`

1. List the range: `git log --reverse --date=short --format='%n%h %ad %s' --name-only <checkpoint>..<upstream ref>`.
2. Classify every commit as **bring over**, **skip** (touches only upstream's game or something we removed) or **partial**. A wrongly skipped fix is silent, so:
   - judge by hunk, not by file: generic fixes ride inside commits that look like game commits;
   - compare with what [cheesy-arena-lite](https://github.com/Team254/cheesy-arena-lite) brought over for the same range (its `UPSTREAM.md` records its checkpoint), and look again at any difference;
   - have a second reader (a cheaper model or a person) argue why each skip might be generic.
3. Bring commits over in upstream order, one commit each where practical. For each, ask whether it adds a screen, a loop over alliance stations or a PLC signal: those need 2v2 handling ([../TwoVTwo.md](../TwoVTwo.md)) or a look at the Arduino's fixed addresses.
4. Run the gates. Move the checkpoint to the last commit you **reviewed**, and record the skips with reasons in `UPSTREAM.md` and the pull request.

Switch to `regenerate` if upstream's game was swapped or most commits are partial.

## `regenerate`

One pull request per step, so each can be reviewed against the one before:

1. **Import**: the chosen upstream tree plus the files `DEVELOPMENT.md` says survive a regeneration, and nothing else. It is verified, not read: `git diff --stat <upstream ref> HEAD` must list only those files.
2. **Remove**: one commit per row of "What we remove", ending at the no-game state. Almost all deletions.
3. **Add**: one pull request per row of "What we add". Carry last year's implementation forward where it still fits; re-implement against the new upstream code where it does not. Bring its tests.
4. Record the checkpoint. Then hand off to `apply-game` with the spec named in `specs/CURRENT`.

## Gates

All must pass before the checkpoint moves. Report each honestly, including the ones you could not run.

- `go generate ./... && go fmt ./... && go vet ./... && go build ./... && go test ./...`, with no new vet warnings.
- **Nothing removed is left.** Build a word list from "What we remove" plus the identifiers of upstream's current game (read them from its score model), and `git grep -n -i -E '<words>'` over Go, templates, scripts and styles, excluding `docs/`, `specs/` and vendored libraries. Explain every hit you keep.
- **Everything kept still works**: the tests for fouls, referee flow, cards, playoffs, alliance selection, displays and the PLC pass without being weakened.
- **The diff against upstream is explainable**: every hunk of `git diff <upstream ref> -- . ':!docs' ':!specs'` is a removal, an addition or the game.
- 2v2 and the 3v3 flip-back check from `TwoVTwo.md`. The PLC address guard test. The Arduino bench test ([../ArduinoPlc.md](../ArduinoPlc.md)) when hardware is available.
- Run the server and play a match end to end in each alliance size: panels, a referee foul, commit, audience final score, rankings, edit result. No page may show an error, `NaN` or `undefined`.

## Report

Upstream commit used and mode; commits brought over, skipped and partial, with reasons; what was re-implemented
rather than carried; each gate and its result; what was not verified; questions for the maintainers.
