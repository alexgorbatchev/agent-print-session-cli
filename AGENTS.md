# agent-print-session-cli

Standalone CLI utility (`print-session`) for inspecting, summarizing, filtering, and printing AI coding agent session logs from disk.

## Shared commands

- Build binary: `just build` (outputs to `bin/print-session`)
- Run CLI: `just run [args...]`
- Run CLI in agent mode: `just run-ai [args...]` (sets `AGENT=1`)
- Run all tests: `just test`
- Enforce 90% statement coverage: `just coverage`
- Lint and check module hygiene: `just lint`
- Run all checks: `just check`
- Format Go source: `just fmt`

## Key gotchas

- **Subject-first CLI hierarchy**: Commands strictly follow subject-first ordering: `print-session [claude|pi] [print|summary|handoff] <session-id>`. Single-argument shorthand `print-session [claude|pi] <session-id>` invokes `print`.
- **Dual-mode output (`AGENT=1`)**: When `AGENT=1`, output must omit decorative dividers and tables in favor of compact, token-conservative formatting.
- **Read-only execution**: Never mutate or create files inside user transcript or session storage directories.

## Boundaries

- **Always**: automatically record all new user instructions in `AGENTS.md` immediately upon receipt (check with user if existing instructions conflict).
- **Always**: any time code is changed such that results from running that code are changed, a test file must be changed as well; 90% code coverage is required.
- **Always**: run `just check` before committing changes.
- **Never**: publish releases, tags, packages, or production deployments automatically without explicit user authorization.
- **Never**: modify files in transcript search directories.

## References

- CLI entrypoint: `cmd/print-session/main.go`
- Session printing: `internal/session/print.go`
- Session resolution: `internal/session/resolve.go`
- Session summaries: `internal/session/summary.go`
- Work handoff: `internal/session/handoff.go`
- Upstream parser library: [`alexgorbatchev/agent-parser`](https://github.com/alexgorbatchev/agent-parser)
