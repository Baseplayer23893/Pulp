# AGENTS.md

## Build and test

```bash
go build -o pulp .          # build binary
go test ./...               # run all tests
go test -run TestFoo ./...  # run a single test
```

- No Makefile, linter, or code generation exists
- Tests use only Go standard library (`testing` package, table-driven)

## Entry points

`main.go` has three modes:
- `pulp` (no args) → TUI if stdout is interactive, else help
- `pulp mcp` → JSON-RPC 2.0 MCP server on stdin/stdout (`mcp.go`)
- `pulp <subcommand>` → Cobra CLI (`cmd/root.go`)

## Version

Version is **hardcoded in two places**: `cmd/root.go` and `mcp.go`. Update both on version bump.

## External dependencies

Must be installed separately:
- `defuddle` (npm) — web page extraction
- `yt-dlp` (pipx) — YouTube/Instagram metadata and transcripts

## Architecture

```
main.go / mcp.go              — dispatch to CLI or MCP
cmd/                          — Cobra commands (extract, youtube, reddit, instagram, pdf, package, doctor, serve, install)
internal/
  cache/                      — URL markdown cache (~/.cache/pulp/)
  defuddle/                   — subprocess wrapper around external defuddle CLI
  cleaner/                    — markdown post-processing
  config/                     — JSON config (~/.config/pulp/)
  storage/                    — file output and skill.zip generation
```

## TUI pattern

The TUI (`cmd/tui/`) shells out to the **same pulp binary** for extractions instead of calling internal packages:
```go
exec.CommandContext(ctx, exe, modeName, url, "-q")
```
Output is captured via `PULP_FORCE_STDOUT=1` env var. CLI changes automatically benefit the TUI.

## Config

Stored as JSON in `~/.config/pulp/` on Linux. Uses `encoding/json`, not Viper.

## Cache

URL markdown cache for faster repeated extractions:
- Location: `~/.cache/pulp/`
- Default TTL: 24 hours
- Commands: `extract`, `youtube`, `instagram`, `pdf` (not Reddit)
- Flag: `--no-cache` to bypass cache and force fresh extraction