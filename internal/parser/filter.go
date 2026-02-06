package parser

import (
	"strings"
	"unicode"
)

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

// isSystemContent returns true if the string is system/internal content
// that should be filtered out of the chat display.
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

// isOnlyToolResults returns true if all blocks in the content list are
// tool_result type (no meaningful text or image content).
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

// ToolDisplayName returns a human-readable label for a tool name.
func ToolDisplayName(name string) string {
	if label, ok := toolDisplayNames[name]; ok {
		return label
	}
	// Unknown mcp__ tool: split on __, take last part, replace -/_ with space, title case
	if strings.HasPrefix(name, "mcp__") {
		parts := strings.Split(name, "__")
		if len(parts) >= 3 {
			return titleCase(strings.NewReplacer("-", " ", "_", " ").Replace(parts[len(parts)-1]))
		}
	}
	// Other unknown tools: replace _ with space, title case
	return titleCase(strings.ReplaceAll(name, "_", " "))
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

// asString safely extracts a string from an any value.
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
