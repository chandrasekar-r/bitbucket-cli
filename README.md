# bb — Bitbucket Cloud CLI

`bb` is a fast, cross-platform command-line tool for Bitbucket Cloud, modeled on GitHub's [`gh`](https://cli.github.com). Manage pull requests, repositories, pipelines, branches, issues, and more — without leaving your terminal.

```
bb pr create --title "Fix login bug" --base main
bb pipeline watch --notify
bb repo clone myworkspace/my-service
bb status
```

## Installation

### macOS — Homebrew

```bash
brew install chandrasekar-r/tap/bb
```

### Windows — Scoop

```powershell
scoop bucket add chandrasekar-r https://github.com/chandrasekar-r/scoop-bucket
scoop install bb
```

### Linux — apt / rpm

```bash
# Debian/Ubuntu
curl -Lo /tmp/bb.deb https://github.com/chandrasekar-r/bitbucket-cli/releases/latest/download/bb_linux_amd64.deb
sudo dpkg -i /tmp/bb.deb

# RHEL/Fedora
sudo rpm -i https://github.com/chandrasekar-r/bitbucket-cli/releases/latest/download/bb_linux_amd64.rpm
```

### Universal — curl installer

```bash
curl -fsSL https://raw.githubusercontent.com/chandrasekar-r/bitbucket-cli/main/scripts/install.sh | sh
```

### From source

```bash
go install github.com/chandrasekar-r/bitbucket-cli/cmd/bb@latest
```

---

## Authentication

### Browser OAuth (recommended)

```bash
bb auth login
```

Opens your browser for Bitbucket OAuth 2.0. Credentials are stored in `~/.config/bb/tokens.json`.

### API Token (CI/headless)

Create a Bitbucket API token at **Bitbucket → Settings → Security → API tokens**, then:

```bash
echo "$BB_TOKEN" | bb auth login --with-token --username myusername
```

Or use environment variables — no `auth login` needed:

```bash
export BITBUCKET_USERNAME=myusername
export BITBUCKET_TOKEN=myapitoken
bb repo list  # uses env vars automatically
```

> **Note:** App passwords are deprecated as of June 2026. Use Bitbucket API tokens.

### Multi-workspace

```bash
bb workspace list          # list all accessible workspaces
bb workspace use myteam    # set default workspace
```

---

## Quick Start

```bash
# Authenticate
bb auth login

# Clone a repo
bb repo clone myworkspace/my-service

# Create a PR from the current branch (description auto-generated from commits)
cd my-service
git checkout -b feat/my-feature
git push -u origin feat/my-feature
bb pr create

# Watch a pipeline and get a desktop notification when it finishes
bb pipeline watch --notify

# See your open PRs and pending reviews across all repos
bb status

# Tidy up merged local branches
bb branch tidy
```

---

## Command Reference

### Authentication

| Command          | Description                                |
| ---------------- | ------------------------------------------ |
| `bb auth login`  | Log in via browser OAuth or `--with-token` |
| `bb auth logout` | Remove stored credentials                  |
| `bb auth status` | Show current authentication state          |
| `bb auth token`  | Print the active access token              |

### Workspaces

| Command                    | Description                |
| -------------------------- | -------------------------- |
| `bb workspace list`        | List accessible workspaces |
| `bb workspace use <slug>`  | Set default workspace      |
| `bb workspace view [slug]` | Show workspace details     |

### Repositories

| Command                                 | Description                        |
| --------------------------------------- | ---------------------------------- |
| `bb repo list`                          | List repos in the active workspace |
| `bb repo view [workspace/repo]`         | Show repo details and clone URLs   |
| `bb repo create [name]`                 | Create a new repository            |
| `bb repo clone <workspace/repo>`        | Clone a repository                 |
| `bb repo fork <workspace/repo>`         | Fork a repository                  |
| `bb repo delete <workspace/repo>`       | Delete a repository                |
| `bb repo rename <workspace/repo> <new>` | Rename a repository                |
| `bb repo browse [workspace/repo]`       | Open in browser                    |

### Branches

| Command                        | Description                                               |
| ------------------------------ | --------------------------------------------------------- |
| `bb branch list`               | List branches                                             |
| `bb branch create <name>`      | Create a branch (`--from` to specify source)              |
| `bb branch delete [name]`      | Delete a branch (picker when name omitted)                |
| `bb branch rename [old] <new>` | Rename a branch (picker when old name omitted)            |
| `bb branch protect <name>`     | Enable branch restrictions                                |
| `bb branch tidy`               | Delete local branches whose PRs are merged or declined    |

