package session

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindClaudeSession locates a Claude Code transcript on disk given an exact file path,
// a session ID, or a partial session ID.
func FindClaudeSession(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("session ID or file path cannot be empty")
	}

	clean := filepath.Clean(trimmed)
	if fi, err := os.Stat(clean); err == nil && !fi.IsDir() {
		return clean, nil
	}

	searchDirs := getClaudeSearchDirs()
	for _, dir := range searchDirs {
		if match := findTranscriptInDir(dir, trimmed); match != "" {
			return match, nil
		}
	}

	return "", fmt.Errorf("claude session %q not found as a file or in search directories (~/.claude/projects, ~/.claude/sessions)", target)
}

func getClaudeSearchDirs() []string {
	var dirs []string

	if custom := os.Getenv("CLAUDE_PROJECTS_DIR"); custom != "" {
		dirs = append(dirs, custom)
	}
	if custom := os.Getenv("CLAUDE_SESSIONS_DIR"); custom != "" {
		dirs = append(dirs, custom)
	}

	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		dirs = append(dirs,
			filepath.Join(xdgData, "ai-registry", "claude-code", "projects"),
			filepath.Join(xdgData, "ai-registry", "claude-code", "sessions"),
		)
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".claude", "projects"),
			filepath.Join(home, ".claude", "sessions"),
			filepath.Join(home, ".local", "share", "ai-registry", "claude-code", "projects"),
			filepath.Join(home, ".local", "share", "ai-registry", "claude-code", "sessions"),
		)
	}

	return dirs
}

// FindPiSession locates a Pi session transcript on disk given an exact file path,
// a session ID, or a partial session ID.
func FindPiSession(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("session ID or file path cannot be empty")
	}

	clean := filepath.Clean(trimmed)
	if fi, err := os.Stat(clean); err == nil && !fi.IsDir() {
		return clean, nil
	}

	searchDirs := getPiSearchDirs()
	for _, dir := range searchDirs {
		if match := findTranscriptInDir(dir, trimmed); match != "" {
			return match, nil
		}
	}

	return "", fmt.Errorf("pi session %q not found as a file or in search directories (~/.pi/agent/sessions)", target)
}

func getPiSearchDirs() []string {
	var dirs []string

	if custom := os.Getenv("PI_SESSIONS_DIR"); custom != "" {
		dirs = append(dirs, custom)
	}

	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		dirs = append(dirs,
			filepath.Join(xdgData, "ai-registry", "pi", "sessions"),
		)
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".pi", "agent", "sessions"),
			filepath.Join(home, ".local", "share", "ai-registry", "pi", "sessions"),
		)
	}

	return dirs
}

// FindCodexSession locates an OpenAI Codex session transcript on disk given an exact file path,
// a session ID, or a partial session ID.
func FindCodexSession(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("session ID or file path cannot be empty")
	}

	clean := filepath.Clean(trimmed)
	if fi, err := os.Stat(clean); err == nil && !fi.IsDir() {
		return clean, nil
	}

	searchDirs := getCodexSearchDirs()
	for _, dir := range searchDirs {
		if match := findTranscriptInDir(dir, trimmed); match != "" {
			return match, nil
		}
	}

	return "", fmt.Errorf("codex session %q not found as a file or in search directories (~/.codex/sessions)", target)
}

func getCodexSearchDirs() []string {
	var dirs []string

	if custom := os.Getenv("CODEX_SESSIONS_DIR"); custom != "" {
		dirs = append(dirs, custom)
	}
	if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
		dirs = append(dirs, filepath.Join(codexHome, "sessions"))
	}

	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		dirs = append(dirs,
			filepath.Join(xdgData, "ai-registry", "codex", "sessions"),
		)
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".codex", "sessions"),
			filepath.Join(home, ".local", "share", "ai-registry", "codex", "sessions"),
		)
	}

	return dirs
}

