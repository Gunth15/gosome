package main

import (
	"fmt"
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
	tools    map[ToolMsg]bool
	pkg      map[PkgMsg]bool
	template map[TemplateMsg]bool
	attempt  string
	spinner  spinner.Model
	state    InstallState
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
		m.Install.pkg[msg] = true
	case ToolMsg:
		m.Install.tools[msg] = true
	case TemplateMsg:
		m.Install.template[msg] = true
	}
	return m, nil
}

func renderToolMap(items map[ToolMsg]bool, spin spinner.Model) string {
	var b strings.Builder
	for k, done := range items {
		status := " " + spin.View()
		if done {
			status = " ✅"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", status, k.string))
	}
	return b.String()
}

func renderPkgMap(items map[PkgMsg]bool, spin spinner.Model) string {
	var b strings.Builder
	for k, done := range items {
		status := " " + spin.View()
		if done {
			status = " ✅"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", status, k.string))
	}
	return b.String()
}

func renderTemplateMap(items map[TemplateMsg]bool, spin spinner.Model) string {
	var b strings.Builder
	for k, done := range items {
		status := " " + spin.View()
		if done {
			status = " ✅"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", status, k.string))
	}
	return b.String()
}

func (m Model) viewInstallation() string {
	var b strings.Builder

	b.WriteString("🛠 Installing tools:\n")
	b.WriteString(renderToolMap(m.Install.tools, m.Install.spinner))

	b.WriteString("\n📦 Installing packages:\n")
	b.WriteString(renderPkgMap(m.Install.pkg, m.Install.spinner))

	b.WriteString("\n📐 Applying templates:\n")
	b.WriteString(renderTemplateMap(m.Install.template, m.Install.spinner))

	b.WriteString("\n\nAttempt: " + m.Install.attempt)
	return b.String()
}
