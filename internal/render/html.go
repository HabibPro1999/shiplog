package render

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/parser"
)

//go:embed template/chat.html
var tmplFS embed.FS

// TemplateMessage is the pre-processed message for the template.
type TemplateMessage struct {
	Role      string
	Texts     []template.HTML // HTML-escaped text (for JS to unescape and render markdown)
	Images    []parser.Image
	ToolLabel string // pre-computed: "Read file" or "5 tool actions performed"
	Timestamp string // pre-formatted
}

// TemplateData holds all data passed to the HTML template.
type TemplateData struct {
	Title          string
	Project        string
	DateRange      string
	Model          string
	UserCount      int
	AssistantCount int
	Messages       []TemplateMessage
}

// Generate renders messages and metadata into a self-contained HTML page.
func Generate(messages []parser.Message, meta parser.SessionMeta, project string) ([]byte, error) {
	funcMap := template.FuncMap{
		"safeHTML": func(s template.HTML) template.HTML { return s },
	}

	tmpl, err := template.New("chat.html").Funcs(funcMap).ParseFS(tmplFS, "template/chat.html")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var tmplMessages []TemplateMessage
	userCount := 0
	assistantCount := 0

	for _, msg := range messages {
		tm := TemplateMessage{
			Role:      msg.Role,
			Timestamp: formatTimestamp(msg.Timestamp),
		}

		switch msg.Role {
		case "user":
			userCount++
			for _, t := range msg.Texts {
				tm.Texts = append(tm.Texts, template.HTML(html.EscapeString(t)))
			}
			tm.Images = msg.Images
		case "assistant":
			assistantCount++
			for _, t := range msg.Texts {
				tm.Texts = append(tm.Texts, template.HTML(html.EscapeString(t)))
			}
		case "tool_group":
			if len(msg.ToolNames) == 1 {
				tm.ToolLabel = parser.ToolDisplayName(msg.ToolNames[0])
			} else if len(msg.ToolNames) > 1 {
				tm.ToolLabel = fmt.Sprintf("%d tool actions performed", len(msg.ToolNames))
			}
		}

		tmplMessages = append(tmplMessages, tm)
	}

	data := TemplateData{
		Title:          meta.Title,
		Project:        project,
		DateRange:      meta.DateRange,
		Model:          meta.Model,
		UserCount:      userCount,
		AssistantCount: assistantCount,
		Messages:       tmplMessages,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// formatTimestamp converts an ISO 8601 timestamp to a display format like "Jan 02, 3:04 PM".
func formatTimestamp(ts string) string {
	if ts == "" {
		return ""
	}
	// Remove trailing Z and try standard formats
	cleaned := strings.Replace(ts, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, cleaned)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.000Z", ts)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05.999999-07:00", cleaned)
			if err != nil {
				return ""
			}
		}
	}
	return t.Format("Jan 02, 3:04 PM")
}
