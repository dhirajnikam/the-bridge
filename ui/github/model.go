package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -- Data Structures --

type GitHubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
}

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// -- Model --

type Model struct {
	list    list.Model
	loading bool
	err     error
}

func New() Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "GitHub Issues (charmbracelet/bubbletea)"
	l.SetShowHelp(false)

	return Model{
		list: l,
	}
}

// -- Messages --

type issuesFetchedMsg []GitHubIssue
type errMsg error

// -- Commands --

func fetchIssues() tea.Cmd {
	return func() tea.Msg {
		// Example: Fetch issues from Bubble Tea repo
		// In a real app, you'd read a config for the repo
		url := "https://api.github.com/repos/charmbracelet/bubbletea/issues?state=open&per_page=10"

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("User-Agent", "TermiFlow")

		// Optional: Add token if present
		token := os.Getenv("GITHUB_TOKEN")
		if token != "" {
			req.Header.Add("Authorization", "Bearer "+token)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return errMsg(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return errMsg(fmt.Errorf("API Error: %s", resp.Status))
		}

		var issues []GitHubIssue
		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			return errMsg(err)
		}

		return issuesFetchedMsg(issues)
	}
}

// -- Update --

func (m Model) Init() tea.Cmd {
	return fetchIssues()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)

	case issuesFetchedMsg:
		var items []list.Item
		for _, issue := range msg {
			items = append(items, item{
				title: fmt.Sprintf("#%d %s", issue.Number, issue.Title),
				desc:  fmt.Sprintf("by %s [%s]", issue.User.Login, issue.State),
			})
		}
		m.list.SetItems(items)
		m.loading = false

	case errMsg:
		m.err = msg
		m.loading = false
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	// We can add a spinner here if m.loading
	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}
