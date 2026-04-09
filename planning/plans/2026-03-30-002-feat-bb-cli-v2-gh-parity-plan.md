---
title: "feat: bb CLI v2 — gh parity additions"
type: feat
status: active
date: 2026-03-30
---

# bb CLI v2 — gh Parity & Bitbucket-Native Additions

## Overview

This plan closes the gap between `bb` and `gh` by adding 10 new command groups derived from a systematic comparison of `gh --help` against the current `bb` command surface. It also adds three Bitbucket-native commands (`bb variable`, `bb deploy-key`, `bb runner`) that have no `gh` equivalent but are critical for Bitbucket Cloud teams.

**Current command groups (v1):**
`auth`, `workspace`, `repo`, `branch`, `pr`, `pipeline`, `issue`, `snippet`, `version`, `completion`

**New command groups (this plan):**
`api`, `browse`, `ssh-key`, `gpg-key`, `alias`, `variable`, `deploy-key`, `webhook`, `search`, `status`, `runner`

---

## Gap Analysis: gh vs bb

### What gh has that bb is missing

| gh command                  | bb equivalent                    | Status                                                              | Plan                                   |
| --------------------------- | -------------------------------- | ------------------------------------------------------------------- | -------------------------------------- |
| `gh api`                    | —                                | ❌ Missing                                                          | `bb api` — raw authenticated API proxy |
| `gh browse`                 | `bb repo browse`, `bb pr browse` | ⚠️ Partial                                                          | `bb browse` — top-level context-aware  |
| `gh ssh-key`                | —                                | ❌ Missing                                                          | `bb ssh-key list/add/delete`           |
| `gh gpg-key`                | —                                | ❌ Missing                                                          | `bb gpg-key list/add/delete`           |
| `gh alias`                  | —                                | ❌ Missing                                                          | `bb alias set/list/delete`             |
| `gh search`                 | —                                | ❌ Missing                                                          | `bb search code/repos`                 |
| `gh status`                 | —                                | ❌ Missing                                                          | `bb status` — cross-repo dashboard     |
| `gh secret` + `gh variable` | —                                | ❌ Missing                                                          | `bb variable` — pipeline variables     |
| `gh release`                | —                                | ✅ N/A — no Bitbucket releases                                      | —                                      |
| `gh codespace`              | —                                | ✅ N/A — no Bitbucket codespaces                                    | —                                      |
| `gh project`                | —                                | ✅ N/A — GitHub Projects only                                       | —                                      |
| `gh cache/run/workflow`     | `bb pipeline`                    | ✅ Covered                                                          | —                                      |
| `gh ruleset`                | `bb branch protect`              | ✅ Covered                                                          | —                                      |
| `gh label`                  | —                                | ❌ Skip — Bitbucket components/milestones are **read-only via API** | —                                      |

### Bitbucket-native additions (no gh equivalent)

| Command         | Description                                             |
| --------------- | ------------------------------------------------------- |
| `bb variable`   | Pipeline variables — workspace, repo, deployment scopes |
| `bb deploy-key` | Repo-level SSH deploy keys                              |
| `bb runner`     | Bitbucket Pipelines Runners (self-hosted agents)        |

---

## Problem Statement

`bb` v1 covers the most common developer workflows. But teams using Bitbucket Cloud daily need:

1. **A raw API escape hatch** (`bb api`) for everything not yet wrapped
2. **Pipeline variables management** (`bb variable`) — setting CI/CD secrets and env vars without the web UI
3. **A cross-repo status dashboard** (`bb status`) — `gh status` is one of the most-used `gh` commands
4. **SSH/GPG key management** — devs add keys from the terminal, not the browser
5. **Webhook management** (`bb webhook`) — critical for CI/CD integrations
6. **Bitbucket Runners management** (`bb runner`) — increasingly common for enterprise teams with self-hosted agents

---

## Technical Approach

All new commands follow the established pattern from v1:

- Factory DI via `*cmdutil.Factory`
- New API methods in `pkg/api/<domain>.go`
- Flat command package layout: `pkg/cmd/<noun>/<noun>.go` + `pkg/cmd/<noun>/<verb>.go`
- `--json`/`--jq`/`--limit` on all list/view commands
- Registered in `pkg/cmd/root/root.go`

Key API constraints discovered during research:

