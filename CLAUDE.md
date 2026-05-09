# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and test

```bash
go build -o pulp .          # build binary
go test ./...               # run all tests
go test -run TestFoo ./...  # run a single test
```

No Makefile, linter config, or code generation exists. Tests use only the Go standard library (`testing` package, table-driven tests).

## Architecture

Pulp is a Go CLI that extracts clean markdown from web content for AI workflows. The module path is `github.com/Baseplayer23893/Pulp`.

**Entry points** (`main.go`):
- `pulp` (no args) → launches TUI if stdout is interactive, else prints help
- `pulp mcp` → bypasses Cobra, runs a JSON-RPC 2.0 MCP server on stdin/stdout (`mcp.go`)
- All other subcommands → Cobra CLI (`cmd/root.go`)

**Layers:**

```
main.go / mcp.go              — dispatch to CLI or MCP
cmd/                           — Cobra commands (extract, youtube, reddit, instagram, pdf, package, doctor, serve, install)
  cmd/tui/                     — Bubble Tea terminal UI
internal/
  defuddle/                    — subprocess wrapper around the external `defuddle` CLI
  cleaner/                     — markdown post-processing (tracking params, whitespace, YAML frontmatter)
  config/                      — JSON config + history persistence (~/.config/pulp/)
  storage/                     — file output and skill.zip generation
  yamlutil/                    — YAML-safe string quoting
```

**External CLI dependencies** (must be installed separately): `defuddle` (npm) for web extraction, `yt-dlp` (pipx) for YouTube/Instagram metadata.

## TUI pattern

The TUI (`cmd/tui/`) uses Bubble Tea + Lip Gloss. There are two parallel implementations — `menu.go` (state-based) and `model.go` (screen-based), likely mid-refactor.

The TUI **shells out to the same `pulp` binary** for extractions instead of calling internal packages directly:
```go
exec.CommandContext(ctx, exe, modeName, url, "-q")
```
Output is captured by setting `PULP_FORCE_STDOUT=1`. This means CLI changes automatically benefit the TUI.

Platform detection: `detectSource()` in both TUI files maps URLs to source types automatically (youtube.com, reddit.com, instagram.com, .pdf, or generic web).

## Config and history

Stored as JSON in the OS config directory (`~/.config/pulp/` on Linux). Managed by `internal/config/`. Key config fields: `output_dir`, `default_format`, `max_history`, `auto_copy_result`. History is a separate `history.json` file with capped entries. No Viper — just `encoding/json`.

## Version

Current version: `0.3.1` (hardcoded in `cmd/root.go` and `mcp.go`). Update both places on version bump.