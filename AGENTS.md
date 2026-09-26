# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the entry point for the Go web server. Core domains live in top-level packages such as `field/`, `game/`, `network/`, `partner/`, `playoff/`, `tournament/`, and `websocket/`. Web UI assets are in `web/`, `static/`, and `templates/`. Pre-generated schedules are in `schedules/`. BoltDB data is stored in `db/` (and test fixtures in `*_test.db` files at the repo root).

## Build, Test, and Development Commands
See `go.mod` for what version of Go to use.
1. `go build`
   Builds the `cheesy-arena` binary in the repo root.
1. `./cheesy-arena`
   Runs the server; open `http://localhost:8080` in a browser.
1. `go test ./...`
   Runs all Go tests across packages. Should be run after making any code changes to ensure nothing is broken.
1. `go fmt ./...`
   Formats all Go code in the repo. Should be run after making any code changes to ensure consistent style.

## Coding Style & Naming Conventions
Follow standard Go style: tabs for indentation, exported names in `CamelCase`, unexported in `camelCase`. Format code with `gofmt` before submitting changes. If you update a set of enum-style constants, run `go generate ./...` to refresh the generated enum string helpers. Keep package names short and domain-focused (matching existing directories like `field`, `game`, `partner`).

Order imports alphabetically without any grouping or empty lines between them or special treatment of standard library vs third-party imports. Update any files that don't adhere to this standard if editing them for other reasons. Don't use goimports.

## Testing Guidelines
Tests are Go `*_test.go` files co-located with packages (for example `field/`, `game/`, `partner/`, `playoff/`). Use `go test ./...` for the full suite and `go test ./field -run TestName` to target specific areas. When adding new behavior, add or update tests in the same package and prefer table-driven tests for coverage.

## Commit & Pull Request Guidelines
Commit messages in this repo are short, imperative sentences (for example “Fix driver station TCP reads”) and often include an issue/PR number in parentheses (for example “... (#258)”). Keep to that style.

PRs should include:
1. A clear summary of the change.
1. Test notes (exact commands run, for example `go test ./...`).
1. UI screenshots when changing pages in `web/`, `static/`, or `templates/`.

## Configuration & Ops Notes
Cheesy Arena is designed to run as a local web server and uses BoltDB for data. For field networking and hardware integrations, see the project README and relevant `field/` or `plc/` code before making behavioral changes.

## M-Ayhem Base Workflows
This repository is Team 766's M-Ayhem base: upstream Cheesy Arena minus a strip list, plus 2v2 mode and compatibility with the M-Ayhem Arduino PLC, with the current year's game. Read `docs/DEVELOPMENT.md` before changing anything.

Two playbooks cover the recurring work. Follow them rather than improvising:

1. `docs/agents/sync-upstream.md`: regenerate or port the base from upstream; the checkpoint lives in `UPSTREAM.md`.
1. `docs/agents/apply-game.md`: implement a year's game from a spec in `specs/` (format: `docs/agents/game-spec-format.md`).

Do not merge upstream into this repository, do not add a runtime game engine, code generator, build tags or mode flags for games, and do not reintroduce anything on the strip list in `docs/DEVELOPMENT.md`. Keep the module path `github.com/Team254/cheesy-arena`.
