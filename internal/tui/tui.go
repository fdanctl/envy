package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fdanctl/envy/internal/store"
)

// type presets struct {
// 	name string
// 	kv   map[string]string
// }

type model struct {
	choices []string
	cursor  int
}

var (
	titleStyle = lipgloss.NewStyle().
			Width(30).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#EE6C4D"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF206E"))
)

func InitialModel() *model {
	dirEntries := store.FindAllPresets()
	choices := make([]string, len(dirEntries))
	for i, v := range dirEntries {
		choices[i] = v.Name()
	}

	return &model{
		choices: choices,
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("ENVY") + "\n\n")

	for i, choice := range m.choices {
		if i == m.cursor {
			s.WriteString(selectedStyle.Render(choice))
		} else {
			s.WriteString(choice)
		}
		s.WriteString("\n")
	}

	s.WriteString("\nPress q to quit.\n")
	return s.String()
}