| Feature                      | Constraint                                                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| SSH keys                     | No `/user/ssh-keys` shortcut — must fetch account UUID from `GET /user` first                                            |
| Pipeline variables (secured) | Value is **write-only** — always returns empty string on GET; display as `[secured]`                                     |
| Pipeline variables URL       | Workspace: `pipelines-config` (hyphen); Repo: `pipelines_config` (underscore) — inconsistent, must be handled separately |
| Webhooks                     | Secret is **never returned** on GET — display `[set]` / `[not set]`; only shown once on create                           |
| Runners                      | OAuth client ID+secret returned **only on creation** — must show clearly and warn user to save them                      |
| Code search                  | Must be **enabled per workspace** in Bitbucket UI; returns 404 if not enabled — surface friendly error                   |
| Issue labels                 | **Read-only via API** — components/milestones cannot be created/modified via REST; skip write commands                   |
| GPG key path param           | Fingerprint string (case-insensitive), not a UUID                                                                        |
| Deploy key ID                | Integer, not UUID                                                                                                        |

---

## Implementation Phases

### Phase 8A: Foundation Additions (Low effort, high value)

**Goal:** Ship the most-requested features with minimal risk. All five commands are leaf or simple group commands with straightforward API calls.

#### 1. `bb api` — Raw API Proxy

The single highest-leverage addition. Gives users access to every Bitbucket endpoint `bb` doesn't wrap.

```
bb api /user
bb api /repositories/myworkspace --jq '.[].slug'
bb api /repositories/ws/repo --method POST --field name=val
bb api /repositories/ws/repo --paginate
```

**Files:**

- `pkg/cmd/api/api.go` — single leaf command (no subcommands)
- `pkg/api/client.go` — add `func (c *Client) RawRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error)`

**Flags:**
| Flag | Description |
|------|-------------|
| `--method` / `-X` | HTTP method (default: GET) |
| `--field` / `-F` | JSON body field `key=value` (implies POST if method not set) |
| `--header` / `-H` | Additional request header `name:value` |
| `--paginate` | Follow all `next` pages, print array of all results |
| `--jq` | jq filter applied to response |
| `--input` | Read request body from file (`@filepath`) or stdin (`-`) |

**Behavior:**

- Accepts a path relative to `https://api.bitbucket.org/2.0` (strips the base if a full URL is given)
- Streams response body to stdout
- `--paginate` collects all pages and outputs a single JSON array
- Non-2xx → exits 1, prints error to stderr
- `--jq` implies JSON mode (no pretty-printing needed, just pipe through jq)

**Registration:** direct on root — `cmd.AddCommand(apicmd.NewCmdAPI(f))`

---

#### 2. `bb browse` — Top-Level Context-Aware Browser

Currently `bb repo browse` and `bb pr browse` exist but there's no top-level `bb browse` like `gh browse`.

```
bb browse                    # open current repo
bb browse --branch feat/x    # open branch view
bb browse --pr 42            # open PR #42
bb browse --issue 7          # open issue #7
bb browse --pipeline <uuid>  # open pipeline run
bb browse --settings         # open repo settings page
bb browse --commits          # open commits view
```

**Files:**

- `pkg/cmd/browse/browse.go` — top-level leaf command

**Logic:** Infers workspace/slug from `gitcontext.FromRemote()`. Constructs Bitbucket URL based on flags. Falls back to opening the repo root if no flags given.

**URL patterns:**

```
Repo:      https://bitbucket.org/{ws}/{slug}
Branch:    https://bitbucket.org/{ws}/{slug}/src/{branch}
PR:        https://bitbucket.org/{ws}/{slug}/pull-requests/{id}
Issue:     https://bitbucket.org/{ws}/{slug}/issues/{id}
Pipeline:  https://bitbucket.org/{ws}/{slug}/pipelines/results/{uuid}
Settings:  https://bitbucket.org/{ws}/{slug}/admin
Commits:   https://bitbucket.org/{ws}/{slug}/commits
```

---

#### 3. `bb ssh-key` — SSH Key Management

```
bb ssh-key list
bb ssh-key add --key "$(cat ~/.ssh/id_ed25519.pub)" --label "My laptop"
bb ssh-key add --key-file ~/.ssh/id_ed25519.pub
bb ssh-key delete <fingerprint>
```

**Files:**

