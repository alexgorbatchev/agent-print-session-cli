package session

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alexgorbatchev/agent-parser"
)

func TestComputeSessionSummary_EmptyEvents(t *testing.T) {
	s := ComputeSessionSummary("/path/to/test.jsonl", nil)
	if s.FilePath != "/path/to/test.jsonl" {
		t.Errorf("FilePath = %s", s.FilePath)
	}
	if s.TotalEvents != 0 {
		t.Errorf("TotalEvents = %d", s.TotalEvents)
	}
}

func TestComputeSessionSummary_RichEvents(t *testing.T) {
	evs := []parser.ParsedEvent{
		{
			EventType: "run_started",
			Timestamp: 1000,
			CWD:       "/test/cwd",
			Data: parser.EventData{
				Title: func() *string { v := "Rich Test Session"; return &v }(),
				Model: func() *string { v := "claude-3-7-sonnet"; return &v }(),
			},
		},
		{
			EventType: "turn_start",
			Timestamp: 2000,
			Data: parser.EventData{
				Content: func() *string { v := "Do something"; return &v }(),
			},
		},
		{
			EventType: "tool_call",
			Timestamp: 3000,
			Data: parser.EventData{
				ToolName:   func() *string { v := "Edit"; return &v }(),
				ToolCallID: func() *string { v := "call_1"; return &v }(),
				FileModification: &parser.FileModification{
					Action:   "EDIT",
					FilePath: "src/main.go",
					Hunks:    make([]parser.DiffHunk, 2),
				},
			},
		},
		{
			EventType: "tool_call",
			Timestamp: 3500,
			Data: parser.EventData{
				ToolName:   func() *string { v := "Edit"; return &v }(),
				ToolCallID: func() *string { v := "call_1b"; return &v }(),
				FileModification: &parser.FileModification{
					Action:   "EDIT",
					FilePath: "src/main.go",
					Hunks:    make([]parser.DiffHunk, 1),
				},
			},
		},
		{
			EventType: "tool_result",
			Timestamp: 4000,
			Data: parser.EventData{
				ToolCallID: func() *string { v := "call_1"; return &v }(),
				IsError:    func() *bool { v := true; return &v }(),
				ToolOutput: func() *string { v := "File modification failed"; return &v }(),
			},
		},
		{
			EventType: "turn_end",
			Timestamp: 5000,
			Data: parser.EventData{
				Usage: &parser.TokenUsage{
					Input:     100,
					Output:    50,
					CacheRead: 25,
					Total:     175,
				},
				RawHarnessCost: func() *float64 { v := 0.05; return &v }(),
			},
		},
		{
			EventType: "error",
			Timestamp: 6000,
			Data: parser.EventData{
				Content: func() *string { v := "Harness crash"; return &v }(),
			},
		},
		{
			EventType: "run_exited",
			Timestamp: 7000,
			Data: parser.EventData{
				ExitCode: func() *int { v := 1; return &v }(),
			},
		},
	}

	s := ComputeSessionSummary("/test/session.jsonl", evs)

	if s.Title != "Rich Test Session" {
		t.Errorf("Title = %q", s.Title)
	}
	if s.CWD != "/test/cwd" {
		t.Errorf("CWD = %q", s.CWD)
	}
	if s.TurnCount != 1 {
		t.Errorf("TurnCount = %d", s.TurnCount)
	}
	if s.DurationMs != 6000 {
		t.Errorf("DurationMs = %d, want 6000", s.DurationMs)
	}
	if len(s.Models) != 1 || s.Models[0] != "claude-3-7-sonnet" {
		t.Errorf("Models = %v", s.Models)
	}
	if s.TokensTotal != 175 {
		t.Errorf("TokensTotal = %d", s.TokensTotal)
	}
	if s.CostUSD != 0.05 {
		t.Errorf("CostUSD = %f", s.CostUSD)
	}
	if len(s.Tools) != 1 || s.Tools[0].Name != "Edit" || s.Tools[0].Calls != 2 || s.Tools[0].ErrorCalls != 1 {
		t.Errorf("Tools = %+v", s.Tools)
	}
	if len(s.FilesModified) != 1 || s.FilesModified[0].Hunks != 3 {
		t.Errorf("FilesModified = %+v", s.FilesModified)
	}
	if len(s.Errors) != 2 {
		t.Errorf("Errors = %+v, want 2", s.Errors)
	}
	if s.ExitCode == nil || *s.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", s.ExitCode)
	}
}

