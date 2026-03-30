# bb — Bitbucket Cloud CLI

A CLI for Bitbucket Cloud modeled on GitHub's `gh`. Written in Go. Binary: `bb`.

## Quick Reference

```bash
# Build
make build          # builds bin/bb
make install        # installs to /usr/local/bin/bb

# Development
make test           # go test -race ./...
make lint           # golangci-lint run ./...
make completions    # generates ./completions/ (bash/zsh/fish)

# Release (dry run — no publish)
make release-dry-run

# Run directly during development
go run ./cmd/bb --help
go run ./cmd/bb version --json version,commit
```

## Architecture

The project follows the [github.com/cli/cli](https://github.com/cli/cli/blob/trunk/docs/project-layout.md) layout:

```
cmd/bb/main.go              — 4-line entry point
pkg/cmd/root/root.go        — NewCmdRoot(): wires Factory, registers all commands
pkg/cmd/<noun>/<verb>/      — one package per command (e.g., pkg/cmd/pr/create/)
pkg/cmdutil/factory.go      — Factory DI container passed to every command
pkg/iostreams/              — TTY detection, color, --no-tty support
pkg/api/                    — HTTP client + Bitbucket pagination
pkg/auth/                   — OAuth 2.0 flow + token storage (Phase 2)
pkg/config/                 — Viper-based config (~/.config/bb/config.yaml)
pkg/gitcontext/             — Infers workspace+repo from git remote URL
pkg/output/                 — Table renderer (tabwriter) + JSON+jq (gojq)
internal/version/           — Build-time version injection via ldflags
```

### Factory Pattern (DI)

Every command receives `*cmdutil.Factory` — never access globals:

```go
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
    opts := &ListOptions{IO: f.IOStreams, HttpClient: f.HttpClient}
    ...
}
```

### Workspace Resolution

Every command uses `f.Workspace()` to resolve the active workspace. Chain:

1. `--workspace` flag
2. `BITBUCKET_WORKSPACE` env var
3. `git remote get-url origin` parsed for workspace slug
4. `default_workspace` in config file
5. Error: "run `bb workspace use <slug>`"

### `--json` / `--jq` Pattern

All data commands use `cmdutil.AddJSONFlags(cmd)`. `--json` requires an explicit field list:

```
bb pr list --json id,title,state,author
bb pr list --jq '.[].title'   # implies --json
```

`--json` alone without fields is an error (matching `gh` behaviour).

### Auth (Phase 2 — not yet implemented)

- Primary: OAuth 2.0 Authorization Code + loopback redirect (`pkg/auth/oauth.go`)
- Fallback: Bitbucket API Token via `BITBUCKET_USERNAME` + `BITBUCKET_TOKEN` env vars (already wired in root.go)
- Token storage: `~/.config/bb/tokens.json` (0600 perms) with exclusive file lock on refresh
- ⚠️ App passwords are sunset **June 9, 2026** — only API tokens supported
- ⚠️ Rotating refresh tokens enforced **May 4, 2026** — store new token on every refresh

## Key Dependencies

| Package                  | Purpose                                           |
| ------------------------ | ------------------------------------------------- |
| `spf13/cobra`            | Command structure                                 |
| `spf13/viper`            | Config + env binding                              |
| `golang.org/x/oauth2`    | OAuth 2.0 + PKCE                                  |
| `itchyny/gojq`           | `--jq` expression support                         |
| `charmbracelet/huh`      | Interactive prompts (replaces archived go-survey) |
| `charmbracelet/lipgloss` | Terminal styling                                  |
| `goreleaser/goreleaser`  | Cross-platform binary releases                    |

## Distribution

Release artifacts (via GoReleaser on tag push):

- GitHub Releases: macOS universal, linux-amd64, linux-arm64, windows-amd64
- Homebrew: `brew install chandrasekar-r/tap/bb`
- Scoop (Windows): `scoop bucket add chandrasekar-r https://github.com/chandrasekar-r/scoop-bucket && scoop install bb`
- deb/rpm packages
- `curl -fsSL .../install.sh | sh`

## GitHub Secrets Required for Release

| Secret               | Purpose                                                 |
| -------------------- | ------------------------------------------------------- |
| `BB_OAUTH_CLIENT_ID` | Bitbucket OAuth consumer client ID (embedded in binary) |
| `HOMEBREW_TAP_TOKEN` | GitHub PAT with write access to homebrew-tap repo       |
| `SCOOP_BUCKET_TOKEN` | GitHub PAT with write access to scoop-bucket repo       |

## Notes for Claude Code

- **Always read a file before editing it** — never assume its current content
- **Never commit credentials** — `BB_OAUTH_CLIENT_ID` is injected at build time, not stored in source
- **New commands go in `pkg/cmd/<noun>/<verb>/`** — one package per command, follow the Factory pattern
- **Test with `go test -race ./...`** — the race detector catches auth token refresh races
- **`CGO_ENABLED=0` is required** — all builds must produce static binaries (enforced in Makefile and GoReleaser)
- The `--no-tty` flag + destructive operations → must return a `*cmdutil.NoTTYError` (not silently proceed)
- When implementing issue commands, always check `has_issues` on the repo first
- Pipeline watch polling interval default: 10s; minimum: 5s with warning

## Implementation Status

- ✅ Phase 1 — Foundation
- ✅ Phase 2 — Authentication (`bb auth login/logout/status/token`)
- ✅ Phase 3 — Workspace & Repositories
- ✅ Phase 4 — Branches & Pull Requests
- ✅ Phase 5 — Pipelines
- ✅ Phase 6 — Issues & Snippets
- ✅ Phase 7 — Distribution & Polish

**All phases complete.** The CLI is feature-complete and production-ready.
