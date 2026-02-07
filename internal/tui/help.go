package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// PrintNoSessionsHelp prints contextual help when no sessions are found.
// When scoped is true, it suggests using -a. When false, it shows onboarding.
func PrintNoSessionsHelp(scoped bool) {
	fmt.Println()

	if scoped {
		fmt.Printf("  %s\n\n", HelpStyle.Render("No sessions found for this project. Use")+
			" "+boldText.Render("shiplog -a")+
			" "+HelpStyle.Render("to show all projects."))
		return
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "135", Dark: "177"})
	fmt.Printf("  %s\n\n", title.Render("shiplog")+HelpStyle.Render(" \u2014 export AI coding sessions to HTML"))

	fmt.Printf("  %s %s\n\n",
		HelpStyle.Render("Supported tools:"),
		lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render(
			"Claude Code, Codex CLI, Aider, OpenCode, Amazon Q"),
	)

	fmt.Printf("  %s\n", HelpStyle.Render("Usage:"))
	fmt.Printf("    %s  %s\n", boldText.Render("shiplog"), HelpStyle.Render("Interactive session picker"))
	fmt.Printf("    %s  %s\n", boldText.Render("shiplog -a"), HelpStyle.Render("Show sessions from all projects"))
	fmt.Printf("    %s  %s\n", boldText.Render("shiplog --tool codex"), HelpStyle.Render("Filter by tool"))
	fmt.Printf("    %s  %s\n", boldText.Render("shiplog -l"), HelpStyle.Render("Static list (for scripting)"))
	fmt.Println()
}

var boldText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})
