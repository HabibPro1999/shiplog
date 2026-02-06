package session

// SessionInfo holds metadata about a single Claude Code session.
type SessionInfo struct {
	Title      string // session title from custom-title entry
	SessionID  string // UUID (filename stem of .jsonl)
	Project    string // readable: "Users/mohamed/projects/foo"
	ProjectDir string // encoded: "-Users-mohamed-projects-foo"
	FilePath   string // absolute path to .jsonl
	Timestamp  string // ISO 8601 from first entry
}
