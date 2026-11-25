default:
    @just --list

build:
    cargo build --locked --bin agent-print-session
    mkdir -p bin
    cp target/debug/agent-print-session bin/agent-print-session

run *args:
    cargo run --locked --bin agent-print-session -- {{args}}

run-ai *args:
    AGENT=1 cargo run --locked --bin agent-print-session -- {{args}}

test:
    cargo test --locked --all-targets

# The workspace contains one CLI package. Tests load the agent skill from source.
test-affected base="origin/main":
    #!/usr/bin/env bash
    set -euo pipefail
    base=$(git merge-base {{ quote(base) }} HEAD)
    files=$(git diff --name-only "$base" --; git ls-files --others --exclude-standard)
    while IFS= read -r file; do
        case "$file" in
            ""|AGENTS.md|README.md|docs/*|LICENSE) ;;
            *)
                echo 'Selected agent-print-session-cli: package or test inputs changed.'
                just test
                exit
                ;;
        esac
    done <<< "$files"
    echo 'No agent-print-session test inputs changed.'

fmt:
    cargo fmt --package agent-print-session-cli

fmt-check:
    cargo fmt --package agent-print-session-cli -- --check

lint:
    cargo clippy --locked --all-targets -- -D warnings

coverage:
    cargo llvm-cov --locked --all-targets --ignore-filename-regex '(^|/)(tests|scripts)/' --fail-under-lines 90 --fail-under-functions 90

check: fmt-check lint test coverage
