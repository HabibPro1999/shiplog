package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// OpenCode implements Provider for the OpenCode AI coding tool.
type OpenCode struct{}

func init() { Register(&OpenCode{}) }

func (o *OpenCode) Name() session.Tool { return session.ToolOpenCode }

// openCodeStorageDirs returns possible storage root directories for OpenCode.
func openCodeStorageDirs(home string) []string {
	dirs := []string{
		filepath.Join(home, ".local", "share", "opencode", "storage"),
	}
	if runtime.GOOS == "darwin" {
		dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "opencode", "storage"))
	}
	return dirs
}

// openCodeProject holds minimal project metadata from OpenCode's project JSON.
type openCodeProject struct {
	Path string `json:"path"`
}

// openCodeSessionFile holds minimal session metadata from OpenCode's session JSON.
type openCodeSessionFile struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ProjectID string `json:"projectID"`
	CreatedAt int64  `json:"createdAt"` // unix millis or seconds
	UpdatedAt int64  `json:"updatedAt"`
}

// openCodeMessage holds a single message from OpenCode's message JSON.
type openCodeMessage struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionID"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"createdAt"`
}

// openCodePart holds a content part from OpenCode's part JSON.
type openCodePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (o *OpenCode) Discover(home string, projectFilter string) ([]session.SessionInfo, error) {
	var sessions []session.SessionInfo

	for _, storageDir := range openCodeStorageDirs(home) {
		sessionDir := filepath.Join(storageDir, "session")
		projectDir := filepath.Join(storageDir, "project")

		// Build project ID → path map
		projectPaths := make(map[string]string)
		projFiles, err := filepath.Glob(filepath.Join(projectDir, "*.json"))
		if err == nil {
			for _, pf := range projFiles {
				data, err := os.ReadFile(pf)
				if err != nil {
					continue
				}
				var proj openCodeProject
				if err := json.Unmarshal(data, &proj); err != nil {
					continue
				}
				projID := strings.TrimSuffix(filepath.Base(pf), ".json")
				projectPaths[projID] = proj.Path
			}
		}

		// Scan session directories (each subdirectory is a projectID)
		projEntries, err := os.ReadDir(sessionDir)
		if err != nil {
			continue
		}

		for _, pe := range projEntries {
			if !pe.IsDir() {
				continue
			}
			projID := pe.Name()
			projectPath := projectPaths[projID]

			// Apply project filter: skip sessions with unknown paths, require exact or subdirectory match
			if projectFilter != "" {
				if projectPath == "" || (projectPath != projectFilter && !strings.HasPrefix(projectPath, projectFilter+string(filepath.Separator))) {
					continue
				}
			}

			// Scan session JSON files within this project directory
			sessFiles, err := filepath.Glob(filepath.Join(sessionDir, projID, "*.json"))
			if err != nil {
				continue
			}

			for _, sf := range sessFiles {
				data, err := os.ReadFile(sf)
				if err != nil {
					continue
				}
				var sess openCodeSessionFile
				if err := json.Unmarshal(data, &sess); err != nil {
					continue
				}

				ts := formatUnixTimestamp(sess.UpdatedAt)
				if ts == "" {
					ts = formatUnixTimestamp(sess.CreatedAt)
				}

				title := sess.Title
				if title == "" {
					title = fallbackOpenCodeTitle(storageDir, sess.ID)
				}
				if title == "" {
					title = "(untitled)"
				}

				displayProject := projectPath
				if displayProject == "" {
					displayProject = projID
				}

				sessions = append(sessions, session.SessionInfo{
					Title:     title,
					SessionID: sess.ID,
					Project:   displayProject,
					Tool:      session.ToolOpenCode,
					FilePath:  sf,
					Timestamp: ts,
				})
			}
		}
	}

	return sessions, nil
}

