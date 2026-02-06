package parser

// Message represents a single chat message in the parsed output.
type Message struct {
	Role      string // "user", "assistant", "tool_group"
	Texts     []string
	Images    []Image
	ToolUses  []string // assistant: tool names used in this message
	ToolNames []string // tool_group: accumulated tool names
	Timestamp string
}

// Image holds a base64-encoded image from a user message.
type Image struct {
	MediaType string
	Data      string // base64
}

// SessionMeta holds extracted metadata about a session.
type SessionMeta struct {
	Title     string
	DateRange string
	Model     string
}
