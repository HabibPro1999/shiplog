package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// ParseFile reads a JSONL file and returns a slice of parsed entries.
// Malformed lines are silently skipped.
func ParseFile(path string) ([]map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []map[string]any
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10 MB max line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}

// extractUserMessage extracts text and images from a user entry.
// Returns nil if the message should be skipped (system content, only tool results, empty).
func extractUserMessage(entry map[string]any) *Message {
	msg, ok := entry["message"].(map[string]any)
	if !ok {
		return nil
	}
	content := msg["content"]
	timestamp := asString(entry["timestamp"])

	var texts []string
	var images []Image

	switch c := content.(type) {
	case string:
		if strings.TrimSpace(c) == "" || isSystemContent(c) {
			return nil
		}
		texts = append(texts, c)

	case []any:
		if isOnlyToolResults(c) {
			return nil
		}
		for _, block := range c {
			bm, ok := block.(map[string]any)
			if !ok {
				continue
			}
			switch asString(bm["type"]) {
			case "text":
				text := strings.TrimSpace(asString(bm["text"]))
				if text != "" && !isSystemContent(text) {
					texts = append(texts, text)
				}
			case "image":
				source, ok := bm["source"].(map[string]any)
				if !ok {
					continue
				}
				if asString(source["type"]) == "base64" {
					mediaType := asString(source["media_type"])
					if mediaType == "" {
						mediaType = "image/png"
					}
					images = append(images, Image{
						MediaType: mediaType,
						Data:      asString(source["data"]),
					})
				}
			}
		}

	default:
		return nil
	}

	if len(texts) == 0 && len(images) == 0 {
		return nil
	}
	return &Message{
		Role:      "user",
		Texts:     texts,
		Images:    images,
		Timestamp: timestamp,
	}
}

// extractAssistantMessage extracts text blocks and tool_use names from an assistant entry.
// Thinking blocks are skipped entirely.
// Returns nil if nothing meaningful was found.
func extractAssistantMessage(entry map[string]any) *Message {
	msg, ok := entry["message"].(map[string]any)
	if !ok {
		return nil
	}
	content, ok := msg["content"].([]any)
	if !ok {
		return nil
	}
	timestamp := asString(entry["timestamp"])

	var texts []string
	var toolUses []string

	for _, block := range content {
		bm, ok := block.(map[string]any)
		if !ok {
			continue
		}
		blockType := asString(bm["type"])
		switch blockType {
		case "text":
			text := strings.TrimSpace(asString(bm["text"]))
			if text != "" {
				texts = append(texts, text)
			}
		case "tool_use":
			name := asString(bm["name"])
			if name == "" {
				name = "unknown"
			}
			toolUses = append(toolUses, name)
			// "thinking" blocks are intentionally skipped
		}
	}

	if len(texts) == 0 && len(toolUses) == 0 {
		return nil
	}
	return &Message{
		Role:      "assistant",
		Texts:     texts,
		ToolUses:  toolUses,
		Timestamp: timestamp,
	}
}

// BuildMessages iterates through parsed JSONL entries and produces a list of
// Messages. Consecutive tool-only assistant messages are collapsed into a
// single tool_group message. A tool_group is flushed whenever a user message
// or an assistant message with text appears.
func BuildMessages(entries []map[string]any) []Message {
	var messages []Message
	var pendingTools []string

	flushTools := func() {
		if len(pendingTools) == 0 {
			return
		}
		messages = append(messages, Message{
			Role:      "tool_group",
			ToolNames: append([]string(nil), pendingTools...),
		})
		pendingTools = pendingTools[:0]
	}

	for _, entry := range entries {
		entryType := asString(entry["type"])

		switch entryType {
		case "user":
			result := extractUserMessage(entry)
			if result == nil {
				continue
			}
			flushTools()
			messages = append(messages, *result)

		case "assistant":
			result := extractAssistantMessage(entry)
			if result == nil {
				continue
			}
			if len(result.Texts) > 0 {
				flushTools()
				messages = append(messages, *result)
			} else if len(result.ToolUses) > 0 {
				pendingTools = append(pendingTools, result.ToolUses...)
			}
		}
	}

	flushTools()
	return messages
}

// ExtractMeta extracts session metadata: title, date range, and model name.
func ExtractMeta(entries []map[string]any) SessionMeta {
	var title string
	var firstTS, lastTS string
	var model string

	for _, entry := range entries {
		if asString(entry["type"]) == "custom-title" {
			title = asString(entry["customTitle"])
		}

		ts := asString(entry["timestamp"])
		if ts == "" {
			if snap, ok := entry["snapshot"].(map[string]any); ok {
				ts = asString(snap["timestamp"])
			}
		}
		if ts != "" {
			if firstTS == "" {
				firstTS = ts
			}
			lastTS = ts
		}

		if model == "" && asString(entry["type"]) == "assistant" {
			if msg, ok := entry["message"].(map[string]any); ok {
				model = asString(msg["model"])
			}
		}
	}

	// Parse dates
	startDate := formatDate(firstTS)
	endDate := formatDate(lastTS)

	var dateRange string
	if startDate != "" && endDate != "" && startDate != endDate {
		dateRange = startDate + " - " + endDate
	} else if startDate != "" {
		dateRange = startDate
	}

	// Clean model name
	modelDisplay := model
	if modelDisplay == "" {
		modelDisplay = "Claude"
	}
	switch {
	case strings.Contains(modelDisplay, "opus"):
		modelDisplay = "Claude Opus"
	case strings.Contains(modelDisplay, "sonnet"):
		modelDisplay = "Claude Sonnet"
	case strings.Contains(modelDisplay, "haiku"):
		modelDisplay = "Claude Haiku"
	}

	if title == "" {
		title = "Untitled Session"
	}

	return SessionMeta{
		Title:     title,
		DateRange: dateRange,
		Model:     modelDisplay,
	}
}

// formatDate parses an ISO 8601 timestamp and returns "Mon DD, YYYY".
func formatDate(ts string) string {
	if ts == "" {
		return ""
	}
	// Replace Z with +00:00 for consistent parsing
	ts = strings.Replace(ts, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try RFC3339Nano
		t, err = time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return ""
		}
	}
	return t.Format("Jan 02, 2006")
}