`bb branch tidy` accepts `--dry-run` to preview and `--force` to skip confirmation.

### Pull Requests

| Command                   | Description                                                |
| ------------------------- | ---------------------------------------------------------- |
| `bb pr list`              | List open PRs (`--state` for MERGED/DECLINED)              |
| `bb pr view [number]`     | Show PR details (picker when number omitted)               |
| `bb pr create`            | Create a PR — description auto-generated from commits      |
| `bb pr edit [number]`     | Edit title, description, base branch, or reviewers         |
| `bb pr merge [number]`    | Merge a PR (`--strategy` merge_commit/squash/fast_forward) |
| `bb pr approve [number]`  | Approve a PR                                               |
| `bb pr decline [number]`  | Decline a PR                                               |
| `bb pr checkout [number]` | Check out PR source branch locally                         |
| `bb pr comment [number]`  | Add a comment (`--body` or opens `$EDITOR`)                |
| `bb pr diff [number]`     | Show unified diff                                          |
| `bb pr browse [number]`   | Open in browser                                            |

Commands that take `[number]` show an interactive picker when run without an argument.

#### Smart PR creation

`bb pr create` automatically pre-fills the description from your commit log, a `.bitbucket/PULL_REQUEST_TEMPLATE.md` file if present, and appends a `Related:` link if your branch name contains a Jira key (e.g. `feat/PROJ-123-my-feature`).

#### Editing a PR

```bash
# Flags — applies changes directly
bb pr edit 42 --title "New title" --base develop --add-reviewer alice

# No flags — opens an interactive form pre-filled with current values
bb pr edit 42
```

### Pipelines

| Command                     | Description                                              |
| --------------------------- | -------------------------------------------------------- |
| `bb pipeline list`          | List recent pipeline runs (`--branch` filter)            |
| `bb pipeline view <uuid>`   | Show pipeline details and step status                    |
| `bb pipeline run`           | Trigger a pipeline (default: current branch)             |
| `bb pipeline cancel <uuid>` | Cancel a running pipeline                                |
| `bb pipeline watch [uuid]`  | Follow logs in real time; `--notify` for desktop alert   |

`bb pipeline watch` exits with the pipeline's result code:

- `0` — SUCCESSFUL
- `1` — FAILED / ERROR / STOPPED
- `2` — watch error (network, timeout)

The `--notify` flag fires a native desktop notification (macOS, Linux, Windows) when the pipeline completes, so you can switch tasks while it runs.

### Issues

| Command                     | Description                            |
| --------------------------- | -------------------------------------- |
| `bb issue list`             | List open issues (`--state` filter)    |
| `bb issue view <number>`    | Show issue details                     |
| `bb issue create`           | Create an issue (interactive or flags) |
| `bb issue close <number>`   | Close an issue (`--status`)            |
| `bb issue reopen <number>`  | Reopen a closed issue                  |
| `bb issue comment <number>` | Add a comment                          |

Issues must be enabled in Repository Settings → Features → Issues.

### Snippets

| Command                  | Description                     |
| ------------------------ | ------------------------------- |
| `bb snippet list`        | List your snippets              |
| `bb snippet view <id>`   | Show snippet metadata and files |
| `bb snippet create`      | Create from stdin or `--file`   |
| `bb snippet edit <id>`   | Edit in `$EDITOR`               |
| `bb snippet delete <id>` | Delete a snippet                |
| `bb snippet clone <id>`  | Clone snippet as a git repo     |

### Status dashboard

```bash
bb status
```

Shows a personal cross-repo summary: your open pull requests (with approval counts) and PRs awaiting your review — across every repo in the workspace where you're a contributor.

```bash
bb status --json my_prs,review_prs   # machine-readable output
```

### Raw API access

```bash
bb api <endpoint>
```

Makes an authenticated request to any Bitbucket API endpoint. The placeholders `{workspace}` and `{repo}` are replaced from git context automatically.

```bash
# GET current user
bb api /user

# GET all repos (auto-paginate)
bb api /repositories/{workspace} --paginate

# POST — create an issue
bb api /repositories/{workspace}/{repo}/issues \
  -f title="Bug report" -f kind="bug" -f priority="major"

# DELETE — close a pipeline step
bb api /repositories/{workspace}/{repo}/pipelines/UUID -X DELETE

# Filter output with jq
bb api /repositories/{workspace}/{repo}/pullrequests \
  --jq '.values[] | select(.state=="OPEN") | {id, title}'
```

