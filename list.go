package main

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	CheckListStyle = lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)
	TitleStyle     = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("pink"))
	SelectedStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#b392ac"))
)

type CheckList struct {
	cursor   int
	options  []Option
	selected []bool
	title    string
	pagesize int
	height   int
	width    int
}

// Sub initialization functions for each scene
func (m Model) initCheckList() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("Gosome"), tea.WindowSize())
}

func (m Model) updateCheckList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.Checklist.cursor > 0 {
				m.Checklist.cursor--
			}

		case "down":
			if m.Checklist.cursor < len(m.Checklist.options) {
				m.Checklist.cursor++
			}
		case "enter":
			if m.Checklist.cursor == len(m.Checklist.options) {
				m.State = Installation
				// setup map and install everything
				return m.StartInstall()
			}
			m.Checklist.selected[m.Checklist.cursor] = !m.Checklist.selected[m.Checklist.cursor]
		}
	case tea.WindowSizeMsg:
		m.Checklist.height = msg.Height
		m.Checklist.width = msg.Width
		m.Checklist.pagesize = max(msg.Height-4, 1)
	}
	return m, nil
}

func (m Model) viewCheckList() string {
	optslen := len(m.Checklist.options)
	if optslen == 0 {
		panic("No options set in config")
	}

	page := m.Checklist.cursor / m.Checklist.pagesize
	start := page * m.Checklist.pagesize
	end := min(start+m.Checklist.pagesize, len(m.Checklist.options))

	var builder strings.Builder

	builder.WriteString(TitleStyle.Render(m.Checklist.title))
	builder.WriteByte('\n')

	for i, selected := range m.Checklist.selected[start:end] {
		cursor_pos := start + i
		element := lipgloss.NewStyle().
			Width(m.Checklist.width)

		var optbuilder strings.Builder
		option := m.Checklist.options[cursor_pos]

		optbuilder.WriteString(option.Title())
		optbuilder.WriteByte('\n')
		optbuilder.WriteString(option.Description())

		if m.Checklist.cursor == cursor_pos {
			element = element.Inherit(PrimaryColor)
		}

		if selected {
			element = element.Inherit(SelectedStyle)
		}

		builder.WriteString(element.Border(lipgloss.ASCIIBorder(), true, true, true, true).Render(optbuilder.String()))
		builder.WriteByte('\n')
	}

	builder.WriteByte('\n')

	done := "[ Done ]"
	if m.Checklist.cursor == len(m.Checklist.options) {
		done = SecondaryColor.Render(done)
	}
	builder.WriteString(done)
	builder.WriteString(lipgloss.NewStyle().Width(m.Checklist.width - 10).Align(lipgloss.Right).Render(strconv.Itoa(page + 1)))

	return CheckListStyle.Height(m.Checklist.height).Render(builder.String())
}
