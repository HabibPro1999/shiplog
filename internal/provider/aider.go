package provider

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// Aider implements the Provider interface for Aider chat history files.
type Aider struct{}

func init() { Register(&Aider{}) }

func (a *Aider) Name() session.Tool { return session.ToolAider }

// Discover looks for .aider.chat.history.md in the project directory.
// In global mode (empty projectFilter), returns empty since Aider has no central session index.
func (a *Aider) Discover(home string, projectFilter string) ([]session.SessionInfo, error) {
	if projectFilter == "" {
		return nil, nil
	}

	histFile := filepath.Join(projectFilter, ".aider.chat.history.md")
	info, err := os.Stat(histFile)
	if err != nil {
		return nil, nil
	}

	h := sha256.Sum256([]byte(histFile))
	sessionID := fmt.Sprintf("%x", h[:8])

	return []session.SessionInfo{
		{
			Title:     "Aider Chat History",
			SessionID: sessionID,
			Project:   filepath.Base(projectFilter),
			Tool:      session.ToolAider,
			FilePath:  histFile,
			Timestamp: info.ModTime().Format(time.RFC3339),
		},
	}, nil
}

var aiderHeadingRe = regexp.MustCompile(`^####\s+(User|Assistant)\s*$`)

// Parse reads an Aider markdown chat history and splits on #### User / #### Assistant headings.
func (a *Aider) Parse(filePath string) ([]parser.Message, parser.SessionMeta, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, parser.SessionMeta{}, err
	}
	defer f.Close()

	var messages []parser.Message
	var currentRole string
	var currentLines []string

	flush := func() {
		if currentRole == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if text != "" {
			role := "user"
			if currentRole == "Assistant" {
				role = "assistant"
			}
			messages = append(messages, parser.Message{
				Role:  role,
				Texts: []string{text},
			})
		}
		currentRole = ""
		currentLines = nil
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if m := aiderHeadingRe.FindStringSubmatch(line); m != nil {
			flush()
			currentRole = m[1]
			continue
		}
		if currentRole != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	// Build metadata
	title := "Aider Chat History"
	if len(messages) > 0 && messages[0].Role == "user" && len(messages[0].Texts) > 0 {
		t := messages[0].Texts[0]
		if len(t) > 60 {
			t = t[:60]
		}
		title = t
	}

	info, _ := os.Stat(filePath)
	var dateRange string
	if info != nil {
		dateRange = info.ModTime().Format("Jan 02, 2006")
	}

	meta := parser.SessionMeta{
		Title:     title,
		DateRange: dateRange,
		Model:     "",
		Tool:      "Aider",
	}

	return messages, meta, nil
}
