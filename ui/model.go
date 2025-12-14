package ui

import (
	"strings"

	"termiflow/ui/chat"
	"termiflow/ui/github"
	"termiflow/ui/jira"
	"termiflow/ui/shell"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	viewShell sessionState = iota
	viewJira
	viewGitHub
	viewChat
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	tabsBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	tabStyle  = lipgloss.NewStyle().
			Border(tabsBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)
	activeTabStyle   = tabStyle.Copy().Border(tabsBorder, true)
	inactiveTabStyle = tabStyle.Copy().Border(tabsBorder, true).BorderForeground(lipgloss.Color("240"))
)

type Model struct {
	state sessionState
	tabs  []string

	shell  shell.Model
	jira   jira.Model
	github github.Model
	chat   chat.Model

	width  int
	height int
}

func New() Model {
	tabs := []string{"Shell", "Jira", "GitHub", "Chat"}
	return Model{
		state:  viewShell,
		tabs:   tabs,
		shell:  shell.New(),
		jira:   jira.New(),
		github: github.New(),
		chat:   chat.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.shell.Init(),
		m.jira.Init(),
		m.github.Init(),
		m.chat.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.state = (m.state + 1) % sessionState(len(m.tabs))
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Pass size down to sub-models
		// Note: We might want closer control over layout later
		contentHeight := msg.Height - 5 // Approx header height

		// Update Jira and Github list sizes
		m.jira.SetSize(msg.Width, contentHeight)
		m.github.SetSize(msg.Width, contentHeight)
		m.chat.SetSize(msg.Width, contentHeight)

		// Shell handles its own sizing in Update usually, but let's pass it if needed
		// For now shell Update handles WindowSizeMsg directly
	}

	// Update the active model
	switch m.state {
	case viewShell:
		m.shell, cmd = m.shell.Update(msg)
		cmds = append(cmds, cmd)
	case viewJira:
		m.jira, cmd = m.jira.Update(msg)
		cmds = append(cmds, cmd)
	case viewGitHub:
		m.github, cmd = m.github.Update(msg)
		cmds = append(cmds, cmd)
	case viewChat:
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	doc := strings.Builder{}

	// Render Tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		var style lipgloss.Style
		if sessionState(i) == m.state {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n\n")

	// Render Active View
	switch m.state {
	case viewShell:
		doc.WriteString(m.shell.View())
	case viewJira:
		doc.WriteString(m.jira.View())
	case viewGitHub:
		doc.WriteString(m.github.View())
	case viewChat:
		doc.WriteString(m.chat.View())
	}

	return docStyle.Render(doc.String())
}
