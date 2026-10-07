package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleClaudeLog = `{"type":"ai-title","title":"Test session","sessionId":"cli-test-123"}
{"type":"user","message":{"role":"user","content":"Hello world"},"uuid":"u-1","sessionId":"cli-test-123","timestamp":"2026-09-20T12:00:00.000Z"}
{"type":"assistant","message":{"id":"msg-1","role":"assistant","model":"claude-3-7-sonnet","content":[{"type":"text","text":"Hi! How can I help you today?"}],"usage":{"input_tokens":10,"output_tokens":12}},"uuid":"a-1","sessionId":"cli-test-123","timestamp":"2026-09-20T12:00:01.000Z"}
`

func setupSampleLog(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "cli-test-123.jsonl")
	if err := os.WriteFile(p, []byte(sampleClaudeLog), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return p
}

func TestCLI_Help(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected nil error on --help, got: %v", err)
	}

	helpText := out.String()
	if !strings.Contains(helpText, "agent-print-session") {
		t.Errorf("expected 'agent-print-session' in help, got: %s", helpText)
	}
	if !strings.Contains(helpText, "claude") {
		t.Errorf("expected 'claude' in help tree, got: %s", helpText)
	}
	if !strings.Contains(helpText, "codex") {
		t.Errorf("expected 'codex' in help tree, got: %s", helpText)
	}
}

func TestCLI_ClaudeHelp(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected nil error on claude --help, got: %v", err)
	}

	helpText := out.String()
	if !strings.Contains(helpText, "print") {
		t.Errorf("expected 'print' in claude help, got: %s", helpText)
	}
}

func TestCLI_Print_Success(t *testing.T) {
	t.Setenv("AGENT", "0")
	logPath := setupSampleLog(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "print", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing print: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "[SESSION LOG]") || !strings.Contains(outputStr, "Hello world") {
		t.Errorf("expected formatted log output, got:\n%s", outputStr)
	}
}

func TestCLI_Claude_ShorthandInvocation(t *testing.T) {
	t.Setenv("AGENT", "0")
	logPath := setupSampleLog(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on shorthand claude <path>: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "[SESSION LOG]") {
		t.Errorf("expected session log output, got:\n%s", outputStr)
	}
}

func TestCLI_Print_JSONFlag(t *testing.T) {
	logPath := setupSampleLog(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "print", "--json", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on --json: %v", err)
	}

	outputStr := out.String()
	if !strings.HasPrefix(strings.TrimSpace(outputStr), "[") {
		t.Errorf("expected JSON array output, got:\n%s", outputStr)
	}
}

func TestCLI_Print_PathFlag(t *testing.T) {
	t.Setenv("AGENT", "0")
	logPath := setupSampleLog(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "print", "--path", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on --path flag: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "[SESSION LOG]") {
		t.Errorf("expected session log output, got:\n%s", outputStr)
	}
}

func TestCLI_Print_MissingArgError(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"claude", "print"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on missing args, got nil")
	}
}

func TestCLI_Print_NotFound(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"claude", "print", "non-existent-session-id-xyz"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent session ID, got nil")
	}
}

func TestCLI_AgentMode(t *testing.T) {
	logPath := setupSampleLog(t)
	t.Setenv("AGENT", "1")

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "print", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in agent mode: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "SESSION:") || !strings.Contains(outputStr, "USER:") {
		t.Errorf("expected agent mode conversational output, got:\n%s", outputStr)
	}
	if strings.Contains(outputStr, "--------------------------------------------------------------------------------") {
		t.Errorf("separators forbidden in agent mode: %s", outputStr)
	}
}

func TestRunHelper(t *testing.T) {
	// Test the run function with --help
	err := run([]string{"--help"})
	if err != nil {
		t.Errorf("run([--help]) = %v, want nil", err)
	}
}

func TestCLI_RootNoArgs(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected nil error on root no args, got: %v", err)
	}
}

func TestCLI_ClaudeNoArgs(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected nil error on claude no args, got: %v", err)
	}
}

func TestMainFunc(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"agent-print-session", "--version"}
	main()
}

