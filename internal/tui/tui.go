package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
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

const (
	normal state = iota
	confirmation
	nameInput
	editing
)

type model struct {
	state       state
	action      tea.Cmd
	editingFile string
	presets     []presetKV
	footer      string
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
	redColor    = lipgloss.Color("#E63946")
	sandColor   = lipgloss.Color("#C1B070")

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

var INSTRUCTIONS = footerStyle.Render(
	"j/k: up/down ・ a: add preset ・ e: edit preset ・ d: delete preset ・ u: use preset(copy)\n",
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
		index := len(name) - len(".env")
		presets[i] = presetKV{name: name[:index], kv: kvArr}
	}

	ti := textinput.New()
	ti.Width = 20

	ta := textarea.New()

	return &model{
		presets:   presets,
		textInput: ti,
		textArea:  ta,
		cursor:    0,
		footer:    INSTRUCTIONS,
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
		m.footer = INSTRUCTIONS
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

				m.textArea.SetValue(getFileContent(m.editingFile))

				return m, m.textArea.Focus()

			case "d":
				m.textInput.Reset()
				m.textInput.Prompt = fmt.Sprintf(
					"Do want to remove %s (y/n)? ",
					m.presets[m.cursor].name,
				)
				path := store.GetPresetsFolderPath() + m.presets[m.cursor].name + ".env"
				m.action = removeFile(path)
				m.state = confirmation
				return m, m.textInput.Focus()

			case "u":
				path := store.GetPresetsFolderPath() + m.presets[m.cursor].name + ".env"
				if _, err := os.Stat(".env"); err == nil {
					m.textInput.Reset()
					m.textInput.Prompt = "Found a .env in this folder. Do you want to remove it (y/n)? "
					m.state = confirmation
					m.action = copyFile(path, ".env")
					return m, m.textInput.Focus()
				}

			case "l":
				// link (confirmation)
			}

		case nameInput:
			switch msg.String() {
			case "enter":
				m.textInput.Blur()
				path := store.GetPresetsFolderPath() + m.textInput.Value() + ".env"
				m.editingFile = m.textInput.Value()
				return m, presetExists(path)

			case "ctrl+c", "esc":
				m.state = normal
				m.textInput.Blur()
			}

		case editing:
			switch msg.String() {
			case "ctrl+s":
				m.state = normal
				return m, saveFile(m.editingFile, m.textArea.Value())

			case "ctrl+c", "esc":
				m.state = normal
			}

		case confirmation:
			switch msg.String() {
			case "enter":
				if m.textInput.Value() == "n" {
					m.state = normal
					return m, nil
				}
				if m.textInput.Value() != "y" {
					m.textInput.Reset()
					return m, nil
				}
				m.textInput.Reset()
				return m, m.action

			case "ctrl+c", "esc":
				m.state = normal
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case someMsg:
		if msg.success && msg.data == "remove" {
			i := m.cursor
			m.presets = append(m.presets[:i], m.presets[i+1:]...)
			if m.cursor >= len(m.presets) {
				m.cursor = len(m.presets) - 1
			}
		} else if msg.success && msg.data == "add" {
			m.textArea.Reset()
			m.state = editing
			return m, m.textArea.Focus()
		} else if msg.success && msg.data == "file created" {
			kvArr := make([][]string, 0)
			path := store.GetPresetsFolderPath() + m.editingFile + ".env"
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				kvArr = append(kvArr, strings.Split(line, "="))
			}

			var exists bool
			for i, v := range m.presets {
				if v.name == m.editingFile {
					exists = true
					m.presets[i] = presetKV{name: v.name, kv: kvArr}
				}
			}

			if !exists {
				m.presets = append(m.presets, presetKV{
					name: m.editingFile,
					kv:   kvArr,
				})
				sort.Slice(m.presets, func(i, j int) bool {
					return m.presets[i].name < m.presets[j].name
				})
			}
			m.footer = lipgloss.NewStyle().Foreground(sandColor).Render(msg.data)
		} else if msg.success {
			m.footer = lipgloss.NewStyle().Foreground(sandColor).Render(msg.data)
		} else if !msg.success {
			m.footer = lipgloss.NewStyle().Foreground(redColor).Render(msg.data)
		}
		m.state = normal
		m.action = nil
		return m, nil

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
			m.footer,
		)
	}

	return lipgloss.NewStyle().
		Margin(0, 0).
		Align(lipgloss.Top).
		AlignVertical(lipgloss.Left).
		Render(view.String())
}