- `pkg/api/ssh_keys.go` — `SSHKey` type, `ListSSHKeys`, `AddSSHKey`, `DeleteSSHKey`
- `pkg/cmd/ssh_key/ssh_key.go` — group command
- `pkg/cmd/ssh_key/list.go`, `add.go`, `delete.go`

**API:** `/users/{account_id}/ssh-keys` — account ID fetched via `client.GetUser()` on first call.

**Key implementation detail:** No `/user/ssh-keys` shortcut exists. Must call `GET /user` to get the Atlassian account ID, then use `/users/{account_id}/ssh-keys`.

**Table columns (list):** UUID (short), Fingerprint, Comment, Created

---

#### 4. `bb gpg-key` — GPG Key Management

```
bb gpg-key list
bb gpg-key add --key "$(gpg --armor --export KEYID)"
bb gpg-key add --key-file my_key.asc
bb gpg-key delete <fingerprint>
```

**Files:**

- `pkg/api/gpg_keys.go` — `GPGKey` type, `ListGPGKeys`, `AddGPGKey`, `DeleteGPGKey`
- `pkg/cmd/gpg_key/gpg_key.go`, `list.go`, `add.go`, `delete.go`

**API:** `/users/{account_id}/gpg-keys` — same pattern as SSH keys (need account ID from `GET /user`).

**Key detail:** Path parameter for delete is the **fingerprint** (uppercase hex), not a UUID. The `key` body field is an armored PGP public key block. Use `?fields=%2Bkey` to retrieve the key material in view commands.

**Table columns (list):** Fingerprint, Name, Key ID, Created, Expires

---

#### 5. `bb alias` — Command Shortcuts

```
bb alias set co 'pr checkout'
bb alias set myrepos 'repo list --workspace boomii --limit 0'
bb alias list
bb alias delete co
```

**Files:**

- `pkg/cmd/alias/alias.go`, `set.go`, `list.go`, `delete.go`
- Pure config manipulation — no API client needed

**Storage:** `~/.config/bb/config.yaml` under an `aliases:` key via `f.Config()`.

**Execution:** In `root.go` `PersistentPreRunE`, check if the first argument matches an alias and expand it before Cobra parses the command tree. This requires hooking into command lookup.

**Alias expansion pattern** (from gh): before `cmd.Execute()` in `main.go`, run alias expansion on `os.Args[1:]` — replace known alias names with their expansion, then pass modified args to Cobra.

---

### Phase 8B: CI/CD Power Commands

**Goal:** Pipeline variables, deploy keys, and webhooks — the most requested Bitbucket-specific additions for CI/CD teams.

#### 6. `bb variable` — Pipeline Variables

Bitbucket's equivalent of `gh secret` + `gh variable`. Three scopes: workspace, repo, deployment environment.

```
# Workspace variables
bb variable list --workspace boomii
bb variable set MY_TOKEN secretvalue --workspace boomii
bb variable set MY_TOKEN secretvalue --workspace boomii --secured
bb variable delete MY_TOKEN --workspace boomii

# Repo variables (default scope when inside a repo)
bb variable list
bb variable set AWS_KEY value123 --secured
bb variable delete AWS_KEY

# Deployment environment variables
bb variable list --env production
bb variable set DB_URL value --env production --secured
bb variable delete DB_URL --env production

# List deployment environments
bb variable env list
bb variable env create production
```

**Files:**

- `pkg/api/variables.go` — `PipelineVariable`, `DeploymentEnvironment` types; workspace/repo/env CRUD methods
- `pkg/cmd/variable/variable.go`, `list.go`, `set.go`, `delete.go`, `env/` subgroup

**Critical API quirks:**

- Workspace URL: `/workspaces/{ws}/pipelines-config/variables/` (hyphen)
- Repo URL: `/repositories/{ws}/{slug}/pipelines_config/variables/` (underscore)
- Secured variables always return `value: ""` — display as `[secured]`
- `variable_uuid` in paths must include curly braces
- Scope resolution: `--env` → deployment vars; `--workspace` → workspace vars; default → repo vars

**Table columns (list):** Key, Value (or `[secured]`), Secured (✓/✗), UUID

---

#### 7. `bb deploy-key` — Repo Deploy Keys

