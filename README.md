# bb — Bitbucket Cloud CLI

`bb` is a fast, cross-platform command-line tool for Bitbucket Cloud, modeled on GitHub's [`gh`](https://cli.github.com). Manage pull requests, repositories, pipelines, branches, issues, and more — without leaving your terminal.

```
bb pr create --title "Fix login bug" --base main
bb pipeline watch
bb repo clone myworkspace/my-service
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

# Create a PR from the current branch
cd my-service
git checkout -b feat/my-feature
git push -u origin feat/my-feature
bb pr create --title "Add my feature"

# Watch a pipeline
bb pipeline watch

# List open PRs
bb pr list

# Merge PR #42
bb pr merge 42
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

| Command                        | Description                                  |
| ------------------------------ | -------------------------------------------- |
| `bb branch list`               | List branches                                |
| `bb branch create <name>`      | Create a branch (`--from` to specify source) |
| `bb branch delete <name>`      | Delete a branch                              |
| `bb branch rename <old> <new>` | Rename a branch                              |
| `bb branch protect <name>`     | Enable branch restrictions                   |

### Pull Requests

| Command                   | Description                                                |
| ------------------------- | ---------------------------------------------------------- |
| `bb pr list`              | List open PRs (`--state` for MERGED/DECLINED)              |
| `bb pr view <number>`     | Show PR details                                            |
| `bb pr create`            | Create a PR (interactive or `--title --base`)              |
| `bb pr merge <number>`    | Merge a PR (`--strategy` merge_commit/squash/fast_forward) |
| `bb pr approve <number>`  | Approve a PR                                               |
| `bb pr decline <number>`  | Decline a PR                                               |
| `bb pr checkout <number>` | Check out PR source branch locally                         |
| `bb pr comment <number>`  | Add a comment (`--body` or opens `$EDITOR`)                |
| `bb pr diff <number>`     | Show unified diff                                          |
| `bb pr browse <number>`   | Open in browser                                            |

### Pipelines

| Command                     | Description                                   |
| --------------------------- | --------------------------------------------- |
| `bb pipeline list`          | List recent pipeline runs (`--branch` filter) |
| `bb pipeline view <uuid>`   | Show pipeline details and step status         |
| `bb pipeline run`           | Trigger a pipeline (default: current branch)  |
| `bb pipeline cancel <uuid>` | Cancel a running pipeline                     |
| `bb pipeline watch [uuid]`  | Follow logs in real time (polling)            |

`bb pipeline watch` exits with the pipeline's result code:

- `0` — SUCCESSFUL
- `1` — FAILED / ERROR / STOPPED
- `2` — watch error (network, timeout)

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
```

`--jq` implies `--json` and supports the full [jq language](https://jqlang.github.io/jq/).

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

## Shell Completions

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
