# Release Checklist

Use this before tagging a release.

## Automated Checks

```bash
GOCACHE=/tmp/pulp-go-cache go test ./...
GOCACHE=/tmp/pulp-go-cache go test -race ./...
GOCACHE=/tmp/pulp-go-cache go vet ./...
GOCACHE=/tmp/pulp-go-cache go build ./...
```

## Environment Checks

```bash
pulp doctor
pulp doctor --network-url https://example.com
```

Confirm:

- `defuddle` is installed.
- `yt-dlp` is installed.
- Clipboard tools are available.
- Config directory is writable.
- Output directory is writable.
- Network fetch succeeds.

## CLI Smoke Tests

```bash
pulp extract https://example.com -o /tmp/pulp-example.md
pulp package test --source /tmp/pulp-example.md -o /tmp/pulp-package-test
pulp config set output_dir /tmp/pulp-output
pulp config
```

Confirm:

- `/tmp/pulp-example.md` contains frontmatter and markdown.
- `/tmp/pulp-package-test/test.zip` or the expected sanitized zip exists.
- Config output directory is persisted.

## TUI Manual Tests

Run:

```bash
pulp tui
```

Test in at least one normal terminal and one constrained context such as `tmux` or a small window.

Confirm:

- `https://example.com` extracts successfully.
- Save writes to the configured output directory.
- Copy either reaches the system clipboard or reports the fallback path.
- Settings output directory persists after restarting the TUI.
- History shows more entries than fit on screen and can navigate to older entries.
- History delete and re-run work.
- Resizing the terminal does not break result/history/settings views.

## Release

Only tag after the checklist above passes:

```bash
git tag v0.4.0
git push origin main --tags
```
