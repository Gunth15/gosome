package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	Select State = iota
	Name
	Installation
	End
)

var (
	PrimaryColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7d6f86"))
	SecondaryColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#99907d"))
)

type Model struct {
	cursor    int
	Checklist CheckList
	Inputs    []textinput.Model
	Install   Install
	EndMsg    string
	State     State
}

func initModel(options []Option) Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	download := make([]bool, 3)
	for i := range download {
		download[i] = true
	}

	tboxes := make([]textinput.Model, 2)
	for i := range tboxes {
		tboxes[i] = textinput.New()
		tboxes[i].Width = 50
		switch i {
		case 0:
			tboxes[i].Focus()
			tboxes[i].Placeholder = "Server"
		case 1:
			tboxes[i].Placeholder = "Web"
		}
	}

	return Model{
		Checklist: CheckList{
			title:    "Select your features",
			options:  options,
			selected: make([]bool, len(options)),
			cursor:   0,
			pagesize: 1,
			height:   1,
		},
		Install: Install{
			downloads: download,
			state:     Installing,
			spinner:   s,
		},
		Inputs: tboxes,
		cursor: 0,
		State:  Name,
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
	case Name:
		return m.updateInputs(msg)
	case Select:
		return m.updateCheckList(msg)
	case Installation:
		return m.updateInstallation(msg)
	case End:
		return m.updateEndMsg(msg)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.State {
	case Name:
		return m.viewInputs()
	case Select:
		return m.viewCheckList()
	case Installation:
		return m.viewInstallation()
	case End:
		return m.viewEndMsg()
	default:
		return "Invalid state"
	}
}

func (m Model) StartInstall() (tea.Model, tea.Cmd) {
	module := m.Inputs[0].Value()
	dir := m.Inputs[1].Value()

	cmds := make([]tea.Cmd, 0)

	// create new project directory and go mod
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return m, func() tea.Msg {
			return ErrorMsg{fmt.Errorf("could not create project directory %s, may already exist", dir)}
		}
	}

	init := exec.Command("go", "mod", "init", module)
	init.Dir = dir
	err = init.Run()
	if err != nil {
		return m, func() tea.Msg { return ErrorMsg{fmt.Errorf("could not initialize go module %s", module)} }
	}

	pkgs := make([]string, 0)
	tools := make([]string, 0)
	files := make([]string, 0)

	m.Install.downloads[2] = false
	files = append(files, "README.md", "internal/api", "cmd", "sqlc.yaml", "env")

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
	if len(pkgs) != 0 {
		m.Install.downloads[1] = false
		cmds = append(cmds, func() tea.Msg { return DownloadPkgs(dir, pkgs...) })
	}
	if len(tools) != 0 {
		m.Install.downloads[0] = false
		cmds = append(cmds, func() tea.Msg { return DownloadTools(dir, tools...) })
	}
	if len(files) != 0 {
		cmds = append(cmds, func() tea.Msg { return CopyFromTemplate(dir, files...) })
	}

	return m, tea.Batch(cmds...)
}

// DownloadPkgs downlads dependencies
func DownloadPkgs(projectDir string, deps ...string) tea.Msg {
	cmd := exec.Command("go", append([]string{"get"}, deps...)...)
	cmd.Dir = projectDir
	err := cmd.Run()
	if err != nil {
		return ErrorMsg{fmt.Errorf("could not install all packages %s", err)}
	}
	return PkgMsg{fmt.Sprintf("Packages Installed: %s", deps)}
}

func DownloadTools(projectDir string, deps ...string) tea.Msg {
	for _, tool := range deps {
		cmd := exec.Command("go", "get", "--tool", tool)
		cmd.Dir = projectDir
		err := cmd.Run()
		if err != nil {
			return ErrorMsg{fmt.Errorf("could not install tool %s", err)}
		}
	}
	return ToolMsg{fmt.Sprintf("Tools installed: %s", deps)}
}

// CopyFromTemplate copys a file to the specified directory form the template
func CopyFromTemplate(projectDir string, filepaths ...string) tea.Msg {
	for _, filepath := range filepaths {

		info, err := fs.Stat(ProjectTemplate, getTemplatePath(filepath))
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
				return ErrorMsg{err}
			}
		}
	}
	return TemplateMsg{fmt.Sprintf("All files downloaded %s", filepaths)}
}

func copydirectory(projectDir string, filepath string) error {
	err := os.MkdirAll(projectDir+"/"+filepath, 0755)
	if err != nil {
		return fmt.Errorf("cannot create directory %s in new project %s", filepath, projectDir)
	}

	dirs, err := ProjectTemplate.ReadDir(getTemplatePath(filepath))
	if err != nil {
		return fmt.Errorf("cannot find file %s in project template directory", filepath)
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
	buff, err := ProjectTemplate.ReadFile(getTemplatePath(projectPath))
	if err != nil {
		return fmt.Errorf("cannot find file %s in project template directory: %s", projectPath, err)
	}

	path := fmt.Sprintf("%s/%s", projectDir, projectPath)
	if err = os.WriteFile(path, buff, 0755); err != nil {
		return fmt.Errorf("cannot write file to %s: %s", path, err)
	}

	return nil
}

func getTemplatePath(filePath string) string {
	if filePath[0:2] == "./" {
		filePath = filePath[2:]
	}
	return "project_template/" + filePath
}