func TestRenderSummary_Formats(t *testing.T) {
	exit := 0
	s := SessionSummary{
		FilePath:    "/path/to/test.jsonl",
		Title:       "Test Session",
		CWD:         "/workspace",
		DurationMs:  125000, // 2m05s
		TotalEvents: 10,
		TurnCount:   2,
		Models:      []string{"claude-3-7-sonnet"},
		TokensIn:    100,
		TokensOut:   50,
		TokensCache: 20,
		TokensTotal: 170,
		CostUSD:     0.02,
		Tools: []ToolUsageStat{
			{Name: "Bash", Calls: 5, ErrorCalls: 1},
		},
		FilesModified: []FileModificationStat{
			{FilePath: "main.go", Action: "EDIT", Hunks: 2},
		},
		Errors: []ErrorRecord{
			{Timestamp: 1000, Source: "tool_call (call_1)", Message: "exit code 1"},
		},
		ExitCode: &exit,
	}

	// 1. JSON output
	var jsonBuf bytes.Buffer
	if err := RenderSummary(&jsonBuf, s, false, true); err != nil {
		t.Fatalf("JSON RenderSummary error: %v", err)
	}
	var parsed SessionSummary
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.Title != "Test Session" {
		t.Errorf("parsed JSON Title = %s", parsed.Title)
	}

	// 2. Human output
	var humanBuf bytes.Buffer
	if err := RenderSummary(&humanBuf, s, false, false); err != nil {
		t.Fatalf("Human RenderSummary error: %v", err)
	}
	humanOut := humanBuf.String()
	if !strings.Contains(humanOut, "[SESSION SUMMARY]") {
		t.Errorf("expected [SESSION SUMMARY] in human mode, got: %s", humanOut)
	}
	if !strings.Contains(humanOut, "2m05s") {
		t.Errorf("expected duration 2m05s, got: %s", humanOut)
	}

	// 3. Agent output
	var agentBuf bytes.Buffer
	if err := RenderSummary(&agentBuf, s, true, false); err != nil {
		t.Fatalf("Agent RenderSummary error: %v", err)
	}
	agentOut := agentBuf.String()
	if !strings.Contains(agentOut, "SESSION_SUMMARY") {
		t.Errorf("expected SESSION_SUMMARY in agent mode, got: %s", agentOut)
	}
	if strings.Contains(agentOut, "--------------------------------------------------------------------------------") {
		t.Errorf("dividers forbidden in agent mode: %s", agentOut)
	}
	if !strings.Contains(agentOut, "Bash(5,err:1)") {
		t.Errorf("expected Bash(5,err:1) in tools list: %s", agentOut)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0s"},
		{-10, "0s"},
		{5000, "5.0s"},
		{65000, "1m05s"},
		{3665000, "1h01m"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.ms)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func TestComputeSessionSummary_AttributesToolErrorsByCallID(t *testing.T) {
	str := func(v string) *string { return &v }
	isErr := func(v bool) *bool { return &v }
	call := func(name, id string) parser.ParsedEvent {
		return parser.ParsedEvent{EventType: "tool_call", Data: parser.EventData{ToolName: str(name), ToolCallID: str(id)}}
	}
	result := func(id string, failed bool) parser.ParsedEvent {
		return parser.ParsedEvent{EventType: "tool_result", Data: parser.EventData{ToolCallID: str(id), IsError: isErr(failed)}}
	}
	evs := []parser.ParsedEvent{
		call("Bash", "b1"), call("Bash", "b2"), call("Bash", "b3"), call("Read", "r1"),
		result("b1", true), result("b2", true), result("b3", false), result("r1", false),
		result("orphan", true),
	}

	s := ComputeSessionSummary("/test/session.jsonl", evs)

	got := map[string]ToolUsageStat{}
	for _, tool := range s.Tools {
		got[tool.Name] = tool
	}
	if got["Bash"].Calls != 3 || got["Bash"].ErrorCalls != 2 {
		t.Errorf("Bash = %+v, want 3 calls with 2 errors", got["Bash"])
	}
	if got["Read"].Calls != 1 || got["Read"].ErrorCalls != 0 {
		t.Errorf("Read = %+v, want 1 call with 0 errors", got["Read"])
	}
	if len(s.Errors) != 3 {
		t.Errorf("Errors = %+v, want 3 including the unmatched result", s.Errors)
	}

	var agentOut strings.Builder
	if err := RenderSummary(&agentOut, s, true, false); err != nil {
		t.Fatalf("RenderSummary: %v", err)
	}
	if !strings.Contains(agentOut.String(), "Bash(3,err:2)") {
		t.Errorf("agent summary missing per-tool error count:\n%s", agentOut.String())
	}
}

// codexRolloutEvents parses a Codex rollout in the shapes codex-rs writes.
func codexRolloutEvents(t *testing.T, lines []string) []parser.ParsedEvent {
	t.Helper()
	p := parser.NewCodexParser()
	var events []parser.ParsedEvent
	for _, line := range lines {
		evs, err := p.ParseLine([]byte(line))
		if err != nil {
			t.Fatalf("ParseLine(%s): %v", line, err)
		}
		events = append(events, evs...)
	}
	return events
}

func TestComputeSessionSummary_CodexRollout(t *testing.T) {
	events := codexRolloutEvents(t, []string{
		`{"timestamp":"2026-11-07T10:00:00Z","type":"session_meta","payload":{"id":"t","cwd":"/work","source":"cli","history_mode":"paginated"}}`,
		`{"timestamp":"2026-11-07T10:00:01Z","type":"event_msg","payload":{"type":"task_started","turn_id":"u1","started_at":1794045601}}`,
		`{"timestamp":"2026-11-07T10:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"Fix the build"}]}}`,
		`{"timestamp":"2026-11-07T10:00:01Z","type":"event_msg","payload":{"type":"item_completed","item":{"type":"UserMessage","id":"m1","content":[{"type":"text","text":"Fix the build","text_elements":[]}]}}}`,
		// Code-mode scripts: one fails, one succeeds although its output mentions an error.
		`{"timestamp":"2026-11-07T10:00:02Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c1","input":"run()"}}`,
		`{"timestamp":"2026-11-07T10:00:03Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c1","output":[{"type":"input_text","text":"Script failed\nWall time 0.1 seconds\nOutput:\n"},{"type":"input_text","text":"Script error:\nboom"}]}}`,
		`{"timestamp":"2026-11-07T10:00:04Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c2","input":"run()"}}`,
		`{"timestamp":"2026-11-07T10:00:05Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c2","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds\nOutput:\n"},{"type":"input_text","text":"error: none found"}]}}`,
		`{"timestamp":"2026-11-07T10:00:06Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":1000,"cached_input_tokens":400,"output_tokens":50,"reasoning_output_tokens":10,"total_tokens":1050},"last_token_usage":{"input_tokens":1000,"cached_input_tokens":400,"output_tokens":50,"reasoning_output_tokens":10,"total_tokens":1050},"model_context_window":258400}}}`,
		// An applied patch that changed two files.
		`{"timestamp":"2026-11-07T10:00:07Z","type":"event_msg","payload":{"type":"item_completed","item":{"type":"FileChange","id":"c3","changes":{"/work/a.go":{"type":"update","unified_diff":"@@ -1 +1 @@\n-a\n+b\n","move_path":null},"/work/b.go":{"type":"add","content":"package work\n"}},"status":"completed"}}}`,
		`{"timestamp":"2026-11-07T10:00:09Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"u1","completed_at":1794045609,"last_agent_message":"Done."}}`,
	})

	s := ComputeSessionSummary("/work/rollout.jsonl", events)

	if s.TurnCount != 1 {
		t.Errorf("TurnCount = %d, want 1", s.TurnCount)
	}
	if len(s.Tools) != 1 || s.Tools[0].Name != "exec" || s.Tools[0].Calls != 2 || s.Tools[0].ErrorCalls != 1 {
		t.Errorf("Tools = %+v, want exec with 2 calls and 1 failure", s.Tools)
	}
	if s.TokensIn != 600 || s.TokensOut != 50 || s.TokensCache != 400 || s.TokensTotal != 1050 {
		t.Errorf("tokens = in %d out %d cache %d total %d, want 600/50/400/1050", s.TokensIn, s.TokensOut, s.TokensCache, s.TokensTotal)
	}
	if len(s.FilesModified) != 2 {
		t.Errorf("FilesModified = %+v, want both patched files", s.FilesModified)
	}

	h := ComputeHandoffContext("/work/rollout.jsonl", events, 10)
	if len(h.FilesModified) != 2 {
		t.Errorf("handoff FilesModified = %+v, want both patched files", h.FilesModified)
	}
}
