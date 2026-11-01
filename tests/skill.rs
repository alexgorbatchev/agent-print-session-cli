use agent_print_session_cli::cli;
use cli_command_kit::Mode;
#[test]
fn canonical_usage_skill_covers_the_live_native_command_tree() {
    let skill = include_str!("../skills/agent-print-session/SKILL.md");
    let missing =
        cli_command_kit::skill_omissions(&cli::command(Mode::Agent), &cli::help_options(), skill);
    assert!(missing.is_empty(), "{missing:#?}");
    assert!(
        skill
            .lines()
            .any(|line| line == "name: agent-print-session")
    );
    let url = INSTALL_TEXT.lines().last().unwrap();
    assert!(url.ends_with("/skills/agent-print-session"));
    assert!(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/agent-print-session/SKILL.md")
            .is_file()
    );
}

const INSTALL_TEXT: &str = "Install the agent-print-session skill using the Skills CLI:\n\n  npx skills add alexgorbatchev/agent-print-session-cli --skill agent-print-session\n\nSkill:\n\n  https://github.com/alexgorbatchev/agent-print-session-cli/tree/main/skills/agent-print-session\n";
