package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/render"
	"github.com/HabibPro1999/shiplog/internal/session"
	pflag "github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	var (
		showAll   bool
		output    string
		sessionID string
		list      bool
		showVer   bool
	)

	pflag.BoolVarP(&showAll, "all", "a", false, "Show all sessions (ignore project context)")
	pflag.StringVarP(&output, "output", "o", "", "Output HTML file path")
	pflag.StringVar(&sessionID, "session-id", "", "Export by session UUID")
	pflag.BoolVarP(&list, "list", "l", false, "List sessions")
	pflag.BoolVarP(&showVer, "version", "v", false, "Show version")
	pflag.Parse()

	if showVer {
		fmt.Printf("shiplog %s (%s)\n", version, commit)
		os.Exit(0)
	}

	query := pflag.Arg(0)

	// Determine claude projects directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	claudeDir := filepath.Join(home, ".claude", "projects")

	// Determine project scope
	var projectFilter []string
	scopeLabel := "all projects"
	if !showAll {
		matchingDirs := session.ProjectDirsForCWD(claudeDir)
		if len(matchingDirs) > 0 {
			projectFilter = matchingDirs
			cwd, _ := os.Getwd()
			scopeLabel = cwd
		}
	}

	fmt.Printf("  Scanning sessions (%s)...\n", scopeLabel)
	sessions, err := session.FindSessions(claudeDir, projectFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error scanning sessions: %v\n", err)
		os.Exit(1)
	}

	// List mode: -l flag, or no query and no session-id
	if list || (query == "" && sessionID == "") {
		if len(sessions) == 0 && projectFilter != nil {
			fmt.Println("  No sessions found for this project. Use -a to show all.")
		} else {
			listSessions(sessions)
		}
		return
	}

	// Export mode
	q := sessionID
	if q == "" {
		q = query
	}

	match, ambiguous := session.FindByQuery(sessions, q)

	// If scoped search fails, retry with all projects
	if match == nil && len(ambiguous) == 0 && projectFilter != nil {
		fmt.Println("  Not found in current project, searching all...")
		sessions, err = session.FindSessions(claudeDir, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error scanning sessions: %v\n", err)
			os.Exit(1)
		}
		match, ambiguous = session.FindByQuery(sessions, q)
	}

	if match == nil {
		if len(ambiguous) > 0 {
			fmt.Printf("  Multiple sessions match '%s':\n", q)
			for _, m := range ambiguous {
				fmt.Printf("    - %s (%s)\n", m.Title, m.Project)
			}
			fmt.Println("  Be more specific.")
		} else {
			fmt.Printf("  No session found matching '%s'\n", q)
		}
		os.Exit(1)
	}

	fmt.Printf("  Found: \"%s\" (%s)\n", match.Title, match.Project)
	fmt.Println("  Parsing transcript...")

	entries, err := parser.ParseFile(match.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error parsing JSONL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  %d entries\n", len(entries))

	messages := parser.BuildMessages(entries)
	meta := parser.ExtractMeta(entries)

	userCount := 0
	assistantCount := 0
	for _, m := range messages {
		switch m.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}
	fmt.Printf("  %d user messages, %d assistant messages\n", userCount, assistantCount)

	fmt.Println("  Generating HTML...")
	htmlBytes, err := render.Generate(messages, meta, match.Project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error generating HTML: %v\n", err)
		os.Exit(1)
	}

	// Determine output path
	outputPath := output
	if outputPath == "" {
		safeTitle := strings.ReplaceAll(match.Title, " ", "_")
		safeTitle = strings.ReplaceAll(safeTitle, "/", "-")
		outputPath = safeTitle + ".html"
	}

	if err := os.WriteFile(outputPath, htmlBytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "  Error writing file: %v\n", err)
		os.Exit(1)
	}

	sizeMB := float64(len(htmlBytes)) / (1024 * 1024)
	fmt.Printf("  Written to: %s (%.1f MB)\n", outputPath, sizeMB)
	fmt.Println("  Done.")
}

// listSessions prints a formatted table of sessions.
func listSessions(sessions []session.SessionInfo) {
	// Sessions are already sorted by timestamp descending from FindSessions
	fmt.Println()
	fmt.Printf("  %-4s %-25s %-10s %-40s %s\n", "#", "Title", "ID", "Project", "Date")
	fmt.Printf("  %s %s %s %s %s\n",
		strings.Repeat("\u2500", 4),
		strings.Repeat("\u2500", 25),
		strings.Repeat("\u2500", 10),
		strings.Repeat("\u2500", 40),
		strings.Repeat("\u2500", 12),
	)

	for i, s := range sessions {
		ts := ""
		if s.Timestamp != "" {
			cleaned := strings.Replace(s.Timestamp, "Z", "+00:00", 1)
			t, err := time.Parse(time.RFC3339, cleaned)
			if err != nil {
				t, err = time.Parse(time.RFC3339Nano, cleaned)
			}
			if err == nil {
				ts = t.Format("Jan 02, 2006")
			}
		}
		title := truncate(s.Title, 24)
		shortID := s.SessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		proj := truncate(s.Project, 39)
		fmt.Printf("  %-4d %-25s %-10s %-40s %s\n", i+1, title, shortID, proj, ts)
	}

	fmt.Printf("\n  Total: %d sessions\n\n", len(sessions))
}

// truncate returns s truncated to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