func TestCLI_ClaudeSummary(t *testing.T) {
	logPath := setupSampleLog(t)

	// 1. Claude summary command
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "summary", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on claude summary: %v", err)
	}
	if !strings.Contains(out.String(), "[SESSION SUMMARY]") && !strings.Contains(out.String(), "SESSION_SUMMARY") {
		t.Errorf("expected session summary output, got: %s", out.String())
	}

	// 2. Claude summary JSON
	var jsonOut bytes.Buffer
	cmdJSON := newRootCmd()
	cmdJSON.SetOut(&jsonOut)
	cmdJSON.SetArgs([]string{"claude", "summary", "--json", logPath})
	if err := cmdJSON.Execute(); err != nil {
		t.Fatalf("unexpected error on claude summary --json: %v", err)
	}
	if !strings.Contains(jsonOut.String(), `"totalEvents"`) {
		t.Errorf("expected JSON summary, got: %s", jsonOut.String())
	}

	// 3. Claude summary with --path flag
	var pathOut bytes.Buffer
	cmdPath := newRootCmd()
	cmdPath.SetOut(&pathOut)
	cmdPath.SetArgs([]string{"claude", "summary", "--path", logPath})
	if err := cmdPath.Execute(); err != nil {
		t.Fatalf("unexpected error on claude summary --path: %v", err)
	}

	// 4. Claude summary missing args
	cmdMissing := newRootCmd()
	cmdMissing.SetArgs([]string{"claude", "summary"})
	if err := cmdMissing.Execute(); err == nil {
		t.Fatal("expected error on claude summary without args, got nil")
	}
}

func TestCLI_ClaudeHandoff(t *testing.T) {
	logPath := setupSampleLog(t)

	// 1. Claude handoff command
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"claude", "handoff", logPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on claude handoff: %v", err)
	}
	if !strings.Contains(out.String(), "HANDOFF") && !strings.Contains(out.String(), "CONTINUATION_CONTEXT") {
		t.Errorf("expected handoff output, got: %s", out.String())
	}

	// 2. Claude continue alias
	cmdCont := newRootCmd()
	var outCont bytes.Buffer
	cmdCont.SetOut(&outCont)
	cmdCont.SetArgs([]string{"claude", "continue", logPath})
	if err := cmdCont.Execute(); err != nil {
		t.Fatalf("unexpected error on claude continue: %v", err)
	}

	// 3. Claude resume alias
	cmdRes := newRootCmd()
	var outRes bytes.Buffer
	cmdRes.SetOut(&outRes)
	cmdRes.SetArgs([]string{"claude", "resume", logPath})
	if err := cmdRes.Execute(); err != nil {
		t.Fatalf("unexpected error on claude resume: %v", err)
	}

	// 4. Claude handoff --json
	cmdJSON := newRootCmd()
	var outJSON bytes.Buffer
	cmdJSON.SetOut(&outJSON)
	cmdJSON.SetArgs([]string{"claude", "handoff", "--json", logPath})
	if err := cmdJSON.Execute(); err != nil {
		t.Fatalf("unexpected error on claude handoff --json: %v", err)
	}
	if !strings.Contains(outJSON.String(), `"initialObjective"`) {
		t.Errorf("expected JSON handoff, got: %s", outJSON.String())
	}

	// 5. Claude handoff --path
	cmdPath := newRootCmd()
	var outPath bytes.Buffer
	cmdPath.SetOut(&outPath)
	cmdPath.SetArgs([]string{"claude", "handoff", "--path", logPath})
	if err := cmdPath.Execute(); err != nil {
		t.Fatalf("unexpected error on claude handoff --path: %v", err)
	}

	// 6. Claude handoff missing args
	cmdMissing := newRootCmd()
	cmdMissing.SetArgs([]string{"claude", "handoff"})
	if err := cmdMissing.Execute(); err == nil {
		t.Fatal("expected error on claude handoff without args, got nil")
	}
}

func TestMain_ErrorExit(t *testing.T) {
	oldArgs := os.Args
	oldExit := exitFunc
	defer func() {
		os.Args = oldArgs
		exitFunc = oldExit
	}()
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}
	os.Args = []string{"agent-print-session", "unknown-command-xyz"}
	main()
	if !exitCalled {
		t.Errorf("expected exitFunc to be called on error")
	}
}

