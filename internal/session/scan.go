package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PathToProjectDir converts a filesystem path to Claude's project dir name.
// Claude replaces both / and _ with - in directory names.
// e.g. /Users/mohamed/projects/value_slim_demo -> -Users-mohamed-projects-value-slim-demo
func PathToProjectDir(path string) string {
	s := strings.ReplaceAll(path, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// ProjectDirsForCWD returns project dir names that match the current working directory.
// Matches the CWD itself and any subdirectory projects.
func ProjectDirsForCWD(claudeDir string) []string {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	cwdEncoded := PathToProjectDir(cwd)

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == cwdEncoded || strings.HasPrefix(name, cwdEncoded+"-") {
			matches = append(matches, name)
		}
	}
	return matches
}

// FindSessions scans project dirs and returns a list of SessionInfo.
// If projectFilter is non-nil, only those dirs are scanned. Otherwise all dirs are scanned.
// Results are sorted by timestamp descending (most recent first).
func FindSessions(claudeDir string, projectFilter []string) ([]SessionInfo, error) {
	var projDirs []string

	if projectFilter != nil {
		for _, d := range projectFilter {
			full := filepath.Join(claudeDir, d)
			if info, err := os.Stat(full); err == nil && info.IsDir() {
				projDirs = append(projDirs, full)
			}
		}
	} else {
		entries, err := os.ReadDir(claudeDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				projDirs = append(projDirs, filepath.Join(claudeDir, e.Name()))
			}
		}
	}

	var sessions []SessionInfo

	for _, projDir := range projDirs {
		jsonlFiles, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err != nil {
			continue
		}

		projName := filepath.Base(projDir)
		// Convert project dir name back to readable path: replace - with /, strip leading /
		readable := strings.TrimLeft(strings.ReplaceAll(projName, "-", "/"), "/")

		for _, jf := range jsonlFiles {
			title, firstTS := scanSessionFile(jf)

			sessions = append(sessions, SessionInfo{
				Title:      title,
				SessionID:  strings.TrimSuffix(filepath.Base(jf), ".jsonl"),
				Project:    readable,
				ProjectDir: projName,
				FilePath:   jf,
				Timestamp:  firstTS,
			})
		}
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp > sessions[j].Timestamp
	})

	return sessions, nil
}

// scanSessionFile reads a .jsonl file line-by-line looking for custom-title and first timestamp.
func scanSessionFile(path string) (title, firstTS string) {
	f, err := os.Open(path)
	if err != nil {
		return "(untitled)", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Allow large lines (some JSONL entries can be big with base64 images)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		var obj map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			continue
		}

		if t, ok := obj["type"].(string); ok && t == "custom-title" {
			if ct, ok := obj["customTitle"].(string); ok {
				title = ct
			}
		}

		if firstTS == "" {
			if ts, ok := obj["timestamp"].(string); ok && ts != "" {
				firstTS = ts
			} else if snap, ok := obj["snapshot"].(map[string]any); ok {
				if ts, ok := snap["timestamp"].(string); ok && ts != "" {
					firstTS = ts
				}
			}
		}

		if title != "" && firstTS != "" {
			break
		}
	}

	if title == "" {
		title = "(untitled)"
	}
	return title, firstTS
}
