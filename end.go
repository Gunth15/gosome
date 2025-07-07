package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateEndMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	default:
		return m, nil
	}
}

func (m Model) viewEndMsg() string {
	return m.EndMsg
}
