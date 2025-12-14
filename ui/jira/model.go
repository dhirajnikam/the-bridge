package jira

import (
	"encoding/base64"
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

type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
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
	l := list.New([]list.Item{
		item{title: "Setup Required", desc: "Please set JIRA_URL, JIRA_EMAIL, JIRA_TOKEN"},
	}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Jira Issues"
	l.SetShowHelp(false)

	return Model{
		list: l,
	}
}

// -- Messages --

type issuesFetchedMsg []JiraIssue
type errMsg error

// -- Commands --

func fetchIssues() tea.Cmd {
	return func() tea.Msg {
		baseURL := os.Getenv("JIRA_URL")
		email := os.Getenv("JIRA_EMAIL")
		token := os.Getenv("JIRA_TOKEN")

		if baseURL == "" || email == "" || token == "" {
			// Return nil or a special msg indicating no config
			return nil
		}

		// Search for assigned issues
		url := fmt.Sprintf("%s/rest/api/3/search?jql=assignee=currentUser()", baseURL)

		req, _ := http.NewRequest("GET", url, nil)
		auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return errMsg(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return errMsg(fmt.Errorf("API Error: %s", resp.Status))
		}

		var result JiraSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return errMsg(err)
		}

		return issuesFetchedMsg(result.Issues)
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
				title: fmt.Sprintf("%s %s", issue.Key, issue.Fields.Summary),
				desc:  fmt.Sprintf("Status: %s", issue.Fields.Status.Name),
			})
		}
		if len(items) > 0 {
			m.list.SetItems(items)
		} else {
			m.list.SetItems([]list.Item{item{title: "No issues found", desc: "You have no assigned issues."}})
		}
		m.loading = false

	case errMsg:
		m.err = msg
		m.loading = false
		m.list.SetItems([]list.Item{item{title: "Error", desc: msg.Error()}})
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}
