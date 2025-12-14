package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	pathStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
)

type Model struct {
	viewport   viewport.Model
	textInput  textinput.Model
	currentDir string
	err        error
}

func New() Model {
	cwd, _ := os.Getwd()

	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	vp := viewport.New(30, 20)
	vp.SetContent(fmt.Sprintf("Welcome to TermiFlow Shell!\nCurrent Directory: %s\n", cwd))

	return Model{
		textInput:  ti,
		viewport:   vp,
		currentDir: cwd,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmdStr := m.textInput.Value()
			m.textInput.Reset()

			// Execute command
			output, newDir := m.executeCommand(cmdStr)

			// Update directory if changed
			if newDir != "" {
				m.currentDir = newDir
			}

			// Format output
			prompt := fmt.Sprintf("%s $ %s", pathStyle.Render(filepath.Base(m.currentDir)), cmdStr)
			newContent := fmt.Sprintf("%s\n%s\n%s", m.viewport.View(), prompt, output)

			// Handle clearing screen separately if we wanted to
			m.viewport.SetContent(newContent)
			m.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textInput.Width = msg.Width
		m.viewport.Height = msg.Height - 3 // Leave more room for input/header
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s\n%s $ %s",
		m.viewport.View(),
		pathStyle.Render(filepath.Base(m.currentDir)), // Show just the base name for brevity
		m.textInput.View(),
	)
}

func (m Model) executeCommand(input string) (string, string) {
	if strings.TrimSpace(input) == "" {
		return "", ""
	}
	parts := strings.Fields(input)
	cmdName := parts[0]
	cmdArgs := parts[1:]

	// Handle 'cd' manually
	if cmdName == "cd" {
		targetDir := ""
		if len(cmdArgs) > 0 {
			targetDir = cmdArgs[0]
		} else {
			targetDir, _ = os.UserHomeDir()
		}

		// Handle relative paths
		if !filepath.IsAbs(targetDir) {
			targetDir = filepath.Join(m.currentDir, targetDir)
		}

		// Verify it exists
		info, err := os.Stat(targetDir)
		if err != nil || !info.IsDir() {
			return errStyle.Render(fmt.Sprintf("cd: %s: No such directory", cmdArgs[0])), ""
		}

		return "", targetDir
	}

	// External commands
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = m.currentDir
	out, err := cmd.CombinedOutput()

	if err != nil {
		return errStyle.Render(fmt.Sprintf("Error: %s", err)), ""
	}
	return string(out), ""
}
