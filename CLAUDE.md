# CLAUDE.md

## Project Overview

fuzz.fish is a TUI fuzzy finder for the Fish shell, written in Go. It provides interactive search across:
- Fish shell history (Ctrl+R)
- Git branches (Ctrl+G)
- Files in the current directory (Ctrl+S)

It is built with [Bubble Tea](https://charm.land/bubbletea) and distributed as a Fish shell plugin.

## Architecture

```
cmd/fuzz/main.go          # Entry point
internal/app/             # TUI application (Bubble Tea MVC)
  model.go                # Application state and async load messages
  run.go                  # Program initialization and result output
  update.go               # Key handling and state transitions
  filter.go               # Fuzzy filtering and scoring
  view.go                 # Rendering
internal/git/             # Git branch listing and preview
internal/history/         # Fish history file parsing
internal/files/           # Directory walker
internal/scoring/         # Frecency scoring algorithm
internal/ui/              # Styles, colors, format helpers
conf.d/fuzz.fish          # Fish shell integration (keybindings, binary install)
```

The app follows the Bubble Tea model: `Init` → async `tea.Cmd` messages → `Update` → `View`.

## Development

### Build

```sh
make build       # build binary to ./fuzz
make install     # build and install to ~/.config/fish/functions/fuzz
```

### Test

```sh
go test ./...
go test -v ./...
```

### Lint

```sh
go vet ./...
# golangci-lint is run in CI (version v2.8.0)
golangci-lint run
```

CI runs on every push and PR to `main`: build, `go vet`, `go test`, and `golangci-lint`.

## Code Conventions

- Follow standard Go package layout: `cmd/` for binaries, `internal/` for non-exported packages.
- The Bubble Tea pattern is `Model` (state) → `Update` (events) → `View` (render). Keep them separated across `model.go`, `update.go`, and `view.go`.
- Async data loading is done via `tea.Cmd` functions returning typed messages (e.g., `historyLoadedMsg`).
- Slice reuse: avoid per-keystroke allocations in hot paths (`filter.go`). Reuse slices with `[:n]` when capacity allows.
- Silenced errors that are intentional must use `//nolint:errcheck` with a comment explaining why.
- Do not add comments or docstrings to code you did not change.
- Do not add speculative abstractions or error handling for scenarios that cannot happen.

## Security

- **Shell injection in fuzz.fish**: The Fish script passes branch names into `fish -c "git switch --quiet '$branch'"`. Branch names come from `git`'s own reference list (not free-form user input), but avoid constructing shell strings with untrusted values in future changes. Prefer passing arguments as a list (`fish -c 'git switch' -- $branch`) or using native `git switch $branch` directly.
- **File path handling**: `files.Collector` uses `filepath.WalkDir` anchored to cwd. Do not allow user-supplied root paths without validation.
- **History file**: The parser reads `~/.local/share/fish/fish_history` directly. Never write to or modify this file.
- **`FUZZ_FISH_BIN_PATH` env var**: `conf.d/fuzz.fish` uses this variable in `rm -f` and build paths. Do not make it user-writable in multi-user environments.
- **Network**: The only network call is `git clone` in `_fuzz_fish_rebuild_binary` during install/update. It always clones from the canonical GitHub repository.
- **`exec.Command` in Go**: The `git pull` call in `run.go` passes the branch name as a separate argument to `exec.Command`, avoiding shell injection. Always use argument lists, never `exec.Command("sh", "-c", userInput)`.

## Git Workflow

### Branching

Use descriptive prefixes:
- `feat/` — new feature
- `fix/` — bug fix
- `docs/` — documentation only
- `refactor/` — no behavior change

### Commit

```sh
git add <specific files>
git commit -m "short imperative summary"
```

Never use `git add -A` or `git add .` unless you have reviewed every changed file. Never skip hooks (`--no-verify`).

### Push and Pull Request

```sh
git push -u origin <branch>
gh pr create --title "<short title>" --body "$(cat <<'EOF'
## Summary
- bullet 1
- bullet 2

## Changes
- bullet 1
EOF
)"
```

**PR description rules:**
- Write in **English**.
- Do **not** include a "Test plan" section.
- Do **not** add "Made with Claude Code" or similar AI attribution.
- Keep the title under 70 characters.
- Use the body for details, not the title.

### Release

Releases are created via GitHub Actions (`release.yml`) with manual `workflow_dispatch`. Choose `patch`, `minor`, or `major` version bump. Do not create tags manually.
