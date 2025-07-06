package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"

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
			pagesize: 1,
			height:   1,
		},
		Install: Install{
			downloads: make([]bool, 3),
			state:     Installing,
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
		case "q", "ctrl+c":
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
	dir := "testing"
	module := "testing"

	cmds := make([]tea.Cmd, 0)

	// create new project directory and go mod
	err := os.Mkdir(dir, 0755)
	if err != nil {
		m.Install.state = Error
		return m, func() tea.Msg {
			return ErrorMsg{fmt.Errorf("could not create project directory %s, may already exist", dir)}
		}
	}

	// TODO: Update this to take a project name
	init := exec.Command("go", "mod", "init", module)
	init.Dir = dir
	err = init.Run()
	if err != nil {
		m.Install.state = Error
		return m, func() tea.Msg { return ErrorMsg{fmt.Errorf("could not initialize go module %s", module)} }
	}

	pkgs := make([]string, 0)
	tools := make([]string, 0)
	files := make([]string, 0)
	for i, selected := range m.Checklist.selected {
		if selected {
			// packages, and tool dependencies are downloaded and the realted template is copied to the new project
			option := &m.Checklist.options[i]

			pkgs = append(pkgs, option.Deps.Package...)
			tools = append(tools, option.Deps.Tools...)
			files = append(files, option.File)

		}
	}

	cmds = append(cmds, m.Install.spinner.Tick)
	cmds = append(cmds, func() tea.Msg { return DownloadPkgs(dir, pkgs...) })
	cmds = append(cmds, func() tea.Msg { return DownloadTools(dir, tools...) })
	cmds = append(cmds, func() tea.Msg { return CopyFromTemplate(dir, files...) })

	return m, tea.Batch(cmds...)
}

// DownloadPkgs downlads dependencies
func DownloadPkgs(project_dir string, deps ...string) tea.Msg {
	cmd := exec.Command("go", append([]string{"get"}, deps...)...)
	cmd.Dir = project_dir
	err := cmd.Run()
	if err != nil {
		return ErrorMsg{fmt.Errorf("could not install all packages %s", err)}
	}
	return PkgMsg{fmt.Sprintf("Packages Installed: %s", deps)}
}

func DownloadTools(projectDir string, deps ...string) tea.Msg {
	cmd := exec.Command("go", append([]string{"get", "--tool"}, deps...)...)
	cmd.Dir = projectDir
	err := cmd.Run()
	if err != nil {
		return ErrorMsg{fmt.Errorf("could not install tool %s", err)}
	}
	return ToolMsg{fmt.Sprintf("Tools installed: %s", deps)}
}

// CopyFromTemplate copys a file to the specified directory form the template
func CopyFromTemplate(projectDir string, filepaths ...string) tea.Msg {
	for _, filepath := range filepaths {

		info, err := fs.Stat(ProjectTemplate, filepath)
		if err != nil {
			return ErrorMsg{err}
		}

		if info.IsDir() {
			err := copydirectory(projectDir, filepath)
			if err != nil {
				return ErrorMsg{err}
			}
		} else {
			if err = copyToProject(projectDir, filepath); err != nil {
				return err
			}
		}
	}
	return TemplateMsg{fmt.Sprintf("All files downloaded %s", filepaths)}
}

func copydirectory(projectDir string, filepath string) error {
	err := os.Mkdir(projectDir+filepath, 0755)
	if err != nil {
		return err
	}

	dirs, err := ProjectTemplate.ReadDir(filepath)
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		fullPath := fmt.Sprintf("%s/%s", filepath, entry.Name())
		if entry.IsDir() {
			err := copydirectory(projectDir, fullPath)
			if err != nil {
				return err
			}
		} else {
			if err = copyToProject(projectDir, fullPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyToProject(projectDir string, projectPath string) error {
	buff, err := ProjectTemplate.ReadFile(projectPath)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", projectDir, projectPath)
	if err = os.WriteFile(path, buff, 0755); err != nil {
		return err
	}

	return nil
}
