package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "tab", "enter", "down":
			m.cursor++
			if key == "enter" && m.cursor == len(m.Inputs) {
				m.State = Select
				return m, nil
			}

			if m.cursor > len(m.Inputs)-1 {
				m.cursor = 0
			}
		case "up", "shift+tab":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.Inputs) - 1
			}
		}
	}

	cmds := make([]tea.Cmd, len(m.Inputs))
	for i := 0; i <= len(m.Inputs)-1; i++ {
		if i == m.cursor {
			// Set focused state
			cmds[i] = m.Inputs[i].Focus()
			continue
		}
		// Remove focused state
		m.Inputs[i].Blur()
	}
	for i, input := range m.Inputs {
		m.Inputs[i], cmds[i] = input.Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) viewInputs() string {
	var b strings.Builder
	for i, input := range m.Inputs {
		switch i {
		case 0:
			b.WriteString("Choose package name\n")
			b.WriteString(input.View() + "\n")
		case 1:
			b.WriteString("Choose directory name\n")
			b.WriteString(input.View() + "\n")
		}
	}
	return fmt.Sprintf("Enter a package name for your project\n\n%s", b.String())
}
