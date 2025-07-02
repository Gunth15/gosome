package main

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	CheckListStyle = lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)
	TitleStyle     = lipgloss.NewStyle().Bold(true)
	SelectedStyle  = lipgloss.NewStyle().BorderLeftBackground(lipgloss.Color("#99907d"))
)

type CheckList struct {
	cursor   int
	options  []Option
	selected []bool
	title    string
	width    int
	height   int
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
	}
	return m, nil
}

func (m Model) viewCheckList() string {
	optslen := len(m.Checklist.options)
	if optslen == 0 {
		panic("No options set in config")
	}

	elementheight := m.Checklist.height / optslen

	pagesize := (elementheight * optslen) / m.Checklist.height
	page := pagesize % m.Checklist.cursor

	var builder strings.Builder

	builder.WriteString(TitleStyle.Render(m.Checklist.title))
	builder.WriteByte('\n')

	for i, selected := range m.Checklist.selected[page : pagesize+page] {
		element := lipgloss.NewStyle().Height(elementheight)

		var optbuilder strings.Builder
		option := m.Checklist.options[i]

		optbuilder.WriteString(option.Title())
		optbuilder.WriteByte('\n')
		optbuilder.WriteString(option.Description())

		if m.Checklist.cursor == i {
			element = element.Inherit(PrimaryColor)
		}

		if selected {
			element = element.Inherit(SelectedStyle)
		}

		builder.WriteString(element.Render(optbuilder.String()))
		builder.WriteByte('\n')
	}

	builder.WriteByte('\n')
	builder.WriteString("[ Done ]")
	builder.WriteString(lipgloss.NewStyle().AlignVertical(lipgloss.Bottom).AlignHorizontal(lipgloss.Right).Render(strconv.Itoa(page + 1)))

	return CheckListStyle.Render(builder.String())
}
