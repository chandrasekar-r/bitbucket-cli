---
title: "feat: bb — Bitbucket Cloud CLI"
type: feat
status: active
date: 2026-03-30
origin: docs/brainstorms/2026-03-30-bb-cli-requirements.md
---

# bb — Bitbucket Cloud CLI

## Overview

Build `bb`, a Go-based CLI tool for Bitbucket Cloud modeled on GitHub's `gh` CLI. It provides authentication, repository management, pull request workflows, branch control, pipeline monitoring, issue tracking, snippets, and workspace management — all from the terminal with `--json` + `--jq` output for scripting and interactive prompts for human use.

**Binary:** `bb` | **Language:** Go | **Target:** Bitbucket Cloud REST API v2.0

---

## Problem Statement

Bitbucket Cloud has no official CLI tool. Teams are forced to use the web UI or raw API calls for common developer workflows. `gh` has defined the gold standard for what a VCS CLI should be. `bb` fills that gap for Bitbucket Cloud users with a familiar command idiom and first-class cross-platform support.

---

## Proposed Solution

A statically compiled Go binary distributed via GitHub Releases, Homebrew, Scoop/Winget, and deb/rpm packages. Commands follow the `bb <noun> <verb>` pattern (`bb pr create`, `bb repo clone`, `bb pipeline watch`). The tool infers workspace and repo context from git remote URLs, uses OAuth 2.0 for authentication with Bitbucket API Token as a headless fallback, and outputs tables for humans and JSON for scripts.

Key decisions carried forward from origin document:
- **Go** — single static binary, factory DI pattern (see origin: `docs/brainstorms/2026-03-30-bb-cli-requirements.md`)
- **Binary name: `bb`** — mirrors `gh`, Bitbucket's natural abbreviation
- **OAuth 2.0 Authorization Code + loopback redirect** as primary auth; **Bitbucket API Token** (Basic auth) as fallback — *NOT app passwords, which are sunset June 9, 2026*
- **Bundled OAuth consumer credentials** injected at build time via `ldflags`
- **Interactive prompts by default**, `--no-tty` disables
- **`--json [fields]` + `--jq [expr]`** on all data commands

---

## Technical Approach

### Architecture

The project lives at `/Users/rc/Boomi/bb/`. Module path: `github.com/yourorg/bb` (update to actual GitHub org).

```
bb/
├── cmd/
│   └── bb/
│       └── main.go                     # 5-line entry point: calls root.Execute()
├── pkg/
│   ├── cmd/
│   │   ├── root/
│   │   │   └── root.go                 # NewCmdRoot(f *Factory) — registers all top-level commands
│   │   ├── auth/
│   │   │   ├── auth.go                 # NewCmdAuth(f) — group command
│   │   │   ├── login/login.go          # OAuth 2.0 loopback + API token fallback
│   │   │   ├── logout/logout.go
│   │   │   ├── status/status.go
│   │   │   └── token/token.go
│   │   ├── repo/
│   │   │   ├── repo.go
│   │   │   ├── list/list.go
│   │   │   ├── view/view.go
│   │   │   ├── create/create.go        # Offers --clone after creation
│   │   │   ├── clone/clone.go
│   │   │   ├── fork/fork.go
│   │   │   ├── delete/delete.go        # Requires --force in --no-tty mode
│   │   │   ├── rename/rename.go
│   │   │   └── browse/browse.go
│   │   ├── pr/
│   │   │   ├── pr.go
│   │   │   ├── list/list.go
│   │   │   ├── view/view.go
│   │   │   ├── create/create.go        # huh prompts; fork detection
│   │   │   ├── merge/merge.go
│   │   │   ├── approve/approve.go
│   │   │   ├── decline/decline.go
│   │   │   ├── checkout/checkout.go
│   │   │   ├── comment/comment.go
│   │   │   ├── diff/diff.go
│   │   │   └── browse/browse.go
│   │   ├── branch/
│   │   │   ├── branch.go
│   │   │   ├── list/list.go
│   │   │   ├── create/create.go
│   │   │   ├── delete/delete.go
│   │   │   ├── rename/rename.go
│   │   │   └── protect/protect.go
│   │   ├── pipeline/
│   │   │   ├── pipeline.go
│   │   │   ├── list/list.go
│   │   │   ├── view/view.go
│   │   │   ├── run/run.go
│   │   │   ├── cancel/cancel.go
│   │   │   └── watch/watch.go          # Polling loop; exit code mirrors pipeline result
│   │   ├── issue/
│   │   │   ├── issue.go
│   │   │   ├── list/list.go
│   │   │   ├── view/view.go
│   │   │   ├── create/create.go
│   │   │   ├── close/close.go
│   │   │   ├── reopen/reopen.go
│   │   │   └── comment/comment.go
│   │   ├── snippet/
│   │   │   ├── snippet.go
│   │   │   ├── list/list.go
│   │   │   ├── view/view.go
│   │   │   ├── create/create.go
│   │   │   ├── edit/edit.go
│   │   │   ├── delete/delete.go
│   │   │   └── clone/clone.go
│   │   ├── workspace/
│   │   │   ├── workspace.go
│   │   │   ├── list/list.go
│   │   │   ├── use/use.go
│   │   │   └── view/view.go
│   │   ├── completion/completion.go    # bb completion [bash|zsh|fish|powershell]
│   │   └── version/version.go         # bb version
│   ├── cmdutil/
│   │   ├── factory.go                  # Factory DI container (IOStreams, HttpClient, Config)
│   │   ├── errors.go                   # Typed CLI errors (AuthError, NotFoundError, etc.)
│   │   └── json_flags.go               # Shared --json / --jq / --limit flag registration
│   ├── iostreams/
│   │   └── iostreams.go                # TTY detection, color, IsStdoutTTY(), IsStderrTTY()
│   ├── api/
│   │   ├── client.go                   # HTTP client: auth middleware, retry, rate limit handling
│   │   ├── pagination.go               # Auto-paginate following `next` URL; respects --limit
│   │   ├── repos.go
│   │   ├── prs.go
│   │   ├── branches.go
│   │   ├── pipelines.go
│   │   ├── issues.go
│   │   ├── snippets.go
│   │   └── workspaces.go
│   ├── auth/
│   │   ├── oauth.go                    # OAuth 2.0 Authorization Code + loopback redirect
│   │   ├── store.go                    # File-based token store (0600); exclusive lock on write
│   │   └── refresh.go                  # Token refresh with rotation; flock before read-refresh-write
│   ├── config/
│   │   └── config.go                   # Viper-based config; XDG on Unix, %APPDATA% on Windows
│   ├── gitcontext/
│   │   └── context.go                  # Infer workspace + repo slug from git remote URL
│   └── output/
│       ├── table.go                    # tabwriter-based human-readable tables
│       └── json.go                     # --json + --jq via itchyny/gojq
├── internal/
│   └── version/
│       └── version.go                  # Version string injected at build via ldflags
├── scripts/
│   ├── completions.sh                  # Generates completion files for GoReleaser packaging
│   └── install.sh                      # curl | sh universal installer
├── .goreleaser.yaml
├── .github/
│   └── workflows/
│       ├── ci.yml                      # go test + go vet + golangci-lint on every PR
│       └── release.yml                 # GoReleaser triggered on tag push
├── Makefile
├── CLAUDE.md
└── go.mod
```

