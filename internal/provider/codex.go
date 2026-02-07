package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// Codex implements the Provider interface for OpenAI Codex CLI sessions.
type Codex struct{}

func init() { Register(&Codex{}) }

func (c *Codex) Name() session.Tool { return session.ToolCodex }

// Discover scans ~/.codex/sessions/ recursively for rollout-*.jsonl files.
func (c *Codex) Discover(home string, projectFilter string) ([]session.SessionInfo, error) {
	sessionsDir := filepath.Join(home, ".codex", "sessions")

	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var sessions []session.SessionInfo

	err := filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if !strings.HasPrefix(base, "rollout-") || !strings.HasSuffix(base, ".jsonl") {
			return nil
		}

		// Read line-by-line and stop after finding session_meta
		cwd, title, timestamp := scanCodexSessionFile(path)

		// Fall back to file mod time for timestamp
		if timestamp == "" {
			timestamp = info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
		}

		// Project filtering: compare cwd against projectFilter
		if projectFilter != "" && cwd != "" {
			if !strings.HasPrefix(cwd, projectFilter) && !strings.HasPrefix(projectFilter, cwd) {
				return nil
			}
		}

		project := cwd
		if project == "" {
			project = "(unknown)"
		}

		if title == "" {
			title = "(untitled)"
		}

		sessionID := strings.TrimSuffix(base, ".jsonl")

		sessions = append(sessions, session.SessionInfo{
			Title:     title,
			SessionID: sessionID,
			Project:   project,
			Tool:      session.ToolCodex,
			FilePath:  path,
			Timestamp: timestamp,
		})

		return nil
	})

	if err != nil {
		return sessions, err
	}

	return sessions, nil
}

// Parse reads a Codex CLI JSONL session file and returns messages and metadata.
func (c *Codex) Parse(filePath string) ([]parser.Message, parser.SessionMeta, error) {
	entries, err := ParseJSONLFile(filePath)
	if err != nil {
		return nil, parser.SessionMeta{}, err
	}

	var messages []parser.Message
	var pendingTools []string
	var meta parser.SessionMeta
	meta.Tool = "Codex CLI"

	var firstTS, lastTS string

	flushTools := func() {
		if len(pendingTools) == 0 {
			return
		}
		messages = append(messages, parser.Message{
			Role:      "tool_group",
			ToolNames: append([]string(nil), pendingTools...),
		})
		pendingTools = pendingTools[:0]
	}

	for _, entry := range entries {
		entryType := asString(entry["type"])
		ts := asString(entry["timestamp"])
		if ts != "" {
			if firstTS == "" {
				firstTS = ts
			}
			lastTS = ts
		}

		switch entryType {
		case "session_meta":
			if payload, ok := entry["payload"].(map[string]any); ok {
				meta.Title = asString(payload["title"])
				if model := asString(payload["model"]); model != "" {
					meta.Model = model
				}
			}

		case "event_msg":
			payload, ok := entry["payload"].(map[string]any)
			if !ok {
				continue
			}
			payloadType := asString(payload["type"])
			if payloadType == "user_message" {
				text := asString(payload["text"])
				if text == "" {
					// Try content field
					text = asString(payload["content"])
				}
				if text != "" {
					flushTools()
					messages = append(messages, parser.Message{
						Role:      "user",
						Texts:     []string{text},
						Timestamp: ts,
					})
				}
			}

		case "response_item":
			role := asString(entry["role"])
			if role != "assistant" {
				// Check inside item
				if item, ok := entry["item"].(map[string]any); ok {
					role = asString(item["role"])
				}
			}
			if role != "assistant" {
				continue
			}

			var texts []string
			var toolUses []string

			// Content can be at entry["content"] or entry["item"]["content"]
			content := getContentArray(entry)

			for _, block := range content {
				bm, ok := block.(map[string]any)
				if !ok {
					continue
				}
				blockType := asString(bm["type"])
				switch blockType {
				case "text", "output_text":
					text := strings.TrimSpace(asString(bm["text"]))
					if text != "" {
						texts = append(texts, text)
					}
				case "function_call", "tool_use":
					name := asString(bm["name"])
					if name == "" {
						name = asString(bm["function"])
					}
					if name == "" {
						name = "tool"
					}
					toolUses = append(toolUses, name)
				}
			}

			if len(texts) > 0 {
				flushTools()
				messages = append(messages, parser.Message{
					Role:      "assistant",
					Texts:     texts,
					ToolUses:  toolUses,
					Timestamp: ts,
				})
			} else if len(toolUses) > 0 {
				pendingTools = append(pendingTools, toolUses...)
			}
		}
	}

	flushTools()

	// Build date range
	startDate := formatDate(firstTS)
	endDate := formatDate(lastTS)
	if startDate != "" && endDate != "" && startDate != endDate {
		meta.DateRange = startDate + " - " + endDate
	} else if startDate != "" {
		meta.DateRange = startDate
	}

	if meta.Title == "" {
		meta.Title = "Untitled Session"
	}

	return messages, meta, nil
}

// scanCodexSessionFile reads a Codex JSONL file line-by-line and returns
// the cwd, title, and timestamp from the first session_meta entry.
// It stops reading as soon as session_meta is found.
func scanCodexSessionFile(path string) (cwd, title, timestamp string) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if asString(entry["type"]) == "session_meta" {
			if payload, ok := entry["payload"].(map[string]any); ok {
				cwd = asString(payload["cwd"])
				title = asString(payload["title"])
			}
			timestamp = asString(entry["timestamp"])
			return cwd, title, timestamp
		}
	}

	return "", "", ""
}

// getContentArray extracts the content array from a Codex response_item entry.
// It checks entry["content"], entry["item"]["content"], and entry["output"] paths.
func getContentArray(entry map[string]any) []any {
	if content, ok := entry["content"].([]any); ok {
		return content
	}
	if item, ok := entry["item"].(map[string]any); ok {
		if content, ok := item["content"].([]any); ok {
			return content
		}
		if output, ok := item["output"].([]any); ok {
			return output
		}
	}
	if output, ok := entry["output"].([]any); ok {
		return output
	}
	return nil
}
