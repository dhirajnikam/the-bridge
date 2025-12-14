package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Message struct {
	Role    string
	Content string
}

type Model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []Message
	client      *genai.Client
	model       *genai.GenerativeModel
	chatSession *genai.ChatSession
	err         error
	initialized bool
}

func New() Model {
	ta := textarea.New()
	ta.Placeholder = "Ask Gemini... (Ensure GEMINI_API_KEY is set)"
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 1000

	ta.SetWidth(50)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter sends message

	vp := viewport.New(50, 10)
	vp.SetContent("Welcome to The Bridge Chat! 🤖\nType a message and press Enter to chat with Gemini.\n")

	return Model{
		textarea: ta,
		viewport: vp,
		messages: []Message{},
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

type errMsg error
type responseMsg string

func (m Model) getKey() string {
	return os.Getenv("GEMINI_API_KEY")
}

func (m *Model) ensureClient() error {
	if m.client != nil {
		return nil
	}
	ctx := context.Background()
	apiKey := m.getKey()
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	m.client = c
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-1.5-flash-002" // Latest stable flash
	}
	m.model = c.GenerativeModel(modelName)
	m.model.Tools = tools
	m.chatSession = m.model.StartChat()
	m.initialized = true
	return nil
}

func (m Model) sendMessage(msg string) tea.Cmd {
	return func() tea.Msg {
		// Because we can't easily modify the model in the command closure if it's not a pointer,
		// and we handled client init lazily.
		// However, m is captured by value.
		// We have to rely on m.client being set.
		// But wait, the client init must happen before or we need a way to pass it.
		// Actually, let's just create a new client or handle state differently?
		// No, we should rely on the Update loop to handle init or check it here.
		// But safely, we should probably do the client init in Update or passed via pointer.
		// Bubble Tea models are value types.
		// So `m.client` will be nil in the command if it was nil when `sendMessage` was called.

		// We'll reload the client here if needed for this closure instance.
		ctx := context.Background()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return errMsg(fmt.Errorf("GEMINI_API_KEY not set"))
		}

		// For the sake of simplicity in this stateless command, let's create a client if we don't carry state well.
		// BUT `chatSession` holds history. We NEED to persist `chatSession`.
		// Since `chatSession` is a pointer in the struct, changes to the struct's pointer fields persist across copies if they point to the same object.
		// So as long as `m.chatSession` was initialized in `Update` *before* `sendMessage` returned this cmd, we are good.
		// But in `Update`, we call `m.sendMessage(userMsg)`.

		return func() tea.Msg {
			// This inner closure executes asynchronously.
			// It needs access to chatSession.
			// If we access m.chatSession, it uses the captured m.
			if m.chatSession == nil {
				return errMsg(fmt.Errorf("Chat session not initialized"))
			}

			resp, err := m.chatSession.SendMessage(ctx, genai.Text(msg))
			if err != nil {
				return errMsg(err)
			}

			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				return errMsg(fmt.Errorf("empty response"))
			}

			// Handle response parts (Text or FunctionCall)
			var responseBuilder strings.Builder
			for _, part := range resp.Candidates[0].Content.Parts {
				switch p := part.(type) {
				case genai.Text:
					responseBuilder.WriteString(string(p))
				case genai.FunctionCall:
					// Execute function
					if fn, ok := toolFunctions[p.Name]; ok {
						// For now, we ignore arguments as our simple tools don't use them or use env vars
						// In a real app, unmarshal p.Args
						res, err := fn()
						if err != nil {
							responseBuilder.WriteString(fmt.Sprintf("\n[Error executing %s: %v]\n", p.Name, err))
						} else {
							// Send result back to model
							// Note: To truly have a conversation loop with tools, we need to send the function response
							// back to the chat session. This simple implementation just shows the result.
							// To do it right:

							// We need to send this back to the model.
							// This requires a more complex loop in `sendMessage` or `Update`.
							// For this simple version, let's just print the JSON result to the chat.
							jsonRes, _ := json.MarshalIndent(res, "", "  ")
							responseBuilder.WriteString(fmt.Sprintf("\n[Tool %s Output]:\n%s\n", p.Name, string(jsonRes)))

							// Ideally: Send function response back to Gemini to get a natural language summary.
							// But that requires chaining commands. Let's start with this.
						}
					} else {
						responseBuilder.WriteString(fmt.Sprintf("\n[Unknown tool: %s]\n", p.Name))
					}
				}
			}

			return responseMsg(responseBuilder.String())
		}()
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textarea.Value() == "" {
				break
			}
			userMsg := m.textarea.Value()

			// Init client if needed
			if err := m.ensureClient(); err != nil {
				m.messages = append(m.messages, Message{Role: "system", Content: fmt.Sprintf("Error: %v", err)})
				m.updateViewport()
				m.textarea.Reset()
				return m, nil
			}

			m.messages = append(m.messages, Message{Role: "user", Content: userMsg})
			m.updateViewport()
			m.textarea.Reset()

			return m, tea.Batch(tiCmd, vpCmd, m.sendMessage(userMsg))
		}
	case responseMsg:
		m.messages = append(m.messages, Message{Role: "model", Content: string(msg)})
		m.updateViewport()
	case errMsg:
		m.messages = append(m.messages, Message{Role: "system", Content: fmt.Sprintf("Error: %v", msg)})
		m.updateViewport()
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *Model) updateViewport() {
	var sb strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(fmt.Sprintf("\nYou: %s\n", msg.Content))
		} else if msg.Role == "model" {
			sb.WriteString(fmt.Sprintf("Gemini: %s\n", msg.Content))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", msg.Content))
		}
	}
	if len(m.messages) > 0 {
		m.viewport.SetContent(sb.String())
		m.viewport.GotoBottom()
	}
}

func (m *Model) SetSize(w, h int) {
	m.textarea.SetWidth(w)
	m.viewport.Width = w
	m.viewport.Height = h - m.textarea.Height() - 2
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}
