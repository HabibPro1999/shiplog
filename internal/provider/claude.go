package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// Claude implements the Provider interface for Claude Code sessions.
type Claude struct{}

func init() { Register(&Claude{}) }

func (c *Claude) Name() session.Tool { return session.ToolClaude }

// Discover scans ~/.claude/projects/ for JSONL session files.
// If projectFilter is non-empty, only matching project directories are scanned.
func (c *Claude) Discover(home string, projectFilter string) ([]session.SessionInfo, error) {
	claudeDir := filepath.Join(home, ".claude", "projects")

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []session.SessionInfo

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		readable := strings.TrimLeft(strings.ReplaceAll(dirName, "-", "/"), "/")

		if projectFilter != "" {
			encoded := pathToProjectDir(projectFilter)
			if dirName != encoded && !strings.HasPrefix(dirName, encoded+"-") {
				continue
			}
		}

		projDir := filepath.Join(claudeDir, dirName)
		jsonlFiles, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err != nil {
			continue
		}

		for _, jf := range jsonlFiles {
			title, firstTS := scanSessionFile(jf)
			sessions = append(sessions, session.SessionInfo{
				Title:     title,
				SessionID: strings.TrimSuffix(filepath.Base(jf), ".jsonl"),
				Project:   readable,
				Tool:      session.ToolClaude,
				FilePath:  jf,
				Timestamp: firstTS,
			})
		}
	}

	return sessions, nil
}

// Parse reads a Claude Code JSONL session file and returns messages and metadata.
func (c *Claude) Parse(filePath string) ([]parser.Message, parser.SessionMeta, error) {
	entries, err := ParseJSONLFile(filePath)
	if err != nil {
		return nil, parser.SessionMeta{}, err
	}
	messages := buildMessages(entries)
	meta := extractMeta(entries)
	meta.Tool = "Claude Code"
	return messages, meta, nil
}

