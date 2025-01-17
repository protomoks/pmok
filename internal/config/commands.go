package config

import (
	tea "github.com/charmbracelet/bubbletea"
)

type CreateProjectErrorMsg struct {
}

func (e CreateProjectErrorMsg) Error() string {
	return "Looks like a project already exists. Found a protomok directory in your working directory"
}

type CreateProjectMsg struct {
	Root        string
	ProjectName string
}

func CreateProject(m ManifestConfig) tea.Cmd {
	return func() tea.Msg {
		root, err := InitializeProject(&m)
		if err != nil {
			return CreateProjectErrorMsg{}
		}
		return CreateProjectMsg{Root: root, ProjectName: m.Project.Name}
	}
}
