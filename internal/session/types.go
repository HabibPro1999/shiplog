package session

// Tool identifies which AI coding tool produced a session.
type Tool string

const (
	ToolClaude   Tool = "claude"
	ToolCodex    Tool = "codex"
	ToolAider    Tool = "aider"
	ToolOpenCode Tool = "opencode"
	ToolAmazonQ  Tool = "amazonq"
)

// Display returns the human-readable name for a tool.
func (t Tool) Display() string {
	switch t {
	case ToolClaude:
		return "Claude Code"
	case ToolCodex:
		return "Codex CLI"
	case ToolAider:
		return "Aider"
	case ToolOpenCode:
		return "OpenCode"
	case ToolAmazonQ:
		return "Amazon Q"
	default:
		return string(t)
	}
}

// SessionInfo holds metadata about a single coding session.
type SessionInfo struct {
	Title     string // session title
	SessionID string // unique identifier
	Project   string // readable project path
	Tool      Tool   // which tool produced this session
	FilePath  string // absolute path to session file
	Timestamp string // ISO 8601
}
