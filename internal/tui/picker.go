package tui

import (
	"fmt"

	"github.com/HabibPro1999/shiplog/internal/session"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PickerResult holds the outcome of the interactive session picker.
type PickerResult struct {
	Selected *session.SessionInfo
	Quit     bool
}

// sessionItem wraps a SessionInfo to implement list.DefaultItem.
type sessionItem struct {
	info        session.SessionInfo
	showProject bool
}

func (s sessionItem) Title() string {
	t := s.info.Title
	if t == "" {
		t = "(untitled)"
	}
	return t
}

func (s sessionItem) Description() string {
	tool := s.info.Tool.Display()
	date := s.info.Timestamp
	if len(date) >= 10 {
		date = date[:10]
	}
	desc := tool + "  " + date
	if s.showProject && s.info.Project != "" {
		desc += "  " + ShortProject(s.info.Project)
	}
	return desc
}

func (s sessionItem) FilterValue() string {
	return s.info.Title + " " + s.info.Tool.Display() + " " + s.info.Project
}

// pickerModel is the bubbletea model for the session picker.
type pickerModel struct {
	list     list.Model
	selected *session.SessionInfo
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept keys when filtering is active
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(sessionItem); ok {
				info := i.info
				m.selected = &info
			}
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	return m.list.View()
}

// RunPicker launches the interactive session picker and returns the user's selection.
func RunPicker(sessions []session.SessionInfo, showProject bool) (PickerResult, error) {
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = sessionItem{info: s, showProject: showProject}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "243"})

	l := list.New(items, delegate, 0, 0)
	l.Title = "shiplog — select a session"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	m := pickerModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return PickerResult{}, fmt.Errorf("picker error: %w", err)
	}

	final := result.(pickerModel)
	if final.selected == nil {
		return PickerResult{Quit: true}, nil
	}
	return PickerResult{Selected: final.selected}, nil
}
