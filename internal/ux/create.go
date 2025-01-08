package ux

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/namegenerator"
)

const ASCIIART = `

  ___         _                 _   
 | _ \_ _ ___| |_ ___ _ __  ___| |__
 |  _/ '_/ _ \  _/ _ \ '  \/ _ \ / /
 |_| |_| \___/\__\___/_|_|_\___/_\_\
                                    

`

const maxWidth = 80

type state int

const (
	stateSucces state = 1
	stateError  state = 2
)

type CreateProjectModel struct {
	logger      io.Writer
	state       state
	lg          *lipgloss.Renderer
	styles      *Styles
	form        *huh.Form
	width       int
	projectRoot string
	projectName string
}

func NewCreateProjectModel(logger io.Writer) CreateProjectModel {
	m := CreateProjectModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.logger = logger

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Placeholder("It's optional. I can generate one for you :)").
				Prompt("Project Name: "),
			huh.NewSelect[string]().
				Key("format").
				Options(huh.NewOptions("yaml", "json")...).
				Title("Choose your config format"),
		),
	).
		WithWidth(80).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m CreateProjectModel) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m CreateProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		}
	case config.CreateProjectMsg:
		m.state = stateSucces
		m.projectRoot = msg.Root
		m.projectName = msg.ProjectName
	case config.CreateProjectErrorMsg:
		m.state = stateError
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		cmds = append(cmds, config.CreateProject(m.getManifestWithDefaults()))
	}

	if m.state == stateSucces || m.state == stateError {
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m CreateProjectModel) View() string {
	s := m.styles

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Render(v)
	header := m.appBoundaryView("Create Project")
	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))

	switch m.state {
	case stateError:
		return s.Base.Render(
			s.ErrorHeaderText.Render("There was an error. It could be because you already have a project created.") + "\n",
		)
	case stateSucces:
		return s.Base.Render(
			s.SuccessText.Render(fmt.Sprintf("Success!\nProject %s has been created.\n", m.projectName)),
		)

	}

	return s.Base.Render(header + "\n" + form + "\n\n" + footer + "\n")
}

func (m CreateProjectModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m CreateProjectModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m CreateProjectModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func (m CreateProjectModel) getManifestWithDefaults() config.Manifest {
	projectName := m.form.GetString("name")
	format := m.form.GetString("format")
	if projectName == "" {
		projectName = namegenerator.Generate()
	}

	if format == string(config.ConfigYaml) {
		return config.NewYAMLConfig(config.Manifest{
			Project: config.Project{
				Name: projectName,
			},
		})
	}
	return config.NewJSONConfig(config.Manifest{
		Project: config.Project{
			Name: projectName,
		},
	})

}