// ParseJSONLFile reads a JSONL file and returns a slice of parsed entries.
// Malformed lines are silently skipped. Exported for use by Codex provider.
func ParseJSONLFile(path string) ([]map[string]any, error) {
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

// ToolDisplayName returns a human-readable label for a tool name.
// Exported because the render package uses it for all providers.
func ToolDisplayName(name string) string {
	if label, ok := toolDisplayNames[name]; ok {
		return label
	}
	if strings.HasPrefix(name, "mcp__") {
		parts := strings.Split(name, "__")
		if len(parts) >= 3 {
			return titleCase(strings.NewReplacer("-", " ", "_", " ").Replace(parts[len(parts)-1]))
		}
	}
	return titleCase(strings.ReplaceAll(name, "_", " "))
}

// --- unexported helpers ---

// pathToProjectDir converts a filesystem path to Claude's project dir name.
func pathToProjectDir(path string) string {
	s := strings.ReplaceAll(path, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// scanSessionFile reads a .jsonl file looking for custom-title, first timestamp,
// and (as a fallback title) the first human message text.
func scanSessionFile(path string) (title, firstTS string) {
	f, err := os.Open(path)
	if err != nil {
		return "(untitled)", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var firstUserText string

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

		// Capture first human message text as fallback title.
		if firstUserText == "" {
			if t, ok := obj["type"].(string); ok && (t == "human" || t == "user") {
				firstUserText = extractFirstUserText(obj)
			}
		}

		if title != "" && firstTS != "" {
			break
		}
	}

	if title == "" {
		if firstUserText != "" {
			title = truncateTitle(firstUserText, 60)
		} else {
			title = "(untitled)"
		}
	}
	return title, firstTS
}

// extractFirstUserText pulls the first text string from a human/user JSONL entry.
func extractFirstUserText(obj map[string]any) string {
	msg, ok := obj["message"].(map[string]any)
	if !ok {
		return ""
	}
	switch c := msg["content"].(type) {
	case string:
		text := strings.TrimSpace(c)
		if text != "" && !isSystemContent(text) {
			return text
		}
	case []any:
		for _, block := range c {
			bm, ok := block.(map[string]any)
			if !ok {
				continue
			}
			if asString(bm["type"]) == "text" {
				text := strings.TrimSpace(asString(bm["text"]))
				if text != "" && !isSystemContent(text) {
					return text
				}
			}
		}
	}
	return ""
}

// truncateTitle shortens s to maxLen characters, appending "..." if truncated.
func truncateTitle(s string, maxLen int) string {
	s = strings.Join(strings.Fields(s), " ") // collapse whitespace
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractUserMessage extracts text and images from a user entry.
func extractUserMessage(entry map[string]any) *parser.Message {
	msg, ok := entry["message"].(map[string]any)
	if !ok {
		return nil
	}
	content := msg["content"]
	timestamp := asString(entry["timestamp"])

	var texts []string
	var images []parser.Image

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
					images = append(images, parser.Image{
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
	return &parser.Message{
		Role:      "user",
		Texts:     texts,
		Images:    images,
		Timestamp: timestamp,
	}
}

// extractAssistantMessage extracts text blocks and tool_use names from an assistant entry.
func extractAssistantMessage(entry map[string]any) *parser.Message {
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
		}
	}

	if len(texts) == 0 && len(toolUses) == 0 {
		return nil
	}
	return &parser.Message{
		Role:      "assistant",
		Texts:     texts,
		ToolUses:  toolUses,
		Timestamp: timestamp,
	}
}

// buildMessages produces a list of Messages from parsed JSONL entries.
// Consecutive tool-only assistant messages are collapsed into tool_group messages.
func buildMessages(entries []map[string]any) []parser.Message {
	var messages []parser.Message
	var pendingTools []string

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

// extractMeta extracts session metadata: title, date range, and model name.
func extractMeta(entries []map[string]any) parser.SessionMeta {
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

	startDate := formatDate(firstTS)
	endDate := formatDate(lastTS)

	var dateRange string
	if startDate != "" && endDate != "" && startDate != endDate {
		dateRange = startDate + " - " + endDate
	} else if startDate != "" {
		dateRange = startDate
	}

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

	return parser.SessionMeta{
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
	ts = strings.Replace(ts, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return ""
		}
	}
	return t.Format("Jan 02, 2006")
}

// asString safely extracts a string from an any value.
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// systemPrefixes are content prefixes that indicate system/internal messages.
var systemPrefixes = []string{
	"<task-notification",
	"<local-command-",
	"<command-name>",
	"<command-message>",
	"[Request interrupted",
	"# Quick Plan",
	"<system-reminder>",
	"This session is being continued from a previous conversation",
}

// isSystemContent returns true if the string is system/internal content.
func isSystemContent(s string) bool {
	trimmed := strings.TrimSpace(s)
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	if strings.Contains(trimmed, "<local-command-") {
		return true
	}
	return false
}

// isOnlyToolResults returns true if all blocks are tool_result type.
func isOnlyToolResults(blocks []any) bool {
	if len(blocks) == 0 {
		return true
	}
	hasNonTool := false
	for _, block := range blocks {
		switch b := block.(type) {
		case map[string]any:
			switch b["type"] {
			case "text":
				text := strings.TrimSpace(asString(b["text"]))
				if text != "" && !isSystemContent(text) {
					hasNonTool = true
				}
			case "image":
				hasNonTool = true
			}
		case string:
			if strings.TrimSpace(b) != "" && !isSystemContent(b) {
				hasNonTool = true
			}
		}
	}
	return !hasNonTool
}

// toolDisplayNames maps internal tool names to human-readable labels.
var toolDisplayNames = map[string]string{
	"Read":                              "Read file",
	"Write":                             "Write file",
	"Edit":                              "Edit file",
	"Bash":                              "Run command",
	"Glob":                              "Search files",
	"Grep":                              "Search content",
	"WebSearch":                         "Web search",
	"WebFetch":                          "Fetch URL",
	"Task":                              "Run agent",
	"Skill":                             "Run skill",
	"TaskCreate":                        "Create task",
	"TaskUpdate":                        "Update task",
	"TaskList":                          "List tasks",
	"TaskGet":                           "Get task",
	"NotebookEdit":                      "Edit notebook",
	"mcp__context7__resolve-library-id": "Lookup docs",
	"mcp__context7__query-docs":         "Query docs",
}

// titleCase capitalises the first letter of each word.
func titleCase(s string) string {
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(rune(prev)) || prev == ' ' {
			prev = r
			return unicode.ToUpper(r)
		}
		prev = r
		return r
	}, s)
}
