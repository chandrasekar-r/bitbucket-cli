---
date: 2026-03-30
topic: bb-cli
---

# bb — Bitbucket Cloud CLI

## Problem Frame

Bitbucket Cloud has no official CLI tool. Teams are forced to use the web UI or raw API calls for everyday dev workflows — creating PRs, watching pipelines, managing repos. GitHub's `gh` CLI has set the bar. `bb` fills that gap for Bitbucket Cloud users with a `gh`-compatible command idiom, fast Go binary, and full cross-platform support.

## Requirements

### Authentication
- R1. `bb auth login` opens a browser-based OAuth 2.0 flow to authenticate against Bitbucket Cloud.
- R2. `bb auth login --with-token` accepts an App Password via stdin as an alternative auth method.
- R3. `bb auth logout` revokes and removes stored credentials.
- R4. `bb auth status` shows the currently authenticated accounts and active workspace.
- R5. `bb auth token` prints the active access token for scripting use.
- R6. Multiple Bitbucket accounts can be added and switched between. The active account is tracked per-workspace config.

### Repositories
- R7. `bb repo list` lists repos in the active workspace with filtering (language, access level).
- R8. `bb repo view [repo]` shows repo metadata, description, clone URLs, and default branch.
- R9. `bb repo create` creates a new repo, with interactive prompts for name, visibility, and description when flags are omitted.
- R10. `bb repo clone [repo]` clones a repo by slug; defaults to the current workspace.
- R11. `bb repo fork [repo]` forks a repo into the authenticated user's workspace.
- R12. `bb repo delete [repo]` deletes a repo (requires explicit confirmation prompt or `--confirm` flag).
- R13. `bb repo rename [repo] [new-name]` renames a repo.
- R14. `bb repo browse [repo]` opens the repo in the system browser.

### Pull Requests
- R15. `bb pr list` lists open PRs for the current repo with status, author, and branch info.
- R16. `bb pr view [number]` shows PR details, description, reviewers, approvals, and diff summary.
- R17. `bb pr create` opens an interactive prompt to set title, description, reviewers, source/target branch, and labels. Can be fully specified via flags.
- R18. `bb pr merge [number]` merges a PR with strategy selection (merge commit, squash, fast-forward). Interactive if strategy not specified.
- R19. `bb pr approve [number]` approves a PR.
- R20. `bb pr request-changes [number]` marks a PR as needing changes.
- R21. `bb pr decline [number]` declines and closes a PR.
- R22. `bb pr checkout [number]` checks out the PR's source branch locally.
- R23. `bb pr comment [number]` adds a general comment to a PR; supports inline file comments via flags.
- R24. `bb pr diff [number]` displays the PR diff in the terminal.
- R25. `bb pr browse [number]` opens the PR in the system browser.

### Branches
- R26. `bb branch list` lists branches with last commit and author info.
- R27. `bb branch create [name]` creates a branch from a specified source (defaults to default branch).
- R28. `bb branch delete [name]` deletes a local and/or remote branch.
- R29. `bb branch rename [old] [new]` renames a branch.
- R30. `bb branch protect [name]` enables branch restrictions (no direct pushes, required approvals) on the target branch.

