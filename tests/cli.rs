use agent_print_session_cli::cli;
use cli_command_kit::Mode;
mod support;
use std::{
    fs,
    io::{self, Write},
    process::Command,
};

const CLAUDE: &str = concat!(
    "{\"type\":\"ai-title\",\"title\":\"Test session\",\"sessionId\":\"cli-test\"}\n",
    "{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":\"Hello world\"},\"uuid\":\"u-1\",\"sessionId\":\"cli-test\",\"timestamp\":\"2026-09-20T12:00:00Z\"}\n",
    "{\"type\":\"assistant\",\"message\":{\"id\":\"m1\",\"role\":\"assistant\",\"model\":\"claude\",\"content\":[{\"type\":\"text\",\"text\":\"Done\"}],\"usage\":{\"input_tokens\":10,\"output_tokens\":12}},\"uuid\":\"a-1\",\"sessionId\":\"cli-test\",\"timestamp\":\"2026-09-20T12:00:01Z\"}\n"
);
const AGY: &str = r#"{"step_index":0,"type":"USER_INPUT","created_at":"2026-11-22T14:38:26Z","content":"Read the release notes."}
{"step_index":1,"type":"PLANNER_RESPONSE","created_at":"2026-11-22T14:38:27Z","model":"gemini-2.5-pro","tool_calls":[{"id":"read-1","name":"view_file","args":{"AbsolutePath":"/repo/RELEASE.md"}},{"id":"read-2","name":"view_file","args":{"TargetFile":"/repo/MISSING.md"}}]}
{"step_index":2,"type":"GENERIC","created_at":"2026-11-22T14:38:28Z","status":"DONE","content":"Release 1.0 is ready."}
{"step_index":3,"type":"GENERIC","created_at":"2026-11-22T14:38:29Z","status":"ERROR","error":"File missing."}
{"step_index":4,"type":"PLANNER_RESPONSE","created_at":"2026-11-22T14:38:30Z","model":"gemini-2.5-pro","content":"The notes are ready for release.","status":"DONE","input_tokens":100,"output_tokens":20}
"#;

fn run(args: &[&str], mode: Mode) -> Result<String, agent_print_session_cli::Error> {
    let mut output = Vec::new();
    cli::run(
        std::iter::once("agent-print-session").chain(args.iter().copied()),
        mode,
        &mut output,
    )?;
    Ok(String::from_utf8(output).unwrap())
}
#[test]
fn command_tree_native_help_version_completion_and_skill_contracts() {
    cli::command(Mode::Human).debug_assert();
    for mode in [Mode::Human, Mode::Agent] {
        for harness in ["agy", "claude", "pi", "codex"] {
            assert!(run(&[harness], mode).unwrap().contains("handoff"));
            for verb in ["print", "summary", "handoff", "continue", "resume"] {
                let text = run(&[harness, verb, "--help"], mode).unwrap();
                assert!(text.contains("--path"));
                if mode == Mode::Agent {
                    assert!(text.starts_with("ALERT: Agents must read"));
                }
                assert!(run(&[harness, verb], mode).is_err());
            }
            assert!(
                run(&["help", harness, "print"], mode)
                    .unwrap()
                    .contains("--json")
            );
        }
        for args in [
            vec![],
            vec!["--help"],
            vec!["help"],
            vec!["help", "completion"],
            vec!["completion", "--help"],
            vec!["completion"],
            vec!["skill", "--help"],
        ] {
            let text = run(&args, mode).unwrap();
            if mode == Mode::Agent {
                assert!(text.starts_with("ALERT:"), "{args:?}: {text}");
            }
        }
        for shell in ["bash", "zsh", "fish", "powershell"] {
            assert!(
                run(&["completion", shell], mode)
                    .unwrap()
                    .contains("agent-print-session")
            );
            assert!(
                run(&["completion", shell, "--help"], mode)
                    .unwrap()
                    .contains(shell)
            );
        }
        assert_eq!(run(&["skill"], mode).unwrap(), INSTALL_TEXT);
        assert_eq!(
            run(&["--version"], mode).unwrap(),
            concat!(env!("CARGO_PKG_VERSION"), "\n")
        );
        assert_eq!(
            run(&["-v"], mode).unwrap(),
            concat!(env!("CARGO_PKG_VERSION"), "\n")
        );
        for args in [
            vec!["skill", "extra"],
            vec!["skill", "--json"],
            vec!["bogus"],
            vec!["claude", "print", "one", "two"],
            vec!["claude", "print", "--path", ""],
            vec!["pi", "print", "--limit", "bad", "id"],
        ] {
            assert!(run(&args, mode).is_err(), "{args:?}");
        }
    }
}
#[test]
fn native_cli_all_harnesses_filters_formats_aliases_and_path_shorthand() {
    let dir = support::temporary_directory();
    let path = dir.path().join("session.jsonl");
    fs::write(&path, CLAUDE).unwrap();
    let path = path.to_str().unwrap();
    for harness in ["agy", "claude", "codex", "pi"] {
        for mode in [Mode::Human, Mode::Agent] {
            for verb in ["print", "summary", "handoff", "continue", "resume"] {
                assert!(!run(&[harness, verb, path], mode).unwrap().is_empty());
                let json = run(&[harness, verb, "--json", "--path", path], mode).unwrap();
                serde_json::from_str::<serde_json::Value>(&json).unwrap();
            }
            assert!(!run(&[harness, "--path", path], mode).unwrap().is_empty());
            assert!(!run(&[harness, path], mode).unwrap().is_empty());
        }
    }
    for flag in [
        "--summary",
        "--handoff",
        "--last-turn",
        "--errors",
        "--files",
        "--tools",
        "--prompts",
        "--all",
    ] {
        run(&["claude", "print", flag, path], Mode::Agent).unwrap();
    }
    let text = run(
        &["claude", "print", "--handoff", "--summary", path],
        Mode::Agent,
    )
    .unwrap();
    assert!(text.starts_with("CONTINUATION_CONTEXT"));
    for (flag, value) in [("--tail", "1"), ("--limit", "1"), ("--limit", "-1")] {
        run(&["claude", "print", flag, value, path], Mode::Agent).unwrap();
    }
    assert!(
        run(&["claude", "print", "--json=false", path], Mode::Human)
            .unwrap()
            .contains("[SESSION LOG]")
    );
    assert!(
        run(&["claude", "print", "--path", "missing"], Mode::Agent)
            .unwrap_err()
            .to_string()
            .contains("opening transcript")
    );
    assert!(
        run(
            &["claude", "print", "definitely-missing-session-zzzz"],
            Mode::Agent
        )
        .is_err()
    );
}

