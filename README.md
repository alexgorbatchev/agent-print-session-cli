`agent-print-session-cli` provides the `print-session` command-line utility for locating, inspecting, filtering, and printing AI coding agent session transcripts from disk. It translates raw JSONL streams into human-readable terminal output, structured JSON, or concise agent-facing context blocks for resuming past work.

# What It Does

- **Automatic transcript discovery**: Locates session files by exact path, full session ID, or prefix across standard storage locations (`~/.claude/projects`, `~/.claude/sessions`, `~/.pi/agent/sessions`, and `$XDG_DATA_HOME`).
- **Subject-first CLI commands**: Provides consistent subcommands across supported harnesses (`print-session claude`, `print-session pi`).
- **Granular event filtering**: Focuses output on errors (`--errors`), file edits (`--files`), tool executions (`--tools`), prompts (`--prompts`), or the most recent turn (`--last-turn`).
- **Work resumption & handoff**: Extracts initial goals, active branches, modified files, recent errors, and next steps via `handoff` (or aliases `continue`, `resume`).
- **High-level summaries**: Aggregates token usage, session duration, and tool execution tallies via `summary`.
- **Dual-mode output**: Automatically switches to dense, token-conservative formatting when invoked with `AGENT=1`.

# How It Works

- The user invokes `print-session <harness> <action> <session-id>` (or uses shorthand `print-session <harness> <session-id>`).
- `print-session` resolves the session target: if a direct file path is provided, it reads that path; otherwise, it scans standard harness directories for matching `.jsonl` files.
- The utility parses each line using `github.com/alexgorbatchev/agent-parser`, constructing a structured event stream.
- The stream passes through requested filters (`--files`, `--errors`, `--tail`, `--last-turn`) and pagination.
- Filtered events are rendered to stdout in styled human mode, machine-readable JSON (`--json`), or token-conservative agent mode (`AGENT=1`).

# How it Really Works

- **Discovery traversal**: Search routines scan projects and session directories non-recursively for exact UUID matches first, falling back to prefix matching across candidate `.jsonl` filenames without invoking shell utilities.
- **Zero file mutation**: `print-session` strictly opens transcript files in read-only mode, never writing or creating temporary files.
- **Agent mode token conservation**: When `AGENT=1`, horizontal divider lines, decorative boxes, and padding are omitted in favor of compact text tags (`OK:`, `ERR:`), and output defaults to the most recent 100 events unless `--all` is specified.
- **Diff formatting**: Tool calls containing file modifications (`Edit`, `Write`, `apply_patch`) are formatted into colored terminal diffs with line numbers and added/deleted indicators.
- **Exit codes**: Returns `0` on successful parsing, `1` on missing sessions, invalid flags, or unreadable files.

# Prerequisites

- macOS or Linux
- No runtime dependencies (pure Go binary)

# Installation

Download the latest prebuilt binary from the [GitHub Releases](https://github.com/alexgorbatchev/agent-print-session-cli/releases) page.

```bash
curl -sSL https://github.com/alexgorbatchev/agent-print-session-cli/releases/download/v1.0.0/agent-print-session-cli_1.0.0_darwin_arm64.tar.gz | tar -xz
sudo mv print-session /usr/local/bin/
```

# Quick Start

### Inspect a session transcript

```bash
print-session claude c0ffee-1234
```

Sample Output:
```text
Session: Refactor Authentication Flow
Branch:  feat/auth-jwt
CWD:     /workspace/repo
Events:  14

[09:15:02] USER PROMPT
Please migrate authentication from sessions to JWT.

[09:15:04] TOOL CALL
Tool: Edit
File: src/auth.go (EDIT)
@@ -10,3 +10,4 @@
-func ValidateSession(s string) bool
+func ValidateToken(t string) (*Claims, error)

[09:15:06] TURN END
Model: claude-3-7-sonnet | Tokens: 4,120 in, 580 out
```

### Resume past work with a handoff summary

```bash
print-session claude handoff c0ffee-1234
```

### Output JSON for automated processing

```bash
print-session pi print --json --last-turn 4d3f21
```

# Options & Flags

### Global Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--help` | `-h` | `false` | Display help screen and command hierarchy |
| `--version` | `-v` | `false` | Display version information |

### `print-session [claude|pi] print` Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--last-turn` | | `false` | Output only the final conversational turn |
| `--errors` | | `false` | Filter to errors and failed tool executions only |
| `--files` | | `false` | Filter to file modifications and diffs only |
| `--tools` | | `false` | Filter to tool invocations and results only |
| `--prompts` | | `false` | Filter to user prompts and turn boundaries only |
| `--summary` | | `false` | Output high-level session summary only |
| `--handoff` | | `false` | Output continuation context for resuming work |
| `--json` | | `false` | Output events as formatted JSON |
| `--tail <N>` | | `0` | Show only the last N events (0 means all) |
| `--limit <N>`| | `0` | Show only the first N events (0 means all) |
| `--all` | | `false` | Output full transcript without default agent-mode limit |
| `--path <path>`| | `""` | Direct path to session file |

# Supported Harnesses

| Harness | Subcommand | Default Search Paths |
| :--- | :--- | :--- |
| **Claude Code** | `print-session claude` | `~/.claude/projects/`, `~/.claude/sessions/` |
| **Pi Coding Agent** | `print-session pi` | `~/.pi/agent/sessions/` |

# License

[MIT](LICENSE)