func (o *OpenCode) Parse(filePath string) ([]parser.Message, parser.SessionMeta, error) {
	// Read session file for metadata
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, parser.SessionMeta{}, err
	}
	var sess openCodeSessionFile
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, parser.SessionMeta{}, err
	}

	// Validate session ID to prevent path traversal
	if strings.Contains(sess.ID, "..") || strings.ContainsAny(sess.ID, `/\`) {
		return nil, parser.SessionMeta{}, fmt.Errorf("invalid session ID: %q", sess.ID)
	}

	// Determine storage root from the session file path.
	// Path is: {storageDir}/session/{projectID}/{sessionID}.json
	storageDir := filepath.Dir(filepath.Dir(filepath.Dir(filePath)))
	messageDir := filepath.Join(storageDir, "message", sess.ID)
	partDir := filepath.Join(storageDir, "part")

	// Read messages for this session
	msgFiles, _ := filepath.Glob(filepath.Join(messageDir, "*.json"))

	type msgEntry struct {
		msg       openCodeMessage
		filePath  string
		createdAt int64
	}
	var msgEntries []msgEntry
	for _, mf := range msgFiles {
		mdata, err := os.ReadFile(mf)
		if err != nil {
			continue
		}
		var m openCodeMessage
		if err := json.Unmarshal(mdata, &m); err != nil {
			continue
		}
		msgEntries = append(msgEntries, msgEntry{msg: m, filePath: mf, createdAt: m.CreatedAt})
	}

	// Sort messages by creation time
	sort.Slice(msgEntries, func(i, j int) bool {
		return msgEntries[i].createdAt < msgEntries[j].createdAt
	})

	var messages []parser.Message
	for _, me := range msgEntries {
		// Validate message ID to prevent path traversal
		if strings.Contains(me.msg.ID, "..") || strings.ContainsAny(me.msg.ID, `/\`) {
			continue
		}
		// Read parts for this message
		partFiles, _ := filepath.Glob(filepath.Join(partDir, me.msg.ID, "*.json"))

		var texts []string
		for _, pf := range partFiles {
			pdata, err := os.ReadFile(pf)
			if err != nil {
				continue
			}
			var part openCodePart
			if err := json.Unmarshal(pdata, &part); err != nil {
				continue
			}
			if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
				texts = append(texts, part.Text)
			}
		}

		if len(texts) == 0 {
			continue
		}

		role := me.msg.Role
		if role != "user" && role != "assistant" {
			continue
		}

		messages = append(messages, parser.Message{
			Role:      role,
			Texts:     texts,
			Timestamp: formatUnixTimestamp(me.msg.CreatedAt),
		})
	}

	// Build metadata
	title := sess.Title
	if title == "" {
		title = "Untitled Session"
	}

	startDate := formatDateFromUnix(sess.CreatedAt)
	endDate := formatDateFromUnix(sess.UpdatedAt)
	var dateRange string
	if startDate != "" && endDate != "" && startDate != endDate {
		dateRange = startDate + " - " + endDate
	} else if startDate != "" {
		dateRange = startDate
	}

	meta := parser.SessionMeta{
		Title:     title,
		DateRange: dateRange,
		Model:     "",
		Tool:      "OpenCode",
	}

	return messages, meta, nil
}

// fallbackOpenCodeTitle reads the earliest user message for a session and returns
// a truncated version of its text, for use when the session has no title.
func fallbackOpenCodeTitle(storageDir string, sessionID string) string {
	if strings.Contains(sessionID, "..") || strings.ContainsAny(sessionID, `/\`) {
		return ""
	}
	messageDir := filepath.Join(storageDir, "message", sessionID)
	msgFiles, _ := filepath.Glob(filepath.Join(messageDir, "*.json"))

	// Find earliest user message by createdAt
	var bestText string
	var bestTime int64 = 1<<63 - 1

	for _, mf := range msgFiles {
		data, err := os.ReadFile(mf)
		if err != nil {
			continue
		}
		var m openCodeMessage
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.Role != "user" || m.CreatedAt >= bestTime {
			continue
		}

		// Read parts to get text
		partDir := filepath.Join(storageDir, "part")
		if strings.Contains(m.ID, "..") || strings.ContainsAny(m.ID, `/\`) {
			continue
		}
		partFiles, _ := filepath.Glob(filepath.Join(partDir, m.ID, "*.json"))
		for _, pf := range partFiles {
			pdata, err := os.ReadFile(pf)
			if err != nil {
				continue
			}
			var part openCodePart
			if err := json.Unmarshal(pdata, &part); err != nil {
				continue
			}
			text := strings.TrimSpace(part.Text)
			if part.Type == "text" && text != "" {
				bestText = text
				bestTime = m.CreatedAt
				break
			}
		}
	}

	if bestText != "" {
		return truncateTitle(bestText, 60)
	}
	return ""
}

// formatUnixTimestamp converts a unix timestamp (millis or seconds) to ISO 8601.
func formatUnixTimestamp(ts int64) string {
	if ts == 0 {
		return ""
	}
	// If the timestamp is too large to be seconds, treat as milliseconds
	if ts > 1e12 {
		return time.UnixMilli(ts).UTC().Format(time.RFC3339)
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

// formatDateFromUnix converts a unix timestamp to "Mon DD, YYYY" format.
func formatDateFromUnix(ts int64) string {
	if ts == 0 {
		return ""
	}
	var t time.Time
	if ts > 1e12 {
		t = time.UnixMilli(ts)
	} else {
		t = time.Unix(ts, 0)
	}
	return t.Format("Jan 02, 2006")
}
