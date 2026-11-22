package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindClaudeSession_EmptyTarget(t *testing.T) {
	_, err := FindClaudeSession("")
	if err == nil {
		t.Fatal("expected error for empty target, got nil")
	}
	_, err = FindClaudeSession("   ")
	if err == nil {
		t.Fatal("expected error for whitespace target, got nil")
	}
}

func TestFindClaudeSession_DirectFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session-123.jsonl")
	if err := os.WriteFile(filePath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	found, err := FindClaudeSession(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != filePath {
		t.Fatalf("expected %s, got %s", filePath, found)
	}
}

func TestFindClaudeSession_SearchDirectories(t *testing.T) {
	tmpRoot := t.TempDir()
	projectsDir := filepath.Join(tmpRoot, "projects", "my-project")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	exactFile := filepath.Join(projectsDir, "session-alpha-123.jsonl")
	if err := os.WriteFile(exactFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	prefixFile := filepath.Join(projectsDir, "session-beta-456.jsonl")
	if err := os.WriteFile(prefixFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	t.Setenv("CLAUDE_PROJECTS_DIR", filepath.Join(tmpRoot, "projects"))

	// 1. Exact match
	found, err := FindClaudeSession("session-alpha-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != exactFile {
		t.Fatalf("expected exact match %s, got %s", exactFile, found)
	}

	// 2. Prefix match
	found, err = FindClaudeSession("session-beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != prefixFile {
		t.Fatalf("expected prefix match %s, got %s", prefixFile, found)
	}

	// 3. Substring match
	found, err = FindClaudeSession("alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != exactFile {
		t.Fatalf("expected substring match %s, got %s", exactFile, found)
	}

	// 4. Not found
	_, err = FindClaudeSession("non-existent-session-id")
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}

func TestFindClaudeSession_SessionsDirAndXDG(t *testing.T) {
	tmpRoot := t.TempDir()
	sessDir := filepath.Join(tmpRoot, "sessions")
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sessFile := filepath.Join(sessDir, "ses-custom.jsonl")
	if err := os.WriteFile(sessFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_SESSIONS_DIR", sessDir)
	found, err := FindClaudeSession("ses-custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != sessFile {
		t.Fatalf("expected %s, got %s", sessFile, found)
	}

	// Test XDG
	xdgRoot := filepath.Join(tmpRoot, "xdg")
	xdgProjects := filepath.Join(xdgRoot, "ai-registry", "claude-code", "projects")
	if err := os.MkdirAll(xdgProjects, 0o755); err != nil {
		t.Fatal(err)
	}
	xdgFile := filepath.Join(xdgProjects, "ses-xdg.jsonl")
	if err := os.WriteFile(xdgFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_PROJECTS_DIR", "")
	t.Setenv("CLAUDE_SESSIONS_DIR", "")
	t.Setenv("XDG_DATA_HOME", xdgRoot)

	found, err = FindClaudeSession("ses-xdg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != xdgFile {
		t.Fatalf("expected %s, got %s", xdgFile, found)
	}
}

func TestFindTranscriptInDir_InvalidDir(t *testing.T) {
	res := findTranscriptInDir("/invalid/non/existent/path/for/sure", "anything")
	if res != "" {
		t.Fatalf("expected empty string, got %s", res)
	}
}

func TestFindPiSession_EmptyAndDirect(t *testing.T) {
	if _, err := FindPiSession(""); err == nil {
		t.Fatal("expected error on empty target, got nil")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pi-sess.jsonl")
	if err := os.WriteFile(filePath, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := FindPiSession(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != filePath {
		t.Fatalf("expected %s, got %s", filePath, found)
	}
}

func TestFindPiSession_SearchDirs(t *testing.T) {
	tmpRoot := t.TempDir()
	cwdDir := filepath.Join(tmpRoot, "--encoded-cwd--")
	if err := os.MkdirAll(cwdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// File with Pi timestamp prefix: 2026-09-25T21-07-10-103Z_<id>.jsonl
	piFile := filepath.Join(cwdDir, "2026-09-25T21-07-10-103Z_pi-test-123.jsonl")
	if err := os.WriteFile(piFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PI_SESSIONS_DIR", tmpRoot)

	// Exact session ID search
	found, err := FindPiSession("pi-test-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != piFile {
		t.Fatalf("expected %s, got %s", piFile, found)
	}

	// Not found
	if _, err := FindPiSession("non-existent-pi-id"); err == nil {
		t.Fatal("expected error on missing session, got nil")
	}
}

func TestFindPiSession_XDG(t *testing.T) {
	tmpRoot := t.TempDir()
	xdgDir := filepath.Join(tmpRoot, "ai-registry", "pi", "sessions")
	if err := os.MkdirAll(xdgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	xdgFile := filepath.Join(xdgDir, "xdg-pi-session.jsonl")
	if err := os.WriteFile(xdgFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PI_SESSIONS_DIR", "")
	t.Setenv("XDG_DATA_HOME", tmpRoot)

	found, err := FindPiSession("xdg-pi-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != xdgFile {
		t.Fatalf("expected %s, got %s", xdgFile, found)
	}
}

func TestFindCodexSession_EmptyAndDirect(t *testing.T) {
	if _, err := FindCodexSession(""); err == nil {
		t.Fatal("expected error on empty target, got nil")
	}
	if _, err := FindCodexSession("   "); err == nil {
		t.Fatal("expected error on whitespace target, got nil")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "rollout-2026-06-12T16-08-36-018f4a7c-1234-7000-8000-abcdef123456.jsonl")
	if err := os.WriteFile(filePath, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := FindCodexSession(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != filePath {
		t.Fatalf("expected %s, got %s", filePath, found)
	}
}

func TestFindCodexSession_SearchDirs(t *testing.T) {
	tmpRoot := t.TempDir()
	// Partitioned date directory: sessions/2026/06/12/
	dateDir := filepath.Join(tmpRoot, "sessions", "2026", "06", "12")
	if err := os.MkdirAll(dateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	codexFile := filepath.Join(dateDir, "rollout-2026-06-12T16-08-36-018f4a7c-1234-7000-8000-abcdef123456.jsonl")
	if err := os.WriteFile(codexFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CODEX_SESSIONS_DIR", filepath.Join(tmpRoot, "sessions"))

	// 1. Exact session ID search
	found, err := FindCodexSession("018f4a7c-1234-7000-8000-abcdef123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != codexFile {
		t.Fatalf("expected %s, got %s", codexFile, found)
	}

	// 2. Prefix search
	found, err = FindCodexSession("018f4a7c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != codexFile {
		t.Fatalf("expected %s, got %s", codexFile, found)
	}

	// 3. Full filename search (with .jsonl)
	found, err = FindCodexSession("rollout-2026-06-12T16-08-36-018f4a7c-1234-7000-8000-abcdef123456.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != codexFile {
		t.Fatalf("expected %s, got %s", codexFile, found)
	}

	// 4. Missing session
	if _, err := FindCodexSession("non-existent-codex-id"); err == nil {
		t.Fatal("expected error on missing session, got nil")
	}
}

func TestFindCodexSession_HomeAndXDG(t *testing.T) {
	tmpRoot := t.TempDir()

	// CODEX_HOME
	codexHome := filepath.Join(tmpRoot, "codex_home")
	codexHomeSess := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(codexHomeSess, 0o755); err != nil {
		t.Fatal(err)
	}
	homeFile := filepath.Join(codexHomeSess, "rollout-2026-01-01T00-00-00-home-test-123.jsonl")
	if err := os.WriteFile(homeFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CODEX_SESSIONS_DIR", "")
	t.Setenv("CODEX_HOME", codexHome)

	found, err := FindCodexSession("home-test-123")
	if err != nil {
		t.Fatalf("unexpected error on CODEX_HOME: %v", err)
	}
	if found != homeFile {
		t.Fatalf("expected %s, got %s", homeFile, found)
	}

	// XDG_DATA_HOME
	t.Setenv("CODEX_HOME", "")
	xdgDir := filepath.Join(tmpRoot, "ai-registry", "codex", "sessions")
	if err := os.MkdirAll(xdgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	xdgFile := filepath.Join(xdgDir, "rollout-2026-02-02T00-00-00-xdg-test-456.jsonl")
	if err := os.WriteFile(xdgFile, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_DATA_HOME", tmpRoot)

	found, err = FindCodexSession("xdg-test-456")
	if err != nil {
		t.Fatalf("unexpected error on XDG: %v", err)
	}
	if found != xdgFile {
		t.Fatalf("expected %s, got %s", xdgFile, found)
	}
}

func TestFindAgySession_Cases(t *testing.T) {
	if _, err := FindAgySession(""); err == nil {
		t.Fatal("expected error on empty target, got nil")
	}

	tmpRoot := t.TempDir()
	t.Setenv("HOME", tmpRoot)
	t.Setenv("GEMINI_HOME", "")
	t.Setenv("AGY_SESSIONS_DIR", "")

	// Direct file path
	directFile := filepath.Join(tmpRoot, "direct.jsonl")
	if err := os.WriteFile(directFile, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	found, err := FindAgySession(directFile)
	if err != nil || found != directFile {
		t.Fatalf("direct file: found=%q err=%v", found, err)
	}

	// Direct conversation directory preferring transcript_full.jsonl over transcript.jsonl
	directDir := filepath.Join(tmpRoot, "direct-conv")
	directLogs := filepath.Join(directDir, ".system_generated", "logs")
	if err := os.MkdirAll(directLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	directShort := filepath.Join(directLogs, "transcript.jsonl")
	directFull := filepath.Join(directLogs, "transcript_full.jsonl")
	if err := os.WriteFile(directShort, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directFull, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	found, err = FindAgySession(directDir)
	if err != nil || found != directFull {
		t.Fatalf("direct dir: found=%q want %q err=%v", found, directFull, err)
	}

	// Search AGY_SESSIONS_DIR: exact, prefix, substring, transcript.jsonl fallback
	brainDir := filepath.Join(tmpRoot, "agy-brain")
	exactLogs := filepath.Join(brainDir, "conv-alpha-111", ".system_generated", "logs")
	prefixLogs := filepath.Join(brainDir, "conv-beta-222", ".system_generated", "logs")
	if err := os.MkdirAll(exactLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prefixLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	exactPath := filepath.Join(exactLogs, "transcript_full.jsonl")
	prefixPath := filepath.Join(prefixLogs, "transcript.jsonl")
	if err := os.WriteFile(exactPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prefixPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AGY_SESSIONS_DIR", brainDir)
	if got, err := FindAgySession("conv-alpha-111"); err != nil || got != exactPath {
		t.Fatalf("exact match: got=%q want=%q err=%v", got, exactPath, err)
	}
	if got, err := FindAgySession("conv-beta"); err != nil || got != prefixPath {
		t.Fatalf("prefix match: got=%q want=%q err=%v", got, prefixPath, err)
	}
	if got, err := FindAgySession("alpha-111"); err != nil || got != exactPath {
		t.Fatalf("substring match: got=%q want=%q err=%v", got, exactPath, err)
	}
	if _, err := FindAgySession("missing-agy-id"); err == nil {
		t.Fatal("expected error for missing agy session")
	}

	// GEMINI_HOME fallback
	t.Setenv("AGY_SESSIONS_DIR", "")
	geminiHome := filepath.Join(tmpRoot, "gemini-home")
	geminiLogs := filepath.Join(geminiHome, "antigravity-cli", "brain", "conv-gemini-333", ".system_generated", "logs")
	if err := os.MkdirAll(geminiLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	geminiPath := filepath.Join(geminiLogs, "transcript_full.jsonl")
	if err := os.WriteFile(geminiPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GEMINI_HOME", geminiHome)
	if got, err := FindAgySession("conv-gemini-333"); err != nil || got != geminiPath {
		t.Fatalf("GEMINI_HOME match: got=%q want=%q err=%v", got, geminiPath, err)
	}
}
