package main

import (
	"embed"
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed project_template
var projectTemplate embed.FS

//go:embed config.json
var config []byte

func main() {
	var conf Config
	err := json.Unmarshal(config, &conf)
	if err != nil {
		panic(fmt.Sprintf("Invalid configuration: %s", err))
	}
	p := tea.NewProgram(initModel(conf.Options))
	if _, err := p.Run(); err != nil {
		panic("FUCK")
	}
}
