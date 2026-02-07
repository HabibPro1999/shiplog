package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// AmazonQ implements Provider for the Amazon Q Developer CLI.
type AmazonQ struct{}

func init() { Register(&AmazonQ{}) }

func (a *AmazonQ) Name() session.Tool { return session.ToolAmazonQ }

// amazonQFile represents the top-level structure of a chat-history JSON file.
type amazonQFile struct {
	// Direct fields (raw format)
	WorkspaceFolder string           `json:"workspaceFolder"`
	Messages        []amazonQMessage `json:"messages"`

	// LokiJS wrapper fields
	Collections []amazonQCollection `json:"collections"`
}

// amazonQCollection represents a LokiJS collection entry.
type amazonQCollection struct {
	Name string            `json:"name"`
	Data []json.RawMessage `json:"data"`
}

// amazonQMessage represents a single message in the Amazon Q chat history.
type amazonQMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Body      string `json:"body"`
	Timestamp string `json:"timestamp"`
	CreatedAt string `json:"createdAt"`
}

func (a *AmazonQ) Discover(home string, projectFilter string) ([]session.SessionInfo, error) {
	histDir := filepath.Join(home, ".aws", "amazonq", "history")
	pattern := filepath.Join(histDir, "chat-history-*.json")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil
	}

	var sessions []session.SessionInfo

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var qf amazonQFile
		if err := json.Unmarshal(data, &qf); err != nil {
			continue
		}

		workspace := qf.WorkspaceFolder
		msgs := qf.Messages

		// Handle LokiJS wrapper format
		if len(msgs) == 0 && len(qf.Collections) > 0 {
			for _, coll := range qf.Collections {
				if coll.Name == "messages" || coll.Name == "chatMessages" {
					for _, raw := range coll.Data {
						var m amazonQMessage
						if err := json.Unmarshal(raw, &m); err == nil {
							msgs = append(msgs, m)
						}
					}
				}
				if coll.Name == "metadata" || coll.Name == "sessions" {
					for _, raw := range coll.Data {
						var meta map[string]any
						if err := json.Unmarshal(raw, &meta); err == nil {
							if ws, ok := meta["workspaceFolder"].(string); ok && ws != "" {
								workspace = ws
							}
						}
					}
				}
			}
		}

		// Apply project filter
		if projectFilter != "" && workspace != "" {
			cleanWorkspace := filepath.Clean(workspace)
			cleanFilter := filepath.Clean(projectFilter)
			if cleanWorkspace != cleanFilter && !strings.HasPrefix(cleanWorkspace, cleanFilter+string(filepath.Separator)) {
				continue
			}
		}

		// Get timestamp from first message
		ts := ""
		for _, m := range msgs {
			if m.Timestamp != "" {
				ts = m.Timestamp
				break
			}
			if m.CreatedAt != "" {
				ts = m.CreatedAt
				break
			}
		}

		// Fallback to file modification time
		if ts == "" {
			if info, err := os.Stat(f); err == nil {
				ts = info.ModTime().UTC().Format(time.RFC3339)
			}
		}

		// Build title from first user message
		title := "(untitled)"
		for _, m := range msgs {
			if m.Role == "user" {
				text := m.Content
				if text == "" {
					text = m.Body
				}
				text = strings.TrimSpace(text)
				if text != "" {
					if len(text) > 60 {
						text = text[:60]
					}
					title = text
					break
				}
			}
		}

		displayProject := workspace
		if displayProject == "" {
			displayProject = "(unknown project)"
		}

		sessionID := strings.TrimSuffix(filepath.Base(f), ".json")
		sessionID = strings.TrimPrefix(sessionID, "chat-history-")

		sessions = append(sessions, session.SessionInfo{
			Title:     title,
			SessionID: sessionID,
			Project:   displayProject,
			Tool:      session.ToolAmazonQ,
			FilePath:  f,
			Timestamp: ts,
		})
	}

	return sessions, nil
}

func (a *AmazonQ) Parse(filePath string) ([]parser.Message, parser.SessionMeta, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, parser.SessionMeta{}, err
	}

	var qf amazonQFile
	if err := json.Unmarshal(data, &qf); err != nil {
		return nil, parser.SessionMeta{}, err
	}

	rawMsgs := qf.Messages

	// Handle LokiJS wrapper format
	if len(rawMsgs) == 0 && len(qf.Collections) > 0 {
		for _, coll := range qf.Collections {
			if coll.Name == "messages" || coll.Name == "chatMessages" {
				for _, raw := range coll.Data {
					var m amazonQMessage
					if err := json.Unmarshal(raw, &m); err == nil {
						rawMsgs = append(rawMsgs, m)
					}
				}
			}
		}
	}

	var messages []parser.Message
	var firstTS, lastTS string

	for _, m := range rawMsgs {
		text := m.Content
		if text == "" {
			text = m.Body
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		role := m.Role
		if role != "user" && role != "assistant" {
			continue
		}

		ts := m.Timestamp
		if ts == "" {
			ts = m.CreatedAt
		}
		if ts != "" {
			if firstTS == "" {
				firstTS = ts
			}
			lastTS = ts
		}

		messages = append(messages, parser.Message{
			Role:      role,
			Texts:     []string{text},
			Timestamp: ts,
		})
	}

	// Build title from first user message
	title := "Untitled Session"
	for _, m := range messages {
		if m.Role == "user" && len(m.Texts) > 0 {
			t := m.Texts[0]
			if len(t) > 60 {
				t = t[:60]
			}
			title = t
			break
		}
	}

	// Build date range
	startDate := formatAmazonQDate(firstTS)
	endDate := formatAmazonQDate(lastTS)
	var dateRange string
	if startDate != "" && endDate != "" && startDate != endDate {
		dateRange = startDate + " - " + endDate
	} else if startDate != "" {
		dateRange = startDate
	}

	// Fallback to file mod time
	if dateRange == "" {
		if info, err := os.Stat(filePath); err == nil {
			dateRange = info.ModTime().Format("Jan 02, 2006")
		}
	}

	meta := parser.SessionMeta{
		Title:     title,
		DateRange: dateRange,
		Model:     "",
		Tool:      "Amazon Q",
	}

	return messages, meta, nil
}

// formatAmazonQDate parses various timestamp formats and returns "Mon DD, YYYY".
func formatAmazonQDate(ts string) string {
	if ts == "" {
		return ""
	}
	cleaned := strings.Replace(ts, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, cleaned)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, cleaned)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05.000Z", ts)
			if err != nil {
				return ""
			}
		}
	}
	return t.Format("Jan 02, 2006")
}
