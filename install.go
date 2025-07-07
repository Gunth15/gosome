package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	InstallState int
	ErrorMsg     struct{ error }
	SuccessMsg   struct{ string }
	ToolMsg      struct{ string }
	PkgMsg       struct{ string }
	TemplateMsg  struct{ string }
)

const (
	Installing InstallState = iota
	Success
)

type Install struct {
	// downloads 0=tools 1=pkg 2=templates
	downloads []bool
	spinner   spinner.Model
	state     InstallState
}

func (m Model) updateInstallation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrorMsg:
		m.State = End
		m.EndMsg = msg.Error()
	case SuccessMsg:
		m.State = End
		m.EndMsg = msg.string
	case PkgMsg:
		m.Install.downloads[1] = true
		m.done()

	case ToolMsg:
		m.Install.downloads[0] = true
		m.done()
	case TemplateMsg:
		m.Install.downloads[2] = true
		m.done()
	}
	var cmd tea.Cmd
	m.Install.spinner, cmd = m.Install.spinner.Update(msg)
	return m, cmd
}

func (m Model) viewInstallation() string {
	var b strings.Builder

	statIndicator := make([]string, 3)
	for i, finished := range m.Install.downloads {
		if finished {
			statIndicator[i] = " ✅\n"
		} else {
			statIndicator[i] = " " + m.Install.spinner.View() + "\n"
		}
	}

	b.WriteString("🛠 Installing tools: ")
	b.WriteString(statIndicator[0])
	b.WriteString("📦 Installing packages:")
	b.WriteString(statIndicator[1])
	b.WriteString("📐 Applying templates:")
	b.WriteString(statIndicator[2])

	b.WriteString("\n\nPress q or ctrl+c to quit")
	return b.String()
}

func (m Model) done() (tea.Model, tea.Msg) {
	allDone := true
	for _, done := range m.Install.downloads {
		allDone = done && allDone
		if !allDone {
			return m, nil
		}
	}
	return m, func() tea.Msg { return SuccessMsg{"Setup completed successfully. Happy building!"} }
}
