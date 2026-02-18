package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fdanctl/envy/internal/store"
)

type presetKV struct {
	name string
	kv   [][]string
}

type state = int8

type (
	errMsg error
)

const (
	normal state = iota
	confirmation
	nameInput
	editing
)

type model struct {
	state       state
	editingFile string
	presets     []presetKV
	cursor      int
	textInput   textinput.Model
	textArea    textarea.Model
	width       int
	height      int
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

	ti := textinput.New()
	ti.Width = 20

	ta := textarea.New()

	return &model{
		presets:   presets,
		textInput: ti,
		textArea:  ta,
		cursor:    0,
		width:     0,
		height:    0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case normal:
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

			case "a":
				m.textInput.Reset()
				m.textInput.Prompt = "New name: "
				m.state = nameInput
				return m, m.textInput.Focus()

			case "e":
				m.textArea.Reset()
				m.state = editing
				m.editingFile = m.presets[m.cursor].name
				return m, m.textArea.Focus()

			case "d":
				m.textInput.Reset()
				m.textInput.Prompt = fmt.Sprintf(
					"Do want to remove %s (y/n)? ",
					m.presets[m.cursor].name,
				)
				m.state = confirmation
				return m, m.textInput.Focus()

			case "c":
				// copy (confirmation)

			case "l":
				// link (confirmation)
			}

		case nameInput:
			switch msg.String() {
			case "enter":
				m.state = normal
				m.textInput.Blur()
				// check if file name exist
				// open textArea

			case "ctrl+c", "esc":
				m.state = normal
				m.textInput.Blur()
			}

		case editing:
			switch msg.String() {
			case "ctrl+s":
				m.state = normal
				// save file

			case "ctrl+c", "esc":
				m.state = normal
			}

		case confirmation:
			switch msg.String() {
			case "enter":
				if m.textInput.Value() == "n" {
					m.state = normal
				}
				if m.textInput.Value() != "y" {
					m.textInput.Reset()
					return m, nil
				}
				// do action

			case "ctrl+c", "esc":
				m.state = normal
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case errMsg:
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
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

	var right strings.Builder
	if m.state == editing {
		right.WriteString(m.textArea.View())
	} else {
		rows := m.presets[m.cursor].kv
		t := table.New().
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

		right.WriteString(t.String())
	}

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

	view.WriteString("\n")
	switch m.state {
	case nameInput, confirmation:
		view.WriteString(m.textInput.View())

	case editing:
		view.WriteString(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().
					Width(m.width/2).
					Render(
						"",
					),
				lipgloss.NewStyle().
					Width(m.width/4).
					Render(
						" "+m.editingFile,
					),
				footerStyle.
					Width(m.width/4).
					Align(lipgloss.Right).
					Render(
						"ctrl+s: save ・ esc: cancel ",
					),
			),
		)

	default:
		view.WriteString(
			footerStyle.Render(
				"j/k: up/down ・ a: add preset ・ e: edit preset ・ d: delete preset ・ c: copy preset to current folder ・ l: link preset to current folder\n",
			),
		)
	}

	return lipgloss.NewStyle().
		Margin(0, 0).
		Align(lipgloss.Top).
		AlignVertical(lipgloss.Left).
		Render(view.String())
}