// FindAgySession locates an Antigravity (`agy`) session transcript on disk given
// an exact file path, conversation directory, conversation ID, or prefix.
func FindAgySession(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", fmt.Errorf("session ID or file path cannot be empty")
	}

	clean := filepath.Clean(trimmed)
	if fi, err := os.Stat(clean); err == nil {
		if !fi.IsDir() {
			return clean, nil
		}
		for _, candidate := range []string{
			filepath.Join(clean, ".system_generated", "logs", "transcript_full.jsonl"),
			filepath.Join(clean, ".system_generated", "logs", "transcript.jsonl"),
			filepath.Join(clean, "transcript_full.jsonl"),
			filepath.Join(clean, "transcript.jsonl"),
		} {
			if cfi, cErr := os.Stat(candidate); cErr == nil && !cfi.IsDir() {
				return candidate, nil
			}
		}
	}

	searchPattern := strings.TrimSuffix(trimmed, ".jsonl")
	for _, dir := range getAgySearchDirs() {
		if match := findAgyTranscriptInDir(dir, searchPattern); match != "" {
			return match, nil
		}
	}

	return "", fmt.Errorf("agy session %q not found as a file or in search directories (~/.gemini/antigravity-cli/brain)", target)
}

func getAgySearchDirs() []string {
	var raw []string
	if custom := os.Getenv("AGY_SESSIONS_DIR"); custom != "" {
		raw = append(raw, custom)
	}
	if geminiHome := os.Getenv("GEMINI_HOME"); geminiHome != "" {
		raw = append(raw, filepath.Join(geminiHome, "antigravity-cli", "brain"))
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		raw = append(raw, filepath.Join(home, ".gemini", "antigravity-cli", "brain"))
	}

	var dirs []string
	seen := make(map[string]struct{}, len(raw))
	for _, dir := range raw {
		cleaned := filepath.Clean(dir)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		dirs = append(dirs, cleaned)
	}
	return dirs
}

func findAgyTranscriptInDir(brainDir string, pattern string) string {
	entries, err := os.ReadDir(brainDir)
	if err != nil {
		return ""
	}

	var prefixMatch, substrMatch string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		convID := entry.Name()
		transcript := selectAgyTranscript(filepath.Join(brainDir, convID))
		if transcript == "" {
			continue
		}
		if convID == pattern {
			return transcript
		}
		if strings.HasPrefix(convID, pattern) && prefixMatch == "" {
			prefixMatch = transcript
		} else if strings.Contains(convID, pattern) && substrMatch == "" {
			substrMatch = transcript
		}
	}
	if prefixMatch != "" {
		return prefixMatch
	}
	return substrMatch
}

func selectAgyTranscript(convDir string) string {
	for _, name := range []string{"transcript_full.jsonl", "transcript.jsonl"} {
		candidate := filepath.Join(convDir, ".system_generated", "logs", name)
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return ""
}

func findTranscriptInDir(rootDir string, pattern string) string {
	if _, err := os.Stat(rootDir); err != nil {
		return ""
	}

	searchPattern := strings.TrimSuffix(pattern, ".jsonl")

	var bestMatch string
	_ = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			return nil
		}

		baseNoExt := strings.TrimSuffix(name, ".jsonl")
		// Exact session ID match (supporting Pi timestamp prefix: <ts>_<id> and Codex rollout prefix: rollout-<ts>-<id>)
		if baseNoExt == searchPattern || strings.HasSuffix(baseNoExt, "_"+searchPattern) || strings.HasSuffix(baseNoExt, "-"+searchPattern) {
			bestMatch = path
			return fs.SkipAll
		}

		// Prefix or delimited match has second priority
		if (strings.HasPrefix(baseNoExt, searchPattern) || strings.Contains(baseNoExt, "_"+searchPattern) || strings.Contains(baseNoExt, "-"+searchPattern)) && bestMatch == "" {
			bestMatch = path
			return nil
		}

		// Substring match
		if strings.Contains(baseNoExt, searchPattern) && bestMatch == "" {
			bestMatch = path
		}

		return nil
	})

	return bestMatch
}