func TestCLI_PrintFilterFlags(t *testing.T) {
	logPath := setupSampleLog(t)

	testCases := []struct {
		name string
		args []string
	}{
		{"summary-flag", []string{"claude", "print", "--summary", logPath}},
		{"handoff-flag", []string{"claude", "print", "--handoff", logPath}},
		{"last-turn-flag", []string{"claude", "print", "--last-turn", logPath}},
		{"errors-flag", []string{"claude", "print", "--errors", logPath}},
		{"files-flag", []string{"claude", "print", "--files", logPath}},
		{"tools-flag", []string{"claude", "print", "--tools", logPath}},
		{"prompts-flag", []string{"claude", "print", "--prompts", logPath}},
		{"tail-flag", []string{"claude", "print", "--tail", "1", logPath}},
		{"limit-flag", []string{"claude", "print", "--limit", "1", logPath}},
		{"all-flag", []string{"claude", "print", "--all", logPath}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error executing %v: %v", tc.args, err)
			}
		})
	}
}

const samplePiLog = `{"type":"session","version":3,"id":"pi-cli-test","timestamp":"2026-09-25T21:07:10.103Z","cwd":"/repo"}
{"type":"message","id":"m1","timestamp":"2026-09-25T21:07:15.131Z","message":{"role":"user","content":[{"type":"text","text":"fix bug in repo"}]}}
{"type":"message","id":"m2","timestamp":"2026-09-25T21:07:18.000Z","message":{"role":"assistant","content":[{"type":"text","text":"Done with fix"}]}}
`

func setupSamplePiLog(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "pi-cli-test.jsonl")
	if err := os.WriteFile(p, []byte(samplePiLog), 0o644); err != nil {
		t.Fatalf("failed to write pi test file: %v", err)
	}
	return p
}

func TestCLI_PiCommands(t *testing.T) {
	piPath := setupSamplePiLog(t)

	// 1. Pi help
	cmdHelp := newRootCmd()
	var outHelp bytes.Buffer
	cmdHelp.SetOut(&outHelp)
	cmdHelp.SetArgs([]string{"pi", "--help"})
	if err := cmdHelp.Execute(); err != nil {
		t.Fatalf("expected nil error on pi --help: %v", err)
	}

	// 2. Pi no args (help)
	cmdNoArgs := newRootCmd()
	var outNoArgs bytes.Buffer
	cmdNoArgs.SetOut(&outNoArgs)
	cmdNoArgs.SetArgs([]string{"pi"})
	if err := cmdNoArgs.Execute(); err != nil {
		t.Fatalf("expected nil error on pi no args: %v", err)
	}

	// 3. Pi shorthand invocation: pi <path>
	cmdShort := newRootCmd()
	var outShort bytes.Buffer
	cmdShort.SetOut(&outShort)
	cmdShort.SetArgs([]string{"pi", piPath})
	if err := cmdShort.Execute(); err != nil {
		t.Fatalf("unexpected error on pi shorthand: %v", err)
	}
	if !strings.Contains(outShort.String(), "fix bug in repo") {
		t.Errorf("expected prompt in pi shorthand output: %s", outShort.String())
	}

	// 4. Pi print command: pi print <path>
	cmdPrint := newRootCmd()
	var outPrint bytes.Buffer
	cmdPrint.SetOut(&outPrint)
	cmdPrint.SetArgs([]string{"pi", "print", piPath})
	if err := cmdPrint.Execute(); err != nil {
		t.Fatalf("unexpected error on pi print: %v", err)
	}
	if !strings.Contains(outPrint.String(), "Done with fix") {
		t.Errorf("expected assistant text in pi print output: %s", outPrint.String())
	}

	// 5. Pi summary: pi summary <path>
	cmdSum := newRootCmd()
	var outSum bytes.Buffer
	cmdSum.SetOut(&outSum)
	cmdSum.SetArgs([]string{"pi", "summary", piPath})
	if err := cmdSum.Execute(); err != nil {
		t.Fatalf("unexpected error on pi summary: %v", err)
	}
	if !strings.Contains(outSum.String(), "SESSION_SUMMARY") && !strings.Contains(outSum.String(), "[SESSION SUMMARY]") {
		t.Errorf("expected summary header, got: %s", outSum.String())
	}

	// 6. Pi handoff: pi handoff <path>
	cmdHandoff := newRootCmd()
	var outHandoff bytes.Buffer
	cmdHandoff.SetOut(&outHandoff)
	cmdHandoff.SetArgs([]string{"pi", "handoff", piPath})
	if err := cmdHandoff.Execute(); err != nil {
		t.Fatalf("unexpected error on pi handoff: %v", err)
	}
	if !strings.Contains(outHandoff.String(), "CONTINUATION_CONTEXT") && !strings.Contains(outHandoff.String(), "HANDOFF") {
		t.Errorf("expected handoff header, got: %s", outHandoff.String())
	}

	// 7. Pi missing args on print
	cmdMissing := newRootCmd()
	cmdMissing.SetArgs([]string{"pi", "print"})
	if err := cmdMissing.Execute(); err == nil {
		t.Fatal("expected error on pi print without args, got nil")
	}

	// 8. Pi missing args on summary
	cmdSumMissing := newRootCmd()
	cmdSumMissing.SetArgs([]string{"pi", "summary"})
	if err := cmdSumMissing.Execute(); err == nil {
		t.Fatal("expected error on pi summary without args, got nil")
	}

	// 9. Pi missing args on handoff
	cmdHandMissing := newRootCmd()
	cmdHandMissing.SetArgs([]string{"pi", "handoff"})
	if err := cmdHandMissing.Execute(); err == nil {
		t.Fatal("expected error on pi handoff without args, got nil")
	}

	// 10. Pi not found
	cmdNotFound := newRootCmd()
	cmdNotFound.SetArgs([]string{"pi", "print", "non-existent-pi-id"})
	if err := cmdNotFound.Execute(); err == nil {
		t.Fatal("expected error on non-existent pi session, got nil")
	}
}