### Pipelines (CI/CD)
- R31. `bb pipeline list` lists recent pipeline runs for the current repo with status and duration.
- R32. `bb pipeline view [run-id]` shows details for a specific pipeline run including step logs.
- R33. `bb pipeline run [branch]` triggers a new pipeline run on the specified branch.
- R34. `bb pipeline cancel [run-id]` cancels a running pipeline.
- R35. `bb pipeline watch [run-id]` streams live pipeline log output until completion (exits with the pipeline's exit code).

### Issues
- R36. `bb issue list` lists open issues with filtering by assignee, label, and state.
- R37. `bb issue view [number]` shows issue details and comments.
- R38. `bb issue create` creates an issue with interactive prompts for title, description, assignee, and priority.
- R39. `bb issue close [number]` closes an issue.
- R40. `bb issue reopen [number]` reopens a closed issue.
- R41. `bb issue comment [number]` adds a comment to an issue.

### Snippets
- R42. `bb snippet list` lists the authenticated user's snippets.
- R43. `bb snippet view [id]` shows snippet content and metadata.
- R44. `bb snippet create` creates a snippet from stdin or a file path.
- R45. `bb snippet edit [id]` opens the snippet in `$EDITOR`.
- R46. `bb snippet delete [id]` deletes a snippet (requires confirmation).
- R47. `bb snippet clone [id]` clones a snippet repository locally.

### Workspaces
- R48. `bb workspace list` lists all workspaces accessible to the authenticated user.
- R49. `bb workspace switch [slug]` sets the active workspace for subsequent commands.
- R50. `bb workspace view [slug]` shows workspace metadata and members.

### Output & Scripting
- R51. All listing and view commands support a `--json [fields]` flag that outputs structured JSON for the specified fields.
- R52. All commands with `--json` also support `--jq [expression]` to apply a jq filter expression to the JSON output.
- R53. `--limit` / `-L` flag on all list commands controls the maximum number of results returned.
- R54. `--no-tty` flag disables all interactive prompts; commands fail with a clear error if required flags are missing.

### Shell Completions & Help
- R55. `bb completion [bash|zsh|fish|powershell]` generates shell completion scripts for all supported shells.
- R56. `--help` on all commands and subcommands produces consistent, concise usage documentation including flag descriptions and examples.
- R57. `bb help [command]` is equivalent to `[command] --help`.

### Configuration
- R58. Config is stored in the OS-appropriate config directory (`~/.config/bb/` on Unix, `%APPDATA%\bb\` on Windows).
- R59. Per-repo config overrides can be stored in `.bb` at the repository root (e.g., default reviewer list, default base branch).
- R60. `bb config set [key] [value]` and `bb config get [key]` manage CLI settings.

### Distribution & Platform
- R61. Pre-built binaries are released on GitHub Releases for macOS (Intel + Apple Silicon), Linux (amd64 + arm64), and Windows (amd64).
- R62. Homebrew tap for macOS/Linux installation.
- R63. Scoop bucket and Winget package for Windows installation.
- R64. `.deb` and `.rpm` packages for Linux package manager installation.
- R65. A shell install script (`curl | sh`) is provided as a universal fallback.

## Success Criteria

- A developer can go from zero to an authenticated session and merged PR using only `bb` commands, without touching the Bitbucket web UI.
- CI/CD scripts can use `bb pipeline watch` to gate on pipeline results and parse JSON output without intermediate tools beyond `jq`.
- All core commands run in under 500ms on a standard machine (excluding network latency).
- The binary installs cleanly on all three platforms via their respective package managers.
- `bb auth login` completes the OAuth flow and stores credentials without requiring manual token creation.

## Scope Boundaries

- Bitbucket Server / Data Center is out of scope for v1. Cloud-only.
- Bitbucket Jira integration (syncing issues between Jira and Bitbucket) is out of scope.
- A TUI (terminal UI dashboard) is out of scope — this is a command-oriented tool, not an interactive dashboard.
- Repository-level access control management (IP allowlists, 2FA enforcement) is out of scope for v1.
- Custom Bitbucket Connect apps or webhook management is out of scope for v1.

## Key Decisions

- **Go**: Industry standard for modern CLI tools (gh, kubectl, terraform). Single static binary, fast compile, excellent cross-platform support.
- **`bb` binary name**: Mirrors `gh`. Short, memorable, matches Bitbucket abbreviation. Users type `bb pr create`, not `bitbucket pr create`.
- **OAuth 2.0 + App Password**: Browser OAuth is the primary path; app password accommodates CI/CD environments and headless setups.
- **Interactive prompts by default**: When required flags are omitted, commands guide the user interactively. `--no-tty` disables this entirely for scripting.
- **`--json` + `--jq`**: First-class scripting support modeled on `gh`. Essential for CI/CD pipeline automation use cases.
- **Bundled OAuth consumer**: Client ID/secret for a registered Bitbucket OAuth consumer are embedded in the binary at build time (injected via `ldflags`). Users run `bb auth login` with zero setup — identical to how `gh` works.

## Dependencies / Assumptions

- Bitbucket Cloud REST API v2.0 (`https://api.bitbucket.org/2.0`) is the sole API target.
- OAuth 2.0 requires registering a Bitbucket OAuth consumer (client ID/secret) that ships with the binary or is user-provided.
- The user's machine has a browser available for the OAuth flow; headless environments fall back to app password auth.
- Bitbucket's Issues feature must be enabled per-repo (not all repos have it).

## Outstanding Questions

### Resolve Before Planning

- [Affects R1, R6][User decision] Should the OAuth consumer credentials (client ID/secret) be bundled with the binary (like `gh` ships with GitHub OAuth app credentials), or must users register their own Bitbucket OAuth consumer? Bundling is the better UX but requires managing a shared OAuth app registration.

### Deferred to Planning

- [Affects R35][Technical] How should `bb pipeline watch` stream logs — polling the Bitbucket Pipelines API at intervals vs. server-sent events (if supported by the API)?
- [Affects R1][Needs research] Does Bitbucket Cloud's OAuth 2.0 support PKCE for native app flows, or does it require a client secret? This affects how credentials are stored and whether the flow is truly secret-free.
- [Affects R59][Technical] Should `.bb` per-repo config be a separate format or merge with the global config schema?
- [Affects R61–R65][Technical] GoReleaser vs. manual Makefile targets for the cross-platform build and packaging pipeline.

## Next Steps

→ Resolve the OAuth consumer bundling question (under `Resolve Before Planning`), then proceed with `/ce:plan` for structured implementation planning.