| Flag | Description |
|------|-------------|
| `-X, --method` | HTTP method (default: GET, or POST if `-f`/`--input` set) |
| `-f, --field key=value` | Request body field; supports dot notation for nested keys |
| `-H, --header key:value` | Additional HTTP header |
| `--paginate` | Fetch all pages and concatenate into one array |
| `--jq <expr>` | Filter JSON output with a jq expression |
| `--input <file>` | Read request body from file (`-` for stdin) |

### Batch operations

Run any `bb` command across multiple repositories in a workspace:

```bash
# List open PRs in every repo
bb batch -- pr list --state OPEN

# Filter to repos matching a glob
bb batch --repos "backend-*" -- pipeline list --limit 5

# View metadata as JSON across all repos
bb batch --workspace myteam -- repo view --json full_name,language

# Control parallelism
bb batch --concurrency 10 -- issue list
```

Output is printed repo by repo with a separator header. Works best with `bb api` and commands that accept `--workspace`; commands that require a local git checkout will fail gracefully per repo.

### Extensions

Install community-built commands from any git repository:

```bash
bb extension install https://github.com/example/bb-ext-jira
bb ext list
bb ext remove jira
```

Installed extensions appear as top-level `bb` subcommands. An extension is any git repository containing an executable named after the repo (e.g. `bb-ext-jira/bb-ext-jira`).

---

## JSON Output & Scripting

All listing and view commands support `--json` and `--jq` for scripting:

```bash
# List PR IDs and titles as JSON
bb pr list --json id,title,state

# Get the UUID of the latest failed pipeline
bb pipeline list --json uuid,state --jq '.[] | select(.state.result.name=="FAILED") | .uuid' | head -1

# Count open issues
bb issue list --json id --jq 'length'

# My open PRs as JSON (via status)
bb status --json my_prs --jq '.my_prs[] | {id, repo, title}'
```

`--jq` implies `--json` and supports the full [jq language](https://jqlang.github.io/jq/).

---

## Shell Completions

All commands that accept a PR number, branch name, or repo name complete dynamically from the Bitbucket API when you press Tab:

```bash
bb pr checkout <TAB>   # shows open PR numbers with titles
bb branch delete <TAB> # shows branch names
bb repo view <TAB>     # shows repo slugs
```

Enable completions for your shell:

```bash
# Bash (~/.bashrc)
source <(bb completion bash)

# Zsh (~/.zshrc)
bb completion zsh > "${fpath[1]}/_bb"

# Fish
bb completion fish | source

# PowerShell ($PROFILE)
bb completion powershell | Out-String | Invoke-Expression
```

---

## CI/CD Integration

Use `BITBUCKET_USERNAME` + `BITBUCKET_TOKEN` environment variables — no interactive login needed:

```yaml
# GitHub Actions example
- name: Watch pipeline
  env:
    BITBUCKET_USERNAME: ${{ secrets.BB_USERNAME }}
    BITBUCKET_TOKEN: ${{ secrets.BB_TOKEN }}
  run: |
    bb pipeline run --branch ${{ github.ref_name }}
    # Exits 0 on success, 1 on failure
```

---

## Configuration

Config file: `~/.config/bb/config.yaml` (macOS/Linux) or `%APPDATA%\bb\config.yaml` (Windows)

```bash
bb config get default_workspace
bb config set default_workspace myworkspace
```

Per-repo override: create `.bb.yaml` in the repository root.

---

## GitHub Actions Secrets (for releases)

| Secret               | Description                                                           |
| -------------------- | --------------------------------------------------------------------- |
| `BB_OAUTH_CLIENT_ID` | Bitbucket OAuth consumer client ID (embedded in binary at build time) |
| `HOMEBREW_TAP_TOKEN` | GitHub PAT with write access to `chandrasekar-r/homebrew-tap`         |
| `SCOOP_BUCKET_TOKEN` | GitHub PAT with write access to `chandrasekar-r/scoop-bucket`         |

---

## Building from Source

```bash
git clone https://github.com/chandrasekar-r/bitbucket-cli
cd bitbucket-cli
make build        # builds ./bin/bb
make test         # go test -race ./...
make lint         # golangci-lint run ./...
make completions  # generates ./completions/
```

Cross-platform snapshot (requires GoReleaser):

```bash
BB_OAUTH_CLIENT_ID=dev make release-dry-run
```

---

## License

MIT — see [LICENSE](LICENSE)