const sampleCodexLog = `{"type":"session_meta","timestamp":"2026-06-12T16:08:36.123Z","payload":{"id":"codex-cli-test","timestamp":"2026-06-12T16:08:36.123Z","cwd":"/repo","originator":"codex-tui","cli_version":"0.121.0","source":"cli","model_provider":"openai","git":{"commit_hash":"abcdef123","branch":"main","repository_url":"git@github.com:org/repo.git"}}}
{"type":"turn_context","timestamp":"2026-06-12T16:08:37.000Z","payload":{"turn_id":"turn-1","cwd":"/repo","model":"o3-mini"}}
{"type":"event_msg","timestamp":"2026-06-12T16:08:38.000Z","payload":{"type":"user_message","message":"implement codex support"}}
{"type":"event_msg","timestamp":"2026-06-12T16:08:41.000Z","payload":{"type":"agent_message","message":"Codex support implemented"}}
`

func setupSampleCodexLog(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "rollout-2026-06-12T16-08-36-codex-cli-test.jsonl")
	if err := os.WriteFile(p, []byte(sampleCodexLog), 0o644); err != nil {
		t.Fatalf("failed to write codex test file: %v", err)
	}
	return p
}

func TestCLI_CodexCommands(t *testing.T) {
	codexPath := setupSampleCodexLog(t)

	// 1. Codex help
	cmdHelp := newRootCmd()
	var outHelp bytes.Buffer
	cmdHelp.SetOut(&outHelp)
	cmdHelp.SetArgs([]string{"codex", "--help"})
	if err := cmdHelp.Execute(); err != nil {
		t.Fatalf("expected nil error on codex --help: %v", err)
	}

	// 2. Codex no args (help)
	cmdNoArgs := newRootCmd()
	var outNoArgs bytes.Buffer
	cmdNoArgs.SetOut(&outNoArgs)
	cmdNoArgs.SetArgs([]string{"codex"})
	if err := cmdNoArgs.Execute(); err != nil {
		t.Fatalf("expected nil error on codex no args: %v", err)
	}

	// 3. Codex shorthand invocation: codex <path>
	cmdShort := newRootCmd()
	var outShort bytes.Buffer
	cmdShort.SetOut(&outShort)
	cmdShort.SetArgs([]string{"codex", codexPath})
	if err := cmdShort.Execute(); err != nil {
		t.Fatalf("unexpected error on codex shorthand: %v", err)
	}
	if !strings.Contains(outShort.String(), "implement codex support") {
		t.Errorf("expected prompt in codex shorthand output: %s", outShort.String())
	}

	// 4. Codex print command: codex print <path>
	cmdPrint := newRootCmd()
	var outPrint bytes.Buffer
	cmdPrint.SetOut(&outPrint)
	cmdPrint.SetArgs([]string{"codex", "print", codexPath})
	if err := cmdPrint.Execute(); err != nil {
		t.Fatalf("unexpected error on codex print: %v", err)
	}
	if !strings.Contains(outPrint.String(), "Codex support implemented") {
		t.Errorf("expected assistant text in codex print output: %s", outPrint.String())
	}

	// 5. Codex summary: codex summary <path>
	cmdSum := newRootCmd()
	var outSum bytes.Buffer
	cmdSum.SetOut(&outSum)
	cmdSum.SetArgs([]string{"codex", "summary", codexPath})
	if err := cmdSum.Execute(); err != nil {
		t.Fatalf("unexpected error on codex summary: %v", err)
	}
	if !strings.Contains(outSum.String(), "SESSION_SUMMARY") && !strings.Contains(outSum.String(), "[SESSION SUMMARY]") {
		t.Errorf("expected summary header, got: %s", outSum.String())
	}

	// 5b. Codex summary with --json and --path
	cmdSumJSON := newRootCmd()
	var outSumJSON bytes.Buffer
	cmdSumJSON.SetOut(&outSumJSON)
	cmdSumJSON.SetArgs([]string{"codex", "summary", "--json", "--path", codexPath})
	if err := cmdSumJSON.Execute(); err != nil {
		t.Fatalf("unexpected error on codex summary --json --path: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(outSumJSON.String()), "{") {
		t.Errorf("expected JSON object from summary, got: %s", outSumJSON.String())
	}

	// 6. Codex handoff: codex handoff <path>
	cmdHandoff := newRootCmd()
	var outHandoff bytes.Buffer
	cmdHandoff.SetOut(&outHandoff)
	cmdHandoff.SetArgs([]string{"codex", "handoff", codexPath})
	if err := cmdHandoff.Execute(); err != nil {
		t.Fatalf("unexpected error on codex handoff: %v", err)
	}
	if !strings.Contains(outHandoff.String(), "CONTINUATION_CONTEXT") && !strings.Contains(outHandoff.String(), "HANDOFF") {
		t.Errorf("expected handoff header, got: %s", outHandoff.String())
	}

	// 6b. Codex handoff with --json and --path
	cmdHandoffJSON := newRootCmd()
	var outHandoffJSON bytes.Buffer
	cmdHandoffJSON.SetOut(&outHandoffJSON)
	cmdHandoffJSON.SetArgs([]string{"codex", "handoff", "--json", "--path", codexPath})
	if err := cmdHandoffJSON.Execute(); err != nil {
		t.Fatalf("unexpected error on codex handoff --json --path: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(outHandoffJSON.String()), "{") {
		t.Errorf("expected JSON object from handoff, got: %s", outHandoffJSON.String())
	}

	// 6c. Codex print with --path and AGENT=1
	t.Setenv("AGENT", "1")
	cmdAgent := newRootCmd()
	var outAgent bytes.Buffer
	cmdAgent.SetOut(&outAgent)
	cmdAgent.SetArgs([]string{"codex", "print", "--path", codexPath})
	if err := cmdAgent.Execute(); err != nil {
		t.Fatalf("unexpected error on codex print agent mode: %v", err)
	}
	if !strings.Contains(outAgent.String(), "SESSION:") {
		t.Errorf("expected SESSION header in agent mode, got: %s", outAgent.String())
	}
	t.Setenv("AGENT", "")

	// 7. Codex missing args on print
	cmdMissing := newRootCmd()
	cmdMissing.SetArgs([]string{"codex", "print"})
	if err := cmdMissing.Execute(); err == nil {
		t.Fatal("expected error on codex print without args, got nil")
	}

	// 8. Codex missing args on summary
	cmdSumMissing := newRootCmd()
	cmdSumMissing.SetArgs([]string{"codex", "summary"})
	if err := cmdSumMissing.Execute(); err == nil {
		t.Fatal("expected error on codex summary without args, got nil")
	}

	// 9. Codex missing args on handoff
	cmdHandMissing := newRootCmd()
	cmdHandMissing.SetArgs([]string{"codex", "handoff"})
	if err := cmdHandMissing.Execute(); err == nil {
		t.Fatal("expected error on codex handoff without args, got nil")
	}

	// 10. Codex not found
	cmdNotFound := newRootCmd()
	cmdNotFound.SetArgs([]string{"codex", "print", "non-existent-codex-id"})
	if err := cmdNotFound.Execute(); err == nil {
		t.Fatal("expected error on non-existent codex session, got nil")
	}
}
