# Agent playbooks

Two playbooks keep this FMS current. They are plain Markdown so any coding agent, or a person, can follow them.

| Playbook | Use it when | Input | Output |
|----------|-------------|-------|--------|
| [sync-upstream.md](sync-upstream.md) | Upstream Cheesy Arena has changes you want | An upstream tag or commit | The base, regenerated or ported, with `UPSTREAM.md` moved |
| [apply-game.md](apply-game.md) | You have this year's rules | A game spec ([format](game-spec-format.md)) and assets | A branch with the game implemented and verified |

Always two phases with a commit between them: first the base, reviewed against upstream; then the game,
reviewed against the base. A typical year: `sync-upstream` to the latest upstream release in the autumn,
write `specs/<year>_<name>.md`, `apply-game`, iterate on the spec as the rules settle, then freeze a
release branch for the event.

Background and contracts: [../DEVELOPMENT.md](../DEVELOPMENT.md), [../TwoVTwo.md](../TwoVTwo.md). Dated inventories that
the playbooks cite: [reference/](reference/).

Tool wrappers (`.claude/skills/`) only point here. Add an equivalent pointer for another tool if it needs one; keep the content in these files.