#[test]
fn antigravity_reads_native_transcripts_and_resolves_conversations() {
    let dir = support::temporary_directory();
    let root = dir.path();
    let custom = root.join("custom");
    let gemini = root.join("gemini");
    let home = root.join("home");
    let gemini_brain = gemini.join("antigravity-cli/brain");
    let home_brain = home.join(".gemini/antigravity-cli/brain");
    let create_transcript = |brain: &std::path::Path, id: &str, name: &str| {
        let logs = brain.join(id).join(".system_generated/logs");
        fs::create_dir_all(&logs).unwrap();
        let file = logs.join(name);
        fs::write(&file, AGY).unwrap();
        file
    };
    let bad_match = create_transcript(&custom, "a-conversation", "transcript.jsonl");
    let prefix = create_transcript(&custom, "conversation-prefix", "transcript.jsonl");
    let exact = create_transcript(&custom, "conversation", "transcript_full.jsonl");
    let fallback = create_transcript(&custom, "conversation", "transcript.jsonl");
    fs::write(
        &fallback,
        AGY.replace("Read the release notes.", "Read the partial log."),
    )
    .unwrap();
    fs::create_dir_all(custom.join("a-conversation-prefix")).unwrap();
    let gemini_log = create_transcript(&gemini_brain, "conversation", "transcript_full.jsonl");
    let home_log = create_transcript(&home_brain, "home-id", "transcript.jsonl");
    let direct_dir = root.join("direct");
    fs::create_dir(&direct_dir).unwrap();
    let direct = direct_dir.join("transcript.jsonl");
    fs::write(&direct, AGY).unwrap();
    let invoke = |args: &[&str], custom_dir: &std::path::Path| {
        Command::new(env!("CARGO_BIN_EXE_agent-print-session"))
            .env("HOME", &home)
            .env("USERPROFILE", &home)
            .env("AGENT", "1")
            .env("AGY_SESSIONS_DIR", custom_dir)
            .env("GEMINI_HOME", &gemini)
            .args(args)
            .output()
            .unwrap()
    };
    for (target, directory, expected) in [
        ("conversation", custom.as_path(), &exact),
        ("conversation-pre", custom.as_path(), &prefix),
        ("a-conver", custom.as_path(), &bad_match),
        ("versation.jsonl", custom.as_path(), &bad_match),
        ("conversation", direct_dir.as_path(), &gemini_log),
        ("home-id", custom.as_path(), &home_log),
    ] {
        let output = invoke(&["agy", "summary", "--json", target], directory);
        assert!(
            output.status.success(),
            "{target}: {}",
            String::from_utf8_lossy(&output.stderr)
        );
        let summary: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
        assert_eq!(summary["filePath"], expected.to_str().unwrap(), "{target}");
        assert_eq!(summary["tokensTotal"], 120);
        assert_eq!(summary["tools"][0]["calls"], 2);
        assert_eq!(summary["tools"][0]["errorCalls"], 1);
    }
    let output = invoke(&["agy", "missing-session-unique"], &custom);
    assert_eq!(output.status.code(), Some(1));
    assert!(String::from_utf8_lossy(&output.stderr).contains("missing-session-unique"));
    for mode in [Mode::Human, Mode::Agent] {
        let text = run(
            &["antigravity", "--path", direct_dir.to_str().unwrap()],
            mode,
        )
        .unwrap();
        assert!(text.contains("Read the release notes."));
        assert!(text.contains("Path: /repo/RELEASE.md"));
        assert!(text.contains("Path: /repo/MISSING.md"));
        assert!(text.contains("The notes are ready for release."));
        let events = run(&["agy", "print", "--json", direct.to_str().unwrap()], mode).unwrap();
        let events: serde_json::Value = serde_json::from_str(&events).unwrap();
        assert_eq!(events[0]["eventType"], "turn_start");
        let handoff = run(&["agy", "resume", exact.to_str().unwrap()], mode).unwrap();
        assert!(handoff.contains("File missing."));
    }
}
struct Broken;
impl Write for Broken {
    fn write(&mut self, _: &[u8]) -> io::Result<usize> {
        Err(io::Error::new(io::ErrorKind::BrokenPipe, "closed"))
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}
#[test]
fn metadata_and_error_output_propagate_failures() {
    for args in [
        vec!["agent-print-session"],
        vec!["agent-print-session", "--help"],
        vec!["agent-print-session", "--version"],
        vec!["agent-print-session", "skill"],
        vec!["agent-print-session", "completion", "bash"],
        vec!["agent-print-session", "claude"],
    ] {
        assert!(cli::run(args, Mode::Agent, &mut Broken).is_err());
    }
    for mode in [Mode::Agent, Mode::Human] {
        let error = run(&["wrong"], mode).unwrap_err();
        let mut out = Vec::new();
        cli::report(&error, mode, &mut out).unwrap();
        assert!(String::from_utf8(out).unwrap().contains("wrong"));
        assert!(cli::report(&error, mode, &mut Broken).is_err());
    }
}
#[test]
fn real_binary_discovers_env_paths_folds_home_and_reports_runtime_error_once() {
    let dir = support::temporary_directory();
    let home = dir.path();
    let envdir = home.join("data");
    fs::create_dir(&envdir).unwrap();
    fs::write(envdir.join("session-unique.jsonl"), CLAUDE).unwrap();
    for harness in ["claude", "pi", "codex"] {
        let mut command = Command::new(env!("CARGO_BIN_EXE_agent-print-session"));
        command
            .env("HOME", home)
            .env("XDG_DATA_HOME", home.join("xdg"))
            .env("AGENT", "1")
            .env("CLAUDE_PROJECTS_DIR", &envdir)
            .env("CLAUDE_SESSIONS_DIR", "")
            .env("PI_SESSIONS_DIR", &envdir)
            .env("CODEX_SESSIONS_DIR", &envdir)
            .env("CODEX_HOME", home.join("codex"));
        let output = command.args([harness, "unique"]).output().unwrap();
        assert!(
            output.status.success(),
            "{}",
            String::from_utf8_lossy(&output.stderr)
        );
        assert!(String::from_utf8_lossy(&output.stdout).starts_with("SESSION: ~/data/"));
    }
    let output = Command::new(env!("CARGO_BIN_EXE_agent-print-session"))
        .env("AGENT", "1")
        .args([
            "claude",
            "print",
            "--path",
            home.join("missing").to_str().unwrap(),
        ])
        .output()
        .unwrap();
    assert_eq!(output.status.code(), Some(1));
    let stderr = String::from_utf8_lossy(&output.stderr);
    assert_eq!(stderr.matches("ERR:").count(), 1);
    assert!(!stderr.contains("Usage:"));
    assert!(output.stdout.is_empty());
    let output = Command::new(env!("CARGO_BIN_EXE_agent-print-session"))
        .arg("--version")
        .output()
        .unwrap();
    assert!(output.status.success());
}

const INSTALL_TEXT: &str = "Install the agent-print-session skill using the Skills CLI:\n\n  npx skills add alexgorbatchev/agent-print-session-cli --skill agent-print-session\n\nSkill:\n\n  https://github.com/alexgorbatchev/agent-print-session-cli/tree/main/skills/agent-print-session\n";