```
bb deploy-key list
bb deploy-key add --key "$(cat ~/.ssh/deploy_key.pub)" --label "CI server"
bb deploy-key add --key-file deploy_key.pub --label "CI server"
bb deploy-key delete <id>
bb deploy-key view <id>
```

**Files:**

- `pkg/api/deploy_keys.go` — `DeployKey` type, `ListDeployKeys`, `AddDeployKey`, `DeleteDeployKey`
- `pkg/cmd/deploy_key/deploy_key.go`, `list.go`, `add.go`, `delete.go`, `view.go`

**Key detail:** `key_id` is an **integer** in URL paths (not UUID). Returned as `id` in JSON response.

**Table columns:** ID, Label, Fingerprint, Comment, Added On, Last Used

---

#### 8. `bb webhook` — Webhook Management

```
bb webhook list
bb webhook list --workspace boomii   # workspace-level hooks
bb webhook create --url https://ci.example.com/hook --events repo:push,pullrequest:created
bb webhook create --url https://... --events repo:push --secret mysecret
bb webhook view <uid>
bb webhook delete <uid>
bb webhook events                    # list all available event types
```

**Files:**

- `pkg/api/webhooks.go` — `WebhookSubscription` type, CRUD methods for repo + workspace scopes
- `pkg/cmd/webhook/webhook.go`, `list.go`, `create.go`, `view.go`, `delete.go`, `events.go`

**Key details:**

- Secret is **never returned** on GET — `secret_set: true/false` only; display as `[set]` / `[not set]`
- To leave secret unchanged on update: omit the field. To clear: send `null`.
- `--scope workspace` flag for workspace-level hooks (vs default repo-level)
- `bb webhook events` → `GET /hook_events/repository` + `GET /hook_events/workspace` to list all event types

**Table columns (list):** UUID, URL, Events (count), Active, Secret Set, Created

---

### Phase 8C: Discovery & Advanced

**Goal:** Search, status dashboard, and runner management.

#### 9. `bb search` — Code and Repo Search

```
bb search code "function authenticate" --workspace boomii
bb search code "TODO" --workspace boomii --limit 20
bb search repos "cli" --workspace boomii
bb search repos --language go --workspace boomii
```

**Files:**

- `pkg/api/search.go` — `CodeSearchResult`, `SearchResultPage` types; `SearchCode`, `SearchRepos` methods
- `pkg/cmd/search/search.go`, `code.go`, `repos.go`

**API details:**

- Code: `GET /workspaces/{ws}/search/code?search_query=...`
- Repos: uses `client.ListRepos` with `q=name ~ "query"` filter (no separate search endpoint)
- Code search requires `search_query` param (NOT `q=`)
- Must surface friendly error if search not enabled: "Code search is not enabled for this workspace. Enable it at https://bitbucket.org/{workspace}/search"

**Code search output format:**

```
pkg/auth/oauth.go
  line 44:  clientID := version.OAuthClientID
  line 62:  ClientID: clientID,

pkg/auth/refresh.go
  line 66:  clientID := version.OAuthClientID
```

---

#### 10. `bb status` — Cross-Repo Dashboard

```
bb status
bb status --workspace boomii
```

**Files:**

- `pkg/cmd/status/status.go` — top-level leaf command (no subcommands)

**What it shows:**

