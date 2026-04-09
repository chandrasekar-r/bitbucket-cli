# Changelog

All notable changes to `bb` are documented here. Dates are UTC. Versions follow [Semantic Versioning](https://semver.org).

---

## [0.3.0] — 2026-04-09

### New commands

- **`bb pr edit`** — edit a PR's title, description, base branch, or reviewers after it has been created. Accepts `--title`, `--body`, `--base`, and `--add-reviewer` flags; runs an interactive form when no flags are given.
- **`bb branch tidy`** — cross-references local branches against Bitbucket PRs and deletes the ones whose PRs are merged or declined. Supports `--dry-run` and `--force`.
- **`bb status`** — personal cross-repo dashboard. Shows your open PRs (with approval counts) and PRs awaiting your review, across all repos in the workspace.
- **`bb api`** — raw authenticated API proxy. Sends any request to the Bitbucket REST API with auth, placeholder substitution (`{workspace}`, `{repo}`), `--paginate`, `--jq` filtering, and nested `-f key=value` body fields.
- **`bb batch`** — runs any `bb` command across all (or glob-filtered) repos in a workspace, with configurable concurrency.
- **`bb extension`** (`bb ext`) — install, list, and remove CLI plugins from git repositories. Installed extensions become top-level `bb` subcommands.

### Enhancements

- **Smart PR descriptions** — `bb pr create` now auto-generates the PR body from `git log`, a `.bitbucket/PULL_REQUEST_TEMPLATE.md` template if present, and appends a `Related:` link when the branch name contains a Jira key (e.g. `feat/PROJ-123-my-feature`).
- **Interactive picker** — PR and branch commands (`view`, `merge`, `approve`, `decline`, `checkout`, `comment`, `diff`, `browse`, `edit`, `delete`, `rename`) now show a filterable interactive list when the number/name argument is omitted. Type to filter, Enter to select.
- **Dynamic shell completions** — Tab-completing PR numbers, branch names, and repo slugs now queries the Bitbucket API in real time. Completions include a description column (PR title, repo description).
- **Pipeline desktop notifications** — `bb pipeline watch --notify` fires a native OS notification (macOS `osascript`, Linux `notify-send`, Windows PowerShell) when the pipeline completes, so you can switch to other work while it runs.

### Bug fixes

- `bb branch tidy`: fixed a critical bug where `ListPRsForBranch` only returned OPEN PRs regardless of the requested state. The Bitbucket API requires state as a dedicated query parameter; passing it via the `q` filter has no effect. The command now correctly fetches MERGED and DECLINED PRs.
- `bb api`: HTTP method is now uppercased before the request is sent. Lowercase methods (e.g. `-X get`) were silently rejected by some servers.
- `bb api`: `--input -` now reads from stdin. Previously it tried to open a literal file named `-`.
- `bb api`: 4xx responses no longer print the error twice (once from the command, once from the Execute handler).
- `bb extension remove`: added path-traversal guard on the extension name argument; now returns a clear error for non-installed extensions instead of silently succeeding.
- `bb extension install`: fixed binary discovery fallback — the secondary path was computing the same filename as the primary path, making it a no-op.
- `bb status`: added a goroutine semaphore (concurrency 5) to avoid firing one API request per repo simultaneously and hitting the Bitbucket rate limit.
- `bb status`: summary stats now go to stdout, not stderr.
- `bb pr create`: Jira `Related:` link no longer starts the PR description with `\n\nRelated:` when the body is empty.
- `pkg/notify`: Windows PowerShell notification now escapes single quotes in the title and message to prevent script injection.

---

## [0.2.0] — 2026-03-14

### Bug fixes

- `bb repo clone`: OAuth token is now injected into the HTTPS clone URL to avoid interactive password prompts.
- Security audit fixes: command injection via `$EDITOR`, git branch flag injection, dangerous clone URL schemes, and temp file path traversal.
- `bb pipeline watch`, `bb pipeline list`: `--limit 0` now fetches all results (documented as 0 = all).
- Install script: auto-elevates with `sudo` when the target directory requires it; extracts all tar contents correctly.

---

## [0.1.0] — 2026-02-15

Initial release.

### Commands

- `bb auth login / logout / status / token` — OAuth 2.0 browser flow and API token authentication; transparent token refresh with file locking.
- `bb workspace list / use / view` — workspace management.
- `bb repo list / view / create / clone / fork / delete / rename / browse` — repository operations.
- `bb branch list / create / delete / rename / protect` — branch management.
- `bb pr list / view / create / merge / approve / decline / checkout / comment / diff / browse` — full pull request lifecycle.
- `bb pipeline list / view / run / cancel / watch` — CI/CD pipeline management with live log polling and exit codes.
- `bb issue list / view / create / close / reopen / comment` — issue tracking.
- `bb snippet list / view / create / edit / delete / clone` — snippet management.
- `bb completion bash / zsh / fish / powershell` — shell completions.
- `bb version` — version info with `--json` output.
- Global `--json` / `--jq` flags on all list/view commands.
- Global `--workspace` / `--no-tty` persistent flags.
- Distributed via Homebrew, Scoop, apt/rpm, and curl installer.

[0.3.0]: https://github.com/chandrasekar-r/bitbucket-cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/chandrasekar-r/bitbucket-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/chandrasekar-r/bitbucket-cli/releases/tag/v0.1.0
