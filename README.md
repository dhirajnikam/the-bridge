# The Bridge 🌉

**The Bridge** is a modern, unified Terminal User Interface (TUI) dashboard designed to bring your essential developer tools into one streamlined command center. 

Built with **Go** and **Bubble Tea**, it integrates your local shell, Jira tasks, and GitHub repositories into a single, keyboard-driven interface.

## 🚀 Features

*   **🖥️ Integrated Shell**: A persistent local terminal environment. Execute commands, navigate directories (`cd`), and manage files without leaving the dashboard.
*   **🐞 Jira Integration**: View and track your assigned Jira tickets in real-time.
*   **🐙 GitHub Integration**: Monitor issues and pull requests for your repositories.
*   **⌨️ Keyboard Driven**: Efficient tab-based navigation (`Tab` to switch views).
*   **🎨 Modern UI**: Beautifully styled with `Lipgloss` for a premium terminal experience.

## 🛠️ Installation

### Prerequisites
*   [Go](https://go.dev/dl/) 1.19 or higher.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/dhirajnikam/the-bridge.git
cd the-bridge

# Install dependencies and run
go mod tidy
go run main.go
```

## ⚙️ Configuration

To unlock the full power of The Bridge, you need to set a few environment variables for the API integrations.

| Variable | Description | Example |
| :--- | :--- | :--- |
| **GitHub** | | |
| `GITHUB_TOKEN` | Personal Access Token with repo scope | `ghp_ABC123...` |
| **Jira** | | |
| `JIRA_URL` | Your Jira instance URL | `https://your-domain.atlassian.net` |
| `JIRA_EMAIL` | Email address for Jira account | `user@example.com` |
| `JIRA_TOKEN` | Jira API Token | `ATATT3...` |

**Quick Setup:**
```bash
export GITHUB_TOKEN="your_token"
export JIRA_URL="https://your-org.atlassian.net"
export JIRA_EMAIL="you@example.com"
export JIRA_TOKEN="your_jira_token"
./termiflow
```

## ⌨️ Usage

*   **Switch Tabs**: Press `Tab` to cycle between Shell, Jira, and GitHub.
*   **Shell**: Type commands as normal (`ls`, `pwd`, `echo "hello"`).
*   **Quit**: Press `Ctrl+C`.

## 🏗️ Built With

*   [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework.
*   [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions.
*   [Bubbles](https://github.com/charmbracelet/bubbles) - Component library.

---
*Maintained by Dhiraj Nikam*
