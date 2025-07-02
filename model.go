package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	Select State = iota
	Installation
)

var (
	PrimaryColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7d6f86"))
	SecondaryColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#99907d"))
)

type Model struct {
	Checklist CheckList
	Install   Install
	State     State
}

func initModel(options []Option) Model {
	return Model{
		Checklist: CheckList{
			options:  options,
			selected: make([]bool, len(options)),
			cursor:   0,
			width:    1,
			height:   1,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return m.initCheckList()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If recieve signal to quit, exit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctr-c":
			return m, tea.Quit
		}
	}

	// Else handle state
	switch m.State {
	case Select:
		return m.updateCheckList(msg)
	case Installation:
		return m.updateInstallation(msg)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.State {
	case Select:
		return m.viewCheckList()
	case Installation:
		return m.viewInstallation()
	default:
		return "Invalid state"
	}
}

func (m Model) StartInstall() (tea.Model, tea.Cmd) {
	install := &m.Install
	cmds := make([]tea.Cmd, 0)

	// create new project directory and go mod
	err := os.Mkdir("testing", 0664)
	if err != nil {
		m.Install.state = Error
		return m, func() tea.Msg { return ErrorMsg{fmt.Errorf("could not create project directory %s", "testing")} }
	}

	// TODO: Update this to take a project name
	init := exec.Command("go", "mod", "init", "test")
	init.Dir = "./testing"
	err = init.Run()
	if err != nil {
		m.Install.state = Error
		return m, func() tea.Msg { return ErrorMsg{fmt.Errorf("could not initialize go module %s", "test")} }
	}

	for i, selected := range m.Checklist.selected {
		if selected {
			// packages, and tool dependencies are downloaded and the realted template is copied to the new project
			option := &m.Checklist.options[i]
			pkgs := option.Deps.Package
			tools := option.Deps.Tools

			// copy from template
			cmds = append(cmds, func() tea.Msg {
				// TODO: Update this to take a project name
				// TODO: Update model to take input for project name
				return CopyFromTemplate("testing", option.File)
			})

			// install all packages
			for _, pkg := range pkgs {
				cmds = append(cmds, func() tea.Msg {
					return DownloadPkg(pkg)
				})
				install.pkg[PkgMsg{pkg}] = false
			}

			// install all tools
			for _, tool := range tools {
				cmds = append(cmds, func() tea.Msg {
					return DownloadTool(tool)
				})
				install.tools[ToolMsg{tool}] = false
			}
		}
	}
	cmds = append(cmds, m.Install.spinner.Tick)
	return m, tea.Batch(cmds...)
}

// DownloadPkg downlad dependencies
func DownloadPkg(dep string) tea.Msg {
	err := exec.Command("go", "get", dep).Run()
	if err != nil {
		return ErrorMsg{fmt.Errorf("could not install package %s", dep)}
	}
	return PkgMsg{dep}
}

func DownloadTool(dep string) tea.Msg {
	err := exec.Command("go", "get", "--tool", dep).Run()
	if err != nil {
		return ErrorMsg{fmt.Errorf("could not install tool %s", dep)}
	}
	return ToolMsg{dep}
}

// CopyFromTemplate copys a file to the specified directory form the template
func CopyFromTemplate(projectdir string, filepath string) tea.Msg {
	err := fs.WalkDir(projectTemplate, ".", func(path string, d fs.DirEntry, err error) error {
		// get path after the project_template directory
		split := strings.SplitN(path, "/", 1)

		if split[1] == filepath {
			///Read contents of original file
			file, err := projectTemplate.Open(path)
			if err != nil {
				return fmt.Errorf("could not open %s", filepath)
			}
			var buff []byte
			_, err = file.Read(buff)
			if err != nil {
				return fmt.Errorf("could not read contents of template %s", filepath)
			}
			file.Close()
			///
			///

			err = os.WriteFile(projectdir+"/"+filepath, buff, 0664)
			if err != nil {
				return fmt.Errorf("could not create new file %s", filepath)
			}
		}
		return nil
	})
	if err != nil {
		return ErrorMsg{err}
	}
	return TemplateMsg{filepath}
}
