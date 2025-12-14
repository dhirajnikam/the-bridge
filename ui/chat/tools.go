package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
)

// -- GitHub Tool --

func getGitHubIssues() (map[string]any, error) {
	// Defaults to charmbracelet/bubbletea for demo, or env var
	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		repo = "charmbracelet/bubbletea" // Fallback
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=5", repo)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "TermiFlow")
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API Error: %s", resp.Status)
	}

	var issues []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	// Simplify output
	var simplified []map[string]any
	for _, i := range issues {
		simplified = append(simplified, map[string]any{
			"number": i["number"],
			"title":  i["title"],
			"user":   i["user"].(map[string]any)["login"],
			"state":  i["state"],
		})
	}

	return map[string]any{"issues": simplified}, nil
}

// -- Jira Tool --

func getJiraIssues() (map[string]any, error) {
	baseURL := os.Getenv("JIRA_URL")
	email := os.Getenv("JIRA_EMAIL")
	token := os.Getenv("JIRA_TOKEN")

	if baseURL == "" || email == "" || token == "" {
		return nil, fmt.Errorf("Jira credentials not set (JIRA_URL, JIRA_EMAIL, JIRA_TOKEN)")
	}

	url := fmt.Sprintf("%s/rest/api/3/search?jql=assignee=currentUser()&maxResults=5", baseURL)
	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Jira API Error: %s", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Simplify
	issues := result["issues"].([]interface{})
	var simplified []map[string]any
	for _, i := range issues {
		issue := i.(map[string]interface{})
		fields := issue["fields"].(map[string]interface{})
		simplified = append(simplified, map[string]any{
			"key":     issue["key"],
			"summary": fields["summary"],
			"status":  fields["status"].(map[string]interface{})["name"],
		})
	}

	return map[string]any{"issues": simplified}, nil
}

// Tool Definitions for Gemini
var tools = []*genai.Tool{
	{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "get_github_issues",
				Description: "Get list of open GitHub issues for the configured repository.",
			},
			{
				Name:        "get_jira_issues",
				Description: "Get list of Jira issues assigned to the current user.",
			},
		},
	},
}

var toolFunctions = map[string]func() (map[string]any, error){
	"get_github_issues": getGitHubIssues,
	"get_jira_issues":   getJiraIssues,
}
