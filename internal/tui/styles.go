package tui

import (
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})
	ToolStyle     = lipgloss.NewStyle().Bold(true)
	TitleStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})
	DateStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "243"})
	IDStyle       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "243"})
	ProjectStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "33", Dark: "75"})
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "135", Dark: "177"})
	HelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "243"})
	CountStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "243"}).Italic(true)
)

// ToolColor returns a styled lipgloss.Style with the tool's signature color.
func ToolColor(tool string) lipgloss.Style {
	switch tool {
	case "Claude Code":
		return ToolStyle.Foreground(lipgloss.AdaptiveColor{Light: "135", Dark: "177"}) // purple/magenta
	case "Codex CLI":
		return ToolStyle.Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "78"}) // green
	case "Aider":
		return ToolStyle.Foreground(lipgloss.AdaptiveColor{Light: "136", Dark: "221"}) // yellow
	case "OpenCode":
		return ToolStyle.Foreground(lipgloss.AdaptiveColor{Light: "27", Dark: "75"}) // blue
	case "Amazon Q":
		return ToolStyle.Foreground(lipgloss.AdaptiveColor{Light: "166", Dark: "208"}) // orange
	default:
		return ToolStyle
	}
}

// ShortProject returns the base directory name from a full project path.
// Returns an empty string for empty input.
func ShortProject(project string) string {
	if project == "" {
		return ""
	}
	return filepath.Base(project)
}
