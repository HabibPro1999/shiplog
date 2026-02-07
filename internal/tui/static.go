package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/session"
)

// PrintSessionTable renders a styled table of sessions to stdout.
// When showProject is false, the Project column is omitted.
func PrintSessionTable(sessions []session.SessionInfo, showProject bool) {
	const (
		toolW    = 14
		titleW   = 30
		projectW = 16
		dateW    = 12
		idW      = 8
	)

	fmt.Println()

	// Header
	if showProject {
		fmt.Printf("  %s  %s  %s  %s  %s\n",
			HeaderStyle.Width(toolW).Render("Tool"),
			HeaderStyle.Width(titleW).Render("Title"),
			HeaderStyle.Width(projectW).Render("Project"),
			HeaderStyle.Width(dateW).Render("Date"),
			HeaderStyle.Width(idW).Render("ID"),
		)
		fmt.Printf("  %s  %s  %s  %s  %s\n",
			strings.Repeat("\u2500", toolW),
			strings.Repeat("\u2500", titleW),
			strings.Repeat("\u2500", projectW),
			strings.Repeat("\u2500", dateW),
			strings.Repeat("\u2500", idW),
		)
	} else {
		fmt.Printf("  %s  %s  %s  %s\n",
			HeaderStyle.Width(toolW).Render("Tool"),
			HeaderStyle.Width(titleW).Render("Title"),
			HeaderStyle.Width(dateW).Render("Date"),
			HeaderStyle.Width(idW).Render("ID"),
		)
		fmt.Printf("  %s  %s  %s  %s\n",
			strings.Repeat("\u2500", toolW),
			strings.Repeat("\u2500", titleW),
			strings.Repeat("\u2500", dateW),
			strings.Repeat("\u2500", idW),
		)
	}

	// Rows
	for _, s := range sessions {
		tool := ToolColor(s.Tool.Display()).Width(toolW).Render(s.Tool.Display())
		title := truncateWithEllipsis(s.Title, titleW)
		title = TitleStyle.Width(titleW).Render(title)
		date := formatDate(s.Timestamp)
		date = DateStyle.Width(dateW).Render(date)
		id := s.SessionID
		if len(id) > idW {
			id = id[:idW]
		}
		id = IDStyle.Width(idW).Render(id)

		if showProject {
			proj := ShortProject(s.Project)
			if len(proj) > projectW {
				proj = proj[:projectW-3] + "..."
			}
			proj = ProjectStyle.Width(projectW).Render(proj)
			fmt.Printf("  %s  %s  %s  %s  %s\n", tool, title, proj, date, id)
		} else {
			fmt.Printf("  %s  %s  %s  %s\n", tool, title, date, id)
		}
	}

	fmt.Println()
	fmt.Printf("  %s\n\n", CountStyle.Render(fmt.Sprintf("%d sessions", len(sessions))))
}

func truncateWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatDate(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	cleaned := strings.Replace(timestamp, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, cleaned)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, cleaned)
	}
	if err != nil {
		return ""
	}
	return t.Format("Jan 02, 2006")
}