### Key Design Decisions

#### Workspace Resolution (SpecFlow Gap 1 — resolved)

Every API call needs a workspace slug. Resolution order (highest to lowest priority):

1. `--workspace` flag on the command
2. `BITBUCKET_WORKSPACE` environment variable
3. Inferred from git remote URL of the current directory (`gitcontext.FromRemote()`)
4. `default_workspace` in `~/.config/bb/config.yaml`
5. If multiple workspaces exist and none of the above apply → prompt user to `bb workspace use <slug>`

`gitcontext.FromRemote()` parses `git remote get-url origin` and extracts the workspace slug from SSH (`git@bitbucket.org:WORKSPACE/repo.git`) and HTTPS (`https://bitbucket.org/WORKSPACE/repo.git`) URLs.

#### Authentication & Token Storage (SpecFlow Gap 2 — resolved)

- **Primary:** OAuth 2.0 Authorization Code + localhost loopback redirect (RFC 8252 §7.3 — any port on `127.0.0.1` accepted)
- **Fallback:** Bitbucket API Token via HTTP Basic auth (`username:token`) — set via `bb auth login --with-token` or env vars `BITBUCKET_USERNAME` + `BITBUCKET_TOKEN`
- ❌ **NOT supported:** App passwords (sunset June 9, 2026) or username+password Basic auth

Token storage: **file-based at `~/.config/bb/tokens.json`** with `0600` permissions for V1. OS keychain integration is a follow-on.

**Token refresh race condition:** Before any read-refresh-write cycle, acquire an exclusive file lock (`syscall.LOCK_EX` via `flock` on Unix, `LockFileEx` on Windows). Bitbucket enforces rotating refresh tokens from **May 4, 2026** — the new refresh token from each refresh response MUST be written before releasing the lock. Two concurrent `bb` invocations sharing credentials are safe.

#### OAuth Consumer Credentials

Client ID and optional client secret are embedded at build time via `ldflags`:

```
-X github.com/yourorg/bb/internal/version.OAuthClientID=$(BB_OAUTH_CLIENT_ID)
```

These are resolved from environment variables in the GoReleaser CI environment (GitHub Actions secrets), not hardcoded in source.

#### `--json` Field List (SpecFlow Gap 12 — resolved)

`--json` requires an explicit comma-separated field list (matching `gh` behavior):

```
bb pr list --json id,title,state,author
```

