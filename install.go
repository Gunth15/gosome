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
	Error
	Success
)

type Install struct {
	// downloads 0=tools 1=pkg 2=templates
	downloads []bool
	attempt   string
	spinner   spinner.Model
	state     InstallState
}

func (m Model) updateInstallation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrorMsg:
		m.Install.state = Error
		m.Install.attempt = msg.Error()
	case SuccessMsg:
		m.Install.state = Success
		m.Install.attempt = msg.string
	case PkgMsg:
		m.Install.downloads[1] = true
	case ToolMsg:
		m.Install.downloads[0] = true
	case TemplateMsg:
		m.Install.downloads[2] = true
	}
	return m, nil
}

func (m Model) viewInstallation() string {
	var b strings.Builder

	statIndicator := make([]string, 3)
	for i, finished := range m.Install.downloads {
		if finished {
			statIndicator[i] = "✅\n"
		} else {
			statIndicator[i] = m.Install.spinner.View() + "\n"
		}
	}

	b.WriteString("🛠 Installing tools")
	b.WriteString(statIndicator[0])
	b.WriteString("\n📦 Installing packages:\n")
	b.WriteString(statIndicator[1])
	b.WriteString("\n📐 Applying templates:\n")
	b.WriteString(statIndicator[2])

	b.WriteString("\n\nAttempt: " + m.Install.attempt)
	return b.String()
}
