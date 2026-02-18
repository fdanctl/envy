package tui

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fdanctl/envy/internal/store"
)

type presetKV struct {
	name string
	kv   [][]string
}

type model struct {
	presets []presetKV
	cursor  int
	width   int
	height  int
}

var (
	purple      = lipgloss.Color("#957FB8")
	lightGray   = lipgloss.Color("241")
	numberColor = lipgloss.Color("#D27E99")
	stringColor = lipgloss.Color("#98BB6C")
	keyColor    = lipgloss.Color("#E6C384")
	boolColor   = lipgloss.Color("#FFA066")

	titleStyle = lipgloss.NewStyle().
			Width(30).
			Bold(true).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			Foreground(lipgloss.Color("#68C8DE"))
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF206E"))
	footerStyle = lipgloss.NewStyle().
			AlignVertical(lipgloss.Bottom).
			Foreground(lightGray)
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(purple)
)

func InitialModel() *model {
	dirEntries := store.FindAllPresets()
	presets := make([]presetKV, len(dirEntries))
	presetsPath := store.GetPresetsFolderPath()

	for i, v := range dirEntries {
		name := v.Name()
		kvArr := make([][]string, 0)
		f, err := os.Open(presetsPath + name)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			kvArr = append(kvArr, strings.Split(line, "="))
		}
		presets[i] = presetKV{name: name, kv: kvArr}
	}

	return &model{
		presets: presets,
		cursor:  0,
		width:   0,
		height:  0,
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
			if m.cursor < len(m.presets)-1 {
				m.cursor++
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading...\n"
	}
	var view strings.Builder
	view.WriteString(titleStyle.Width(m.width-2).Render("ENVY") + "\n")
	var left strings.Builder

	for i, choice := range m.presets {
		if i == m.cursor {
			left.WriteString(selectedStyle.Render(choice.name))
		} else {
			left.WriteString(choice.name)
		}
		left.WriteString("\n")
	}

	// var right strings.Builder
	rows := m.presets[m.cursor].kv

	right := table.New().
		Border(lipgloss.NormalBorder()).
		BorderColumn(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}
			if col == 0 {
				return lipgloss.NewStyle().
					Align(lipgloss.Right).
					Foreground(keyColor)
			}
			if len(rows[row]) > 1 {
				v := rows[row][1]
				_, err := strconv.Atoi(v)
				if err == nil {
					return lipgloss.NewStyle().Foreground(numberColor)
				}
				_, err = strconv.ParseBool(v)
				if err == nil {
					return lipgloss.NewStyle().Foreground(boolColor)
				}
			}
			return lipgloss.NewStyle().Foreground(stringColor)
		}).
		Headers("KEY", "VALUE").
		Rows(rows...)

	view.WriteString(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().
				Height(m.height-7).
				Width(m.width/2-2).
				Border(lipgloss.NormalBorder()).
				Render(left.String()),
			lipgloss.NewStyle().
				Height(m.height-7).
				Width(m.width/2-2).
				Border(lipgloss.NormalBorder()).
				Render(right.String()),
		),
	)

	view.WriteString(footerStyle.Render("\nPress q to quit.\n"))
	return lipgloss.NewStyle().
		Margin(0, 0).
		Align(lipgloss.Top).
		AlignVertical(lipgloss.Left).
		Render(view.String())
}