`--json` alone is an error: `"specify fields with --json field1,field2 (run bb pr list --json --help for available fields)"`. `--jq` implies `--json` and accepts any field list or `*` for all fields.

#### Exit Codes (SpecFlow Gap 13 — resolved)

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error (API error, not found, etc.) |
| `2` | Command misuse (bad arguments, missing required flags) |

`bb pipeline watch` additionally:
| Code | Meaning |
|------|---------|
| `0` | Pipeline completed successfully |
| `1` | Pipeline failed / errored / stopped |
| `2` | Watch itself failed (network error, rate limited, timeout) |

#### `--no-tty` + Destructive Operations (SpecFlow Gap 9 — resolved)

When `--no-tty` is active (or stdout is not a TTY), commands that would normally prompt for confirmation instead **error with a clear message**:

```
Error: this operation requires confirmation. Run with --force to skip the prompt.
```

Affected commands: `repo delete`, `branch delete`, `pr merge`, `snippet delete`.

#### Pipeline Watch (SpecFlow Gap 6 — resolved)

No streaming endpoint exists — uses polling with HTTP byte-range requests:

1. `POST /pipelines/` → get `{pipeline_uuid}`
2. Poll `GET /pipelines/{uuid}` every 10 seconds (default; `--poll-interval` flag)
3. For each step in `IN_PROGRESS`: `GET /steps/{step_uuid}/log` with `Range: bytes=N-` to fetch new bytes
4. Print new log bytes as received
5. Warn user if rate approaching limit (based on request count, not headers — Bitbucket doesn't return `X-RateLimit-Remaining`)
6. On completion: exit with code reflecting pipeline result

Poll interval default: 10 seconds. `--poll-interval` minimum enforced at 5 seconds with a warning. A 30-minute pipeline at 10-second intervals consumes ~360 requests (36% of hourly limit).

#### Issues Guard (SpecFlow Gap 7 — resolved)

Before any `bb issue` command, fetch the repo resource and check `has_issues`. If false:

```
Error: Issues are not enabled for this repository.
Enable them under Repository Settings → Features → Issues.
```

#### Fork PR Detection (SpecFlow Gap 8 — resolved)

`bb pr create` compares the workspace extracted from `git remote get-url origin` against the authenticated workspace. If they differ (fork scenario), prompt:

```
? This appears to be a fork. Create PR in upstream repo (upstream-workspace/repo)? [Y/n]
```

If yes, set destination repo to the upstream. If no, create PR within the fork itself.

#### `bb repo create` → Clone Offer (SpecFlow Gap 7 — resolved)

After successful `bb repo create`, when TTY is present:

```
✓ Created repo workspace/my-new-repo
? Clone the repository locally? [Y/n]
```

In `--no-tty` mode: respects a `--clone` flag; does not prompt.

### Library Stack

| Purpose | Library | Version |
|---------|---------|---------|
| Commands | `spf13/cobra` | v1.9+ |
| Config | `spf13/viper` | v1.19+ |
| OAuth 2.0 | `golang.org/x/oauth2` | latest |
| Token storage | File-based (V1); `99designs/keyring` (V2) | — |
| jq filtering | `itchyny/gojq` | v0.12+ |
| Prompts | `charmbracelet/huh` | v0.6+ |
| TUI (future) | `charmbracelet/bubbletea` + `bubbles` | — |
| Terminal styling | `charmbracelet/lipgloss` | v1.0+ |
| Release automation | `goreleaser/goreleaser` | v2.x |
| Linux packages | `goreleaser/nfpm` | bundled with GoReleaser |

> **Note:** `AlecAivazis/survey` is archived — do not use. Use `charmbracelet/huh` for all interactive prompts.

---

## Implementation Phases

### Phase 1: Foundation

**Goal:** A buildable, testable project skeleton with core infrastructure. End state: `bb version` works and `bb --help` shows all command groups.

**Tasks:**

- [ ] Initialize Go module at `/Users/rc/Boomi/bb/` — `go mod init github.com/yourorg/bb`
- [ ] `pkg/iostreams/iostreams.go` — `IOStreams` struct with `Out`, `ErrOut`, `In`; `IsStdoutTTY()`, `IsStderrTTY()`, color detection, `--no-tty` override
- [ ] `pkg/cmdutil/factory.go` — `Factory` struct: `IOStreams *iostreams.IOStreams`, `HttpClient func() (*http.Client, error)`, `Config func() (config.Config, error)`, `BaseURL string`
- [ ] `pkg/config/config.go` — Viper config loader; OS-appropriate path (`~/.config/bb/config.yaml` Unix, `%APPDATA%\bb\config.yaml` Windows); `BITBUCKET_WORKSPACE` env var binding; `bb config get/set` subcommands
- [ ] `pkg/output/table.go` — tabwriter-based table renderer; respects color and `IsStdoutTTY()`
- [ ] `pkg/output/json.go` — `--json fields` + `--jq expr` implementation using `itchyny/gojq`; `--json` alone is an error
- [ ] `pkg/cmdutil/json_flags.go` — `AddJSONFlags(cmd)` helper used by all data commands
- [ ] `pkg/api/client.go` — base HTTP client; `Accept: application/json` header; `User-Agent: bb/VERSION`; 429 exponential backoff; 401 triggers token refresh
- [ ] `pkg/api/pagination.go` — `PaginateAll(client, url, opts)` follows Bitbucket `next` URL; respects `--limit` flag (default 30, max 100)
- [ ] `cmd/bb/main.go` — entry point; `root.NewCmdRoot(factory.New())` → `Execute()`
- [ ] `pkg/cmd/root/root.go` — `NewCmdRoot(f)`: `--workspace`, `--no-tty`, `--json`/`--jq` persistent flags; Viper binding in `PersistentPreRunE`; registers all subcommand groups
- [ ] `pkg/cmd/version/version.go` — `bb version`; `--json` support; build info from `internal/version/`
- [ ] `internal/version/version.go` — `Version`, `BuildDate`, `Commit`, `OAuthClientID` injected via `ldflags`
- [ ] `pkg/cmd/completion/completion.go` — `bb completion [bash|zsh|fish|powershell]`
- [ ] `Makefile` — `.PHONY` + `help` as first target; targets: `build`, `test`, `lint`, `completions`, `install`, `release-dry-run`, `clean`
- [ ] `.goreleaser.yaml` — full config: 5 platforms (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64), universal macOS binary, `CGO_ENABLED=0`, `ldflags` with version/OAuth injection, nfpm `.deb`/`.rpm`, Homebrew cask, Scoop, Winget
- [ ] `.github/workflows/ci.yml` — `go test ./...`, `go vet ./...`, `golangci-lint` on every PR
- [ ] `.github/workflows/release.yml` — GoReleaser on tag push `v*`
- [ ] `scripts/completions.sh` — generate bash/zsh/fish completions for packaging
- [ ] `CLAUDE.md` — project overview, quick commands, architecture, key file roles, notes for Claude

**Key files:**

```go
// cmd/bb/main.go
func main() {
    f := factory.New()
    rootCmd := root.NewCmdRoot(f)
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

### Phase 2: Authentication

**Goal:** `bb auth login/logout/status/token` working end-to-end with OAuth 2.0 and API Token fallback.

**Tasks:**

- [ ] `pkg/gitcontext/context.go` — parse `git remote get-url origin`; extract workspace slug + repo slug from SSH and HTTPS Bitbucket URLs; return `(workspace, repo, error)`. Used by all commands for context inference.
- [ ] `pkg/auth/oauth.go` — OAuth 2.0 Authorization Code loopback flow:
  - Start `net.Listen("tcp", "127.0.0.1:0")` for callback port
  - Build auth URL: `https://bitbucket.org/site/oauth2/authorize?client_id=...&response_type=code&redirect_uri=http://127.0.0.1:PORT/callback&state=RANDOM`
  - Attempt PKCE (`oauth2.S256ChallengeOption`) — if Bitbucket rejects, retry without PKCE
  - Open browser with `pkg/browser` (cross-platform)
  - Wait for callback; extract `code`; exchange for token
  - Request scopes: `repository repository:write pullrequest pullrequest:write issue issue:write pipeline pipeline:write snippet snippet:write account team project`
- [ ] `pkg/auth/store.go` — file-based token store at `~/.config/bb/tokens.json`; `0600` permissions; stores: `access_token`, `refresh_token`, `token_type`, `expiry`, `username`, `workspace_slugs[]`
- [ ] `pkg/auth/refresh.go` — `RefreshToken(store)`:
  - Acquire exclusive file lock before reading stored refresh token
  - POST to `https://bitbucket.org/site/oauth2/access_token` with `grant_type=refresh_token`
  - Atomically write new `access_token` AND new `refresh_token` to store (rotation enforced May 4, 2026)
  - Release lock
  - On refresh failure: return `ErrSessionExpired` → caller shows "run `bb auth login`"
- [ ] `pkg/cmd/auth/login/login.go`:
  - Default: OAuth browser flow
  - `--with-token`: read API token from stdin; prompt for username interactively (if TTY) or from `--username` flag
  - Validate token by calling `GET /user`
  - Store credentials; set default workspace if only one workspace found
  - Show: `✓ Logged in as username (workspace)`
- [ ] `pkg/cmd/auth/logout/logout.go` — remove stored tokens; `--workspace` flag to logout of specific workspace
- [ ] `pkg/cmd/auth/status/status.go` — for each stored account: transparently attempt token refresh; show username, active workspace, token expiry, scopes
- [ ] `pkg/cmd/auth/token/token.go` — print active access token; exit 1 if not authenticated
- [ ] Wire `BITBUCKET_TOKEN` + `BITBUCKET_USERNAME` env vars in `pkg/cmdutil/factory.go` HTTP client builder — if set, bypass stored credentials; used for CI headless mode

**Test stubs:**
- `pkg/auth/oauth_test.go` — test state validation, code extraction from callback URL
- `pkg/auth/store_test.go` — test concurrent lock behavior (two goroutines racing to refresh)
- `pkg/cmd/auth/login/login_test.go` — mock HTTP, test `--with-token` flow

---

### Phase 3: Workspace & Repositories

**Goal:** Full `bb workspace` and `bb repo` command coverage.

**Tasks:**

- [ ] `pkg/api/workspaces.go` — `ListWorkspaces()`, `GetWorkspace(slug)`, `ListMembers(slug)`, `ListProjects(slug)`
- [ ] `pkg/api/repos.go` — `ListRepos(workspace, opts)`, `GetRepo(workspace, slug)`, `CreateRepo(workspace, opts)`, `ForkRepo(workspace, slug, destWorkspace)`, `DeleteRepo(workspace, slug)`, `RenameRepo(workspace, slug, newSlug)` — all return typed structs; pagination via `api.PaginateAll`
- [ ] `pkg/cmd/workspace/list` — table: slug, name, type, role; `--json slug,name,role`
- [ ] `pkg/cmd/workspace/use` — set `default_workspace` in config; validate slug exists via API
- [ ] `pkg/cmd/workspace/view` — show workspace details, member count, project count
- [ ] `pkg/cmd/repo/list` — table: name, description, language, visibility, updated; `--workspace`, `--limit`, `--json name,slug,language,is_private,updated_on`
- [ ] `pkg/cmd/repo/view` — show full repo info: description, clone URLs, default branch, language, open PR count; `--json`
- [ ] `pkg/cmd/repo/create` — `huh` form for name, description, visibility (public/private), init with README; after create: offer `--clone` interactively; `--no-tty` respects `--clone` flag
- [ ] `pkg/cmd/repo/clone` — resolve `workspace/repo` from arg or git context; call `git clone` as subprocess; HTTPS by default, SSH with `--ssh`
- [ ] `pkg/cmd/repo/fork` — fork into authenticated workspace; `--workspace` to fork into specific workspace; offer to clone after
- [ ] `pkg/cmd/repo/delete` — requires `--force` in `--no-tty` mode; interactive confirmation in TTY; soft-check "type the repo name to confirm" pattern
- [ ] `pkg/cmd/repo/rename` — `PUT /repositories/{workspace}/{slug}` with new name/slug
- [ ] `pkg/cmd/repo/browse` — open `https://bitbucket.org/{workspace}/{slug}` in system browser; `--branch`, `--pr`, `--commit` flags to open specific views

**Common `--workspace` resolution** used by all repo commands: `gitcontext` → env var → config default → error.

---

### Phase 4: Branches & Pull Requests

**Goal:** Full `bb branch` and `bb pr` coverage including interactive creation.

**Tasks:**

- [ ] `pkg/api/branches.go` — `ListBranches()`, `GetBranch()`, `CreateBranch()`, `DeleteBranch()`, `GetBranchRestrictions()`, `CreateBranchRestriction(type, pattern)`
- [ ] `pkg/api/prs.go` — `ListPRs()`, `GetPR()`, `CreatePR()`, `MergePR()`, `ApprovePR()`, `RemoveApproval()`, `DeclinePR()`, `AddComment()`, `GetDiff()`, `GetDiffStat()`
- [ ] `pkg/cmd/branch/list` — table: name, last commit SHA (short), last commit date, author; `--json name,target.hash,target.date`
- [ ] `pkg/cmd/branch/create` — `bb branch create <name> [--from <source>]`; source defaults to default branch
- [ ] `pkg/cmd/branch/delete` — `--force` required in `--no-tty`; confirm interactively in TTY
- [ ] `pkg/cmd/branch/rename` — delete old + create new (Bitbucket has no rename endpoint)
- [ ] `pkg/cmd/branch/protect` — create branch restriction: `push` type requiring groups; `--pattern` flag for glob patterns
- [ ] `pkg/cmd/pr/list` — table: #, title, author, from-branch → to-branch, state, approvals; filter flags: `--state`, `--author`, `--assignee`; `--json id,title,state,author.display_name,source.branch.name,destination.branch.name`
- [ ] `pkg/cmd/pr/view` — full PR display: title, description (rendered as markdown in TTY), reviewers + approval status, build status, comments count
- [ ] `pkg/cmd/pr/create`:
  - Fork detection: compare remote workspace to auth workspace; prompt destination if fork
  - `huh` form: title (pre-filled from branch name), description (open in `$EDITOR` or inline), reviewers (multi-select from workspace members), base branch, labels
  - `--draft` flag; `--title`, `--body`, `--reviewer`, `--base` flags bypass prompts
  - `--no-tty`: all flags required or error
- [ ] `pkg/cmd/pr/merge` — strategy selection: `merge commit` (default), `squash`, `fast-forward`; `--strategy` flag; confirm with `--force` in `--no-tty`; `--delete-branch` flag
- [ ] `pkg/cmd/pr/approve` — `POST /pullrequests/{id}/approve`
- [ ] `pkg/cmd/pr/decline` — `--message` for decline reason
- [ ] `pkg/cmd/pr/checkout` — fetch PR source branch; `git checkout -b {branch} --track origin/{branch}`
- [ ] `pkg/cmd/pr/comment` — general comment via `--body`; open `$EDITOR` if no `--body` and TTY present
- [ ] `pkg/cmd/pr/diff` — fetch from `/pullrequests/{id}/diff`; pipe through `git diff` color formatter or `delta` if installed
- [ ] `pkg/cmd/pr/browse` — open PR URL in browser

---

### Phase 5: Pipelines

**Goal:** `bb pipeline` commands with robust polling-based log streaming.

**Tasks:**

- [ ] `pkg/api/pipelines.go` — `ListPipelines()`, `GetPipeline()`, `TriggerPipeline(branch/tag/commit)`, `StopPipeline()`, `ListSteps()`, `GetStepLog(stepUUID, rangeStart int64)` — byte-range log fetching
- [ ] `pkg/cmd/pipeline/list` — table: #, status (colored), branch, triggered by, duration, started; `--branch` filter; `--json`
- [ ] `pkg/cmd/pipeline/view` — show pipeline steps, each step's status and duration; `--step` flag to show specific step log
- [ ] `pkg/cmd/pipeline/run` — trigger on branch (default: current branch); `--branch`, `--tag`, `--commit` flags; prints pipeline URL and UUID
- [ ] `pkg/cmd/pipeline/cancel` — `POST /pipelines/{uuid}/stopPipeline`; confirm in TTY
- [ ] `pkg/cmd/pipeline/watch`:
  - If no UUID given: trigger new run on current branch, then watch it
  - Poll loop (default 10s, `--poll-interval` flag, minimum 5s with warning):
    1. `GET /pipelines/{uuid}` — check overall state
    2. For each step in `IN_PROGRESS`: `GET /steps/{step_uuid}/log` with `Range: bytes=N-`; print new bytes to stdout
    3. Rate limit awareness: count requests, warn at 80% of hourly limit (800/hr); slow poll automatically if nearing limit
  - On `COMPLETED`: print final status; exit 0 for `SUCCESSFUL`, exit 1 for `FAILED`/`ERROR`/`STOPPED`, exit 2 for watch errors
  - `--timeout` flag (default: no timeout) to abort watch after N minutes

---

### Phase 6: Issues & Snippets

**Goal:** Full `bb issue` and `bb snippet` coverage.

**Tasks:**

- [ ] `pkg/api/issues.go` — `ListIssues()`, `GetIssue()`, `CreateIssue()`, `UpdateIssue(status)`, `AddComment()`
- [ ] **`pkg/api/repos.go` (update):** Add `HasIssues(workspace, slug) bool` — used as guard in all issue commands
- [ ] `pkg/cmd/issue/` (all subcommands):
  - All issue commands: fetch repo metadata first; if `!has_issues` → error "Issues are not enabled for this repository. Enable them under Repository Settings → Features → Issues."
  - `list`: table: #, title, status, assignee, priority; filter: `--state`, `--assignee`, `--kind`
  - `view`: full issue with comments
  - `create`: `huh` form for title, description, assignee, kind (bug/enhancement/etc.), priority
  - `close`: `PUT` with `{"status": "resolved"}` — `--status` flag for other close states
  - `reopen`: `PUT` with `{"status": "open"}`
  - `comment`: `--body` or open `$EDITOR`
- [ ] `pkg/api/snippets.go` — `ListSnippets()`, `GetSnippet()`, `CreateSnippet()`, `UpdateSnippet()`, `DeleteSnippet()`
- [ ] `pkg/cmd/snippet/` (all subcommands):
  - `list`: table: ID, title, files, visibility, created; `--workspace` flag targets workspace snippets; default is authenticated user's snippets
  - `view`: show file list + content of first file; `--file` to show specific file
  - `create`: from stdin or `--file`; `--title`, `--private`/`--public`
  - `edit`: open in `$EDITOR`; `PUT` updated content
  - `delete`: confirm in TTY; `--force` in `--no-tty`
  - `clone`: `git clone` the snippet repository

---

### Phase 7: Distribution & Polish

**Goal:** Releasable binaries, package manager listings, install script, and project polish.

**Tasks:**

- [ ] Finalize `.goreleaser.yaml`:
  - 5 build targets: `darwin/amd64`, `darwin/arm64` → merged to universal binary, `linux/amd64`, `linux/arm64`, `windows/amd64`
  - `CGO_ENABLED=0` enforced
  - `ldflags`: version, commit, date, `OAuthClientID` from env
  - nfpm: `.deb` + `.rpm` with shell completion files
  - `homebrew_casks:` (not deprecated `brews:`)
  - `scoops:` for Windows Scoop
  - `winget:` with PR to `microsoft/winget-pkgs`
  - `release: draft: true` for human review before publishing
- [ ] `scripts/install.sh` — universal curl-pipe installer; detects OS/arch; downloads from GitHub Releases
- [ ] Set up Homebrew tap repository (`yourorg/homebrew-tap`) with initial structure
- [ ] Set up Scoop bucket repository (`yourorg/scoop-bucket`)
- [ ] Register Bitbucket OAuth consumer in Bitbucket workspace settings; store `client_id` in GitHub Actions secret `BB_OAUTH_CLIENT_ID`
- [ ] GoReleaser smoke test: `goreleaser release --snapshot --clean` — verify all 5 binaries build
- [ ] Shell completion registration instructions in README

---

## System-Wide Impact

### Interaction Graph

```
bb pr create
  └─> gitcontext.FromRemote() [git subprocess]
        └─> api.GetRepo() → checks fork, resolves destination
              └─> api.ListWorkspaceMembers() [for reviewer suggestions]
                    └─> huh.Form [TTY prompt]
                          └─> api.CreatePR() [POST /pullrequests]
                                └─> On 401: auth.RefreshToken() [exclusive file lock]
                                      └─> Retry original CreatePR()
```

### Error Propagation

- `401` from any API call → `auth.RefreshToken()` → retry once → if still 401 → `ErrSessionExpired`
- `429` → exponential backoff (2s, 4s, 8s, max 30s) → retry → if still 429 → surface rate limit error
- `404` on issue endpoint → check `has_issues` → if false, friendly error; if true, generic 404
- Subcommand errors propagate up through `cobra.Command.RunE` return values; `os.Exit` only in `main.go`

### State Lifecycle Risks

- **Token refresh race:** Two concurrent `bb` calls share `~/.config/bb/tokens.json`. File lock in `auth/refresh.go` prevents both from invalidating the session. Without the lock, the second refresh attempt uses an already-rotated (now invalid) refresh token → permanent logout.
- **`bb repo delete` is irreversible.** Double confirmation (interactive type-name check) guards against accidental deletion. In `--no-tty`, `--force` must be explicit.

### Integration Test Scenarios

1. `bb auth login --with-token` → `bb repo list` → `bb pr list` — full authenticated flow from token input to data retrieval
2. `bb pipeline watch` on a 5-minute pipeline — verify exit code 0 on success, exit code 1 on failure; verify logs print incrementally
3. `bb pr create` inside a forked repo directory — verify fork detection, correct destination repo
4. Two concurrent `bb repo list` calls at token expiry — verify only one refresh happens, both succeed
5. `bb issue list` on a repo with `has_issues: false` — verify friendly error, not 404

---

## Acceptance Criteria

### Functional

- [ ] `bb auth login` completes OAuth flow and stores credentials without manual token creation (R1)
- [ ] `bb auth login --with-token` accepts Bitbucket API Token via stdin (R2) — ❌ NOT app passwords
- [ ] All 65 requirements from origin document are implemented (R1–R65)
- [ ] Concurrent token refresh (two simultaneous `bb` invocations at token expiry) results in both succeeding — not one being permanently logged out
- [ ] `bb pipeline watch` exit code reflects pipeline result (0 = success, 1 = failure)
- [ ] `bb issue list` on repo with `has_issues: false` shows actionable friendly error
- [ ] `bb pr create` inside a fork repo correctly detects the upstream and prompts for destination

### Non-Functional

- [ ] All binaries produced with `CGO_ENABLED=0` — no glibc dependency; runs on Alpine Linux
- [ ] Core commands complete in <500ms excluding network I/O (R from success criteria)
- [ ] `bb --help` and all subcommand `--help` produce consistent, complete documentation
- [ ] `bb completion zsh | source /dev/stdin; bb <TAB>` produces correct completions

### Quality Gates

- [ ] `go test ./...` passes with no mocked API calls making real network requests
- [ ] `golangci-lint run` passes with zero errors
- [ ] GoReleaser snapshot build produces all 5 expected binary artifacts
- [ ] `bb version --json version,commit,buildDate` returns valid JSON

---

## Success Metrics

- A developer can go from zero to an authenticated session and merged PR using only `bb` commands, without touching the Bitbucket web UI
- CI/CD scripts can use `bb pipeline watch` to gate on pipeline results and parse output without tools beyond `jq`
- The binary installs via `brew install yourorg/tap/bb` on macOS in a single command
- All core commands run in under 500ms on a standard machine (excluding network latency)

---

## Dependencies & Prerequisites

- Go 1.22+ installed for development
- Bitbucket Cloud workspace with at least one OAuth consumer registered (for `BB_OAUTH_CLIENT_ID` secret)
- GitHub repository created for `bb` source code (for releases)
- Separate GitHub repos for Homebrew tap and Scoop bucket
- GoReleaser v2.x installed locally for testing

---

## Risk Analysis

| Risk | Severity | Mitigation |
|------|----------|-----------|
| App password sunset June 9, 2026 | High | Already mitigated — spec explicitly forbids app password auth; only API tokens supported |
| Rotating refresh token enforcement May 4, 2026 | High | File lock in `auth/refresh.go` prevents race; must ship before this date |
| PKCE not confirmed by Bitbucket | Medium | Auth Code + loopback works without PKCE; attempt PKCE first, retry without on error |
| Bitbucket rate limit (1,000 req/hr) | Medium | Pipeline watch default 10s interval; auto-slow on 80% usage; pagelen=100 to minimize calls |
| GoReleaser Homebrew `brews:` deprecation | Low | Use `homebrew_casks:` from day one; not a migration issue |
| `go-survey` is archived | Low | Using `charmbracelet/huh` from day one; not a migration issue |

---

## Deferred to Planning → Now Resolved

| Question | Resolution |
|----------|-----------|
| Pipeline log streaming vs. polling | Polling with byte-range requests; no SSE endpoint exists in Bitbucket API |
| PKCE support | Attempt PKCE; if Bitbucket rejects `code_challenge`, retry without — not documented as supported |
| `.bb` per-repo config format | Same YAML schema as global config; Viper `AddConfigPath(".")` picks up `.bb.yaml` in CWD |
| GoReleaser vs. Makefile | GoReleaser for all release artifacts; Makefile wraps `goreleaser release --snapshot` for local testing |
| Default workspace determination | Resolution order: `--workspace` flag → `BITBUCKET_WORKSPACE` env → git remote inference → config default → error |

---

## Outstanding Questions

### Resolve Before Work

- [Affects Phase 1, all phases] What is the actual GitHub org/username where `bb` will be hosted? The Go module path (`github.com/yourorg/bb`) and all import paths depend on this. Recommend resolving before writing any Go code to avoid a mass-rename later.

### Deferred to Implementation

- [Affects Phase 2] Does Bitbucket Cloud actually accept `code_challenge` + `code_challenge_method=S256` in the authorization request? The official docs do not document PKCE. Test empirically during Phase 2 OAuth implementation.
- [Affects Phase 7] Winget `publisher` and `package_identifier` format for the official `microsoft/winget-pkgs` submission — verify current submission requirements at time of release.

---

## Future Considerations (Post-V1)

- **OS keychain integration** (`99designs/keyring`) for more secure token storage
- **Webhook management** — `bb webhook create/list/delete`
- **`bb browse`** as a top-level command (like `gh browse`) opening the current repo in browser based on git context
- **SSH key management** — `bb ssh-key add/list/delete`
- **Pipeline variables management** — `bb pipeline variable set/list`
- **Rich TUI dashboard** using Bubbletea for interactive PR/pipeline views

---

## Sources & References

### Origin

- **Origin document:** [docs/brainstorms/2026-03-30-bb-cli-requirements.md](../brainstorms/2026-03-30-bb-cli-requirements.md)
  - Key decisions carried forward: Go language, `bb` binary name, OAuth 2.0 + API Token auth (app passwords explicitly excluded), bundled OAuth consumer credentials, interactive prompts with `--no-tty`, `--json`+`--jq` scripting, GitHub Releases + Homebrew/Scoop/Winget distribution

### Internal References

- Best Makefile pattern: `/Users/rc/Boomi/Bedrock-Jira-Analyser/Makefile`
- Best CLAUDE.md pattern: `/Users/rc/Boomi/Bedrock-Jira-Analyser/CLAUDE.md`
- Existing Bitbucket-connected tool: `/Users/rc/Boomi/relutil/pom.xml` (workspace slug: `boomii`)

### External References

- [Bitbucket Cloud REST API v2.0](https://developer.atlassian.com/cloud/bitbucket/rest/intro/)
- [Bitbucket OAuth 2.0 docs](https://developer.atlassian.com/cloud/bitbucket/oauth-2/)
- [Bitbucket App Password Deprecation](https://www.atlassian.com/blog/bitbucket/bitbucket-cloud-transitions-to-api-tokens-enhancing-security-with-app-password-deprecation) — sunset June 9, 2026
- [GitHub CLI project layout](https://github.com/cli/cli/blob/trunk/docs/project-layout.md) — reference architecture
- [GoReleaser v2 docs](https://goreleaser.com/) — `homebrew_casks:`, nfpm, Winget
- [charmbracelet/huh](https://github.com/charmbracelet/huh) — replaces archived go-survey
- [itchyny/gojq](https://github.com/itchyny/gojq) — pure-Go jq implementation

### Related Work

- [RFC 8252 — OAuth 2.0 for Native Apps](https://datatracker.ietf.org/doc/html/rfc8252) — loopback redirect port flexibility (§7.3)
