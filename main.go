package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HabibPro1999/shiplog/internal/provider"
	"github.com/HabibPro1999/shiplog/internal/render"
	"github.com/HabibPro1999/shiplog/internal/session"
	"github.com/HabibPro1999/shiplog/internal/tui"
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
		toolFlag  string
		recent    int
	)

	pflag.BoolVarP(&showAll, "all", "a", false, "Show all sessions (ignore project context)")
	pflag.StringVarP(&output, "output", "o", "", "Output HTML file path")
	pflag.StringVar(&sessionID, "session-id", "", "Export by session UUID")
	pflag.BoolVarP(&list, "list", "l", false, "List sessions")
	pflag.BoolVarP(&showVer, "version", "v", false, "Show version")
	pflag.StringVar(&toolFlag, "tool", "", "Filter by tool (claude, codex, aider, opencode, amazonq)")
	pflag.IntVar(&recent, "recent", 0, "Show only sessions from the last N days")
	pflag.Parse()

	if showVer {
		fmt.Printf("shiplog %s (%s)\n", version, commit)
		os.Exit(0)
	}

	query := pflag.Arg(0)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}

	// Parse tool filter
	var toolFilter session.Tool
	if toolFlag != "" {
		toolFilter = session.ParseTool(toolFlag)
		if toolFilter == "" {
			fmt.Fprintf(os.Stderr, "  Unknown tool: %s\n", toolFlag)
			fmt.Fprintf(os.Stderr, "  Available: claude, codex, aider, opencode, amazonq\n")
			os.Exit(1)
		}
	}

	// Build provider list as ProviderDiscoverer interface
	allProviders := provider.All()
	providers := make([]session.ProviderDiscoverer, len(allProviders))
	for i, p := range allProviders {
		providers[i] = p
	}

	// Determine project scope
	projectFilter := ""
	if !showAll {
		cwd, err := os.Getwd()
		if err == nil {
			projectFilter = cwd
		}
	}

	sessions, err := session.FindAllSessions(home, projectFilter, toolFilter, recent, providers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error scanning sessions: %v\n", err)
		os.Exit(1)
	}

	// No sessions: show help
	if len(sessions) == 0 && query == "" && sessionID == "" {
		tui.PrintNoSessionsHelp(projectFilter != "")
		return
	}

	// Interactive/list mode: no query and no session-id
	if query == "" && sessionID == "" {
		if list || !isTerminal() {
			tui.PrintSessionTable(sessions, showAll)
			return
		}
		// Interactive picker
		result, err := tui.RunPicker(sessions, showAll)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		if result.Quit || result.Selected == nil {
			return
		}
		match := result.Selected
		exportSession(match, output)
		return
	}

	// Export mode: query or session-id given
	q := sessionID
	if q == "" {
		q = query
	}

	match, ambiguous := session.FindByQuery(sessions, q)

	// If scoped search fails, retry with all projects
	if match == nil && len(ambiguous) == 0 && projectFilter != "" {
		fmt.Println("  Not found in current project, searching all...")
		sessions, err = session.FindAllSessions(home, "", toolFilter, recent, providers)
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
				fmt.Printf("    - %s (%s, %s)\n", m.Title, m.Project, m.Tool.Display())
			}
			fmt.Println("  Be more specific.")
		} else {
			fmt.Printf("  No session found matching '%s'\n", q)
		}
		os.Exit(1)
	}

	exportSession(match, output)
}

// exportSession parses and exports a session to HTML.
func exportSession(match *session.SessionInfo, output string) {
	fmt.Printf("  Found: \"%s\" (%s, %s)\n", match.Title, match.Project, match.Tool.Display())
	fmt.Println("  Parsing transcript...")

	p := provider.ForTool(match.Tool)
	if p == nil {
		fmt.Fprintf(os.Stderr, "  Error: no provider for tool %s\n", match.Tool)
		os.Exit(1)
	}

	messages, meta, err := p.Parse(match.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error parsing session: %v\n", err)
		os.Exit(1)
	}

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

	outputPath := output
	if outputPath == "" {
		safeTitle := strings.ReplaceAll(match.Title, " ", "_")
		safeTitle = strings.ReplaceAll(safeTitle, "/", "-")
		safeTitle = filepath.Base(safeTitle)
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

// isTerminal reports whether stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