1. **PRs authored by me** that are OPEN, grouped by repo
2. **PRs waiting for my review** (user is a participant with REVIEWER role and hasn't approved)
3. **Issues assigned to me** (OPEN)
4. **Recent pipeline failures** (FAILED/ERROR in last 24h across repos)

**Implementation approach:**

1. `GET /user` → get username
2. `GET /workspaces` → list workspace slugs (or use `--workspace` flag)
3. For each workspace: `GET /repositories/{ws}` with `role=member&pagelen=100` → get repo slugs
4. For each repo (parallel with goroutines):
   - PRs: `GET /repositories/{ws}/{slug}/pullrequests?q=state="OPEN" AND author.nickname="{me}"&pagelen=10`
   - Review requests: `GET /repositories/{ws}/{slug}/pullrequests?q=state="OPEN" AND reviewers.nickname="{me}"&pagelen=10`
   - Issues: `GET /repositories/{ws}/{slug}/issues?q=status="open" AND assignee.nickname="{me}"&pagelen=10`
   - Pipelines: `GET /repositories/{ws}/{slug}/pipelines/?sort=-created_on&pagelen=5` → filter FAILED in last 24h
5. Aggregate and display grouped by category

**Output format (example):**

```
✓ Assigned to you
  Issues (2)
  · boomii/my-service #12 — Login page broken
  · boomii/api-gateway #8  — Timeout on health check

● Waiting for your review (1)
  · boomii/my-service PR #45 — Add rate limiting

● Your open PRs (1)
  · boomii/api-gateway PR #23 — Fix deployment pipeline

⚠ Recent pipeline failures (1)
  · boomii/my-service — build #142 (main) failed 2h ago
```

**Rate limit consideration:** Fan-out across many repos can hit Bitbucket's 1,000 req/hr limit. Implement a concurrency cap (max 5 goroutines) and `--repos` flag to limit which repos are checked.

---

#### 11. `bb runner` — Bitbucket Pipelines Runners

```
bb runner list
bb runner list --workspace boomii
bb runner list --repo-scope               # repo-scoped runners
bb runner view <uuid>
bb runner create --name "my-linux-runner" --label linux --label large
bb runner delete <uuid>
```

**Files:**

- `pkg/api/runners.go` — `PipelineRunner`, `RunnerState` types; workspace + repo CRUD
- `pkg/cmd/runner/runner.go`, `list.go`, `view.go`, `create.go`, `delete.go`

**Key detail:** On runner creation, the response includes `oauth_client.id` and `oauth_client.secret` — **these are only returned once** and are needed to configure the runner binary on the host machine. The create output must display them prominently with a warning:

```
✓ Created runner: my-linux-runner
  UUID: {abc123}

⚠ Save these credentials — they are shown only once:
  Client ID:     abc123...
  Client Secret: xyz789...

Install and configure the runner:
  https://support.atlassian.com/bitbucket-cloud/docs/runners/
```

**Table columns (list):** UUID, Name, Labels, State (colored), Created

---

## Acceptance Criteria

### Phase 8A

- [ ] `bb api /user` prints the authenticated user JSON
- [ ] `bb api /repositories/ws/slug --paginate` fetches all pages and prints a single JSON array
- [ ] `bb api /repositories/ws/slug --method POST --field name=val` makes a POST with JSON body
- [ ] `bb browse` (no flags, inside a cloned Bitbucket repo) opens the repo URL in the system browser
- [ ] `bb browse --pr 42` opens the PR URL
- [ ] `bb ssh-key list` shows all SSH keys for the authenticated user
- [ ] `bb ssh-key add --key-file ~/.ssh/id_ed25519.pub` adds a key
- [ ] `bb ssh-key delete <fingerprint>` removes a key
- [ ] `bb gpg-key list/add/delete` work identically to ssh-key with GPG keys
- [ ] `bb alias set foo 'repo list'` and `bb foo` runs the aliased command
- [ ] `bb alias list` shows all aliases
- [ ] `bb alias delete foo` removes the alias

### Phase 8B

- [ ] `bb variable list` shows workspace/repo variables with secured values shown as `[secured]`
- [ ] `bb variable set KEY val --secured` creates a secured variable
- [ ] `bb variable delete KEY` removes a variable
- [ ] `bb variable env list` shows deployment environments
- [ ] `bb variable list --env production` shows environment-scoped variables
- [ ] `bb deploy-key list/add/delete` work with integer key IDs
- [ ] `bb webhook list/create/view/delete` work; secret never shown in GET responses
- [ ] `bb webhook create --secret mysecret --url ...` shows secret in create response only
- [ ] `bb webhook events` lists all available event types

### Phase 8C

- [ ] `bb search code "func main"` returns colored code matches with file:line context
- [ ] `bb search code` on a workspace where search is disabled shows a friendly enable-it message
- [ ] `bb status` shows aggregated PRs, issues, pipeline failures across workspace repos
- [ ] `bb runner list` shows runners with state colored (green=ONLINE, red=OFFLINE)
- [ ] `bb runner create` shows OAuth credentials and install instructions prominently (once only)
- [ ] `bb runner delete <uuid>` removes the runner

### All new commands

- [ ] All list commands support `--json fields` and `--jq expr`
- [ ] All list commands support `--limit N` (with `0 = all`)
- [ ] All destructive commands support `--force` in `--no-tty` mode
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` returns 0 issues

---

## Success Metrics

- `bb api` enables any Bitbucket API operation without leaving the terminal
- `bb status` becomes the morning dashboard for Bitbucket teams (mirrors `gh status` usage pattern)
- `bb variable` eliminates the need to open the Bitbucket web UI for CI/CD secret management
- All existing tests continue to pass; new commands covered by mock HTTP tests

---

## Dependencies & Prerequisites

- Existing `pkg/api/client.go` is the foundation — `bb api` adds a `RawRequest` method
- `pkg/auth/store.go` already handles token refresh — all new commands inherit authentication
- `pkg/output/table.go` and `pkg/output/json.go` are reused unchanged
- `pkg/browser/browser.go` already exists — `bb browse` reuses it
- `gitcontext.FromRemote()` provides repo context for all repo-scoped commands

---

## Risk Analysis

| Risk                                                       | Severity | Mitigation                                                                               |
| ---------------------------------------------------------- | -------- | ---------------------------------------------------------------------------------------- |
| Code search not enabled per workspace                      | Medium   | Detect 404 response and show URL to enable it                                            |
| Pipeline variable URL inconsistency (hyphen vs underscore) | Medium   | Explicit constants for each scope's base path                                            |
| Secured variable value exposure                            | High     | Never attempt to display value for `secured: true` vars; show `[secured]`                |
| Runner OAuth creds lost if not saved                       | High     | Print prominently, mention once-only nature, link to docs                                |
| `bb status` rate limit on large workspaces                 | Medium   | Cap goroutine concurrency at 5; add `--repos` flag to limit scope                        |
| Alias execution hijacking                                  | Medium   | Strict allowlist: aliases can only expand to known `bb` subcommands or flag combinations |
| GPG fingerprint case sensitivity                           | Low      | Always uppercase before using as path parameter                                          |
| `bb api --paginate` with large result sets                 | Low      | Stream output rather than buffering all pages in memory                                  |

---

## Future Considerations (v3 and beyond)

- `bb pr checks` — show CI status checks on a PR (maps to Bitbucket commit statuses API)
- `bb repo archive/unarchive` — Bitbucket repo archiving (if API supports it)
- `bb project` — Bitbucket Projects (repo grouping within workspaces)
- `bb config edit` — open config in `$EDITOR` (like `gh config edit`)
- `bb extension` — a plugin system for custom commands (complex; low priority)
- `bb label` — if Bitbucket ever adds write access to issue components/milestones API

---

## Sources & References

### Internal

- `pkg/api/client.go` — `Client.Get/Post/Put/Delete/GetRaw` methods to reuse
- `pkg/cmd/pr/create.go` — `huh` interactive form pattern
- `pkg/cmd/pipeline/watch.go` — goroutine + fan-out pattern for `bb status`
- `pkg/cmd/repo/clone.go` — `validateCloneURL` pattern for input validation
- `pkg/browser/browser.go` — browser.Open() for `bb browse`
- `pkg/cmd/root/root.go` — command registration pattern

### Bitbucket API

- SSH Keys: `/users/{account_id}/ssh-keys` — GET/POST/PUT/DELETE
- GPG Keys: `/users/{account_id}/gpg-keys` — GET/POST/DELETE (fingerprint as path key)
- Pipeline Variables (workspace): `/workspaces/{ws}/pipelines-config/variables/` (note: hyphen)
- Pipeline Variables (repo): `/repositories/{ws}/{slug}/pipelines_config/variables/` (note: underscore)
- Deploy Keys: `/repositories/{ws}/{slug}/deploy-keys` — key_id is integer, not UUID
- Webhooks (repo): `/repositories/{ws}/{slug}/hooks` — secret write-only after create
- Webhooks (workspace): `/workspaces/{ws}/hooks`
- Code Search: `/workspaces/{ws}/search/code?search_query=...`
- Runners (workspace): `/workspaces/{ws}/pipelines-config/runners/`
- Runners (repo): `/repositories/{ws}/{slug}/pipelines-config/runners/`

## Next Steps

→ `/ce:work docs/plans/2026-03-30-002-feat-bb-cli-v2-gh-parity-plan.md`

Implement in order: Phase 8A (api, browse, ssh-key, gpg-key, alias) → Phase 8B (variable, deploy-key, webhook) → Phase 8C (search, status, runner).
