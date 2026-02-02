package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/client"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/ui"
)

// View represents an overlay view (when not chatting)
type View int

const (
	ViewChat View = iota
	ViewHelp
	ViewAbout
	ViewProjects
	ViewProjectDetail
	ViewResume
	ViewExperience
)

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    string
	Content string
}

// Model is the main Bubble Tea model
type Model struct {
	width  int
	height int

	themeManager *theme.Manager

	resume   *content.Resume
	projects *content.Projects
	bio      string

	view          View
	selectedProj  string
	errorMessage  string
	statusMessage string

	input    textinput.Model
	viewport viewport.Model

	aiClient     *client.AIClient
	chatHistory  []ChatMessage
	chatResponse *strings.Builder
	isStreaming  bool
	sessionID    string
	showWelcome  bool
	streamCtx    context.Context
	streamCancel context.CancelFunc
	streamMu     *sync.Mutex
	chunkChan    chan string
	errChan      chan error

	mouseEnabled bool
	quitting     bool
}

// Config holds initialization options
type Config struct {
	ThemeManager *theme.Manager
	Resume       *content.Resume
	Projects     *content.Projects
	Bio          string
	AIClient     *client.AIClient
	SessionID    string
	Width        int
	Height       int
}

// NewModel creates a new app model
func NewModel(cfg Config) Model {
	input := textinput.New()
	input.Placeholder = "enter command or chat..."
	input.Focus()
	input.CharLimit = 1000
	input.Width = cfg.Width - 8

	// Header (3) + Footer (5) = 8 lines reserved
	vp := viewport.New(cfg.Width-4, cfg.Height-8)
	vp.Style = lipgloss.NewStyle()

	return Model{
		width:        cfg.Width,
		height:       cfg.Height,
		themeManager: cfg.ThemeManager,
		resume:       cfg.Resume,
		projects:     cfg.Projects,
		bio:          cfg.Bio,
		view:         ViewChat,
		input:        input,
		viewport:     vp,
		aiClient:     cfg.AIClient,
		chatHistory:  make([]ChatMessage, 0),
		chatResponse: &strings.Builder{},
		streamMu:     &sync.Mutex{},
		sessionID:    cfg.SessionID,
		showWelcome:  true,
		mouseEnabled: true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.EnableBracketedPaste,
		func() tea.Msg { return tea.EnableMouseCellMotion() },
	)
}

type StreamChunkMsg struct {
	Chunk string
}

type StreamDoneMsg struct {
	Error error
}

type ClearStatusMsg struct{}

type QuitMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

func quitAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return QuitMsg{}
	})
}

func listenForChunks(ch <-chan string, errCh <-chan error) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			select {
			case err := <-errCh:
				return StreamDoneMsg{Error: err}
			default:
				return StreamDoneMsg{}
			}
		}
		return StreamChunkMsg{Chunk: chunk}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle paste events - pass directly to input
		if msg.Paste {
			var inputCmd tea.Cmd
			m.input, inputCmd = m.input.Update(msg)
			return m, inputCmd
		}
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.quitting = true
			return m, quitAfter(1500 * time.Millisecond)

		case tea.KeyEnter:
			if m.isStreaming {
				return m, nil
			}
			input := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			m.errorMessage = ""
			m.statusMessage = ""
			if input == "" {
				return m, nil
			}
			return m.handleInput(input)

		case tea.KeyEsc:
			if m.isStreaming && m.streamCancel != nil {
				m.streamCancel()
				m.isStreaming = false
				m.streamMu.Lock()
				if m.chatResponse.Len() > 0 {
					m.chatHistory = append(m.chatHistory, ChatMessage{
						Role:    "assistant",
						Content: m.chatResponse.String(),
					})
					m.chatResponse.Reset()
				}
				m.streamMu.Unlock()
				m.updateViewport()
				return m, nil
			}
			if m.view != ViewChat {
				m.view = ViewChat
				// Show welcome if no chat history
				if len(m.chatHistory) == 0 {
					m.showWelcome = true
				}
				m.updateViewport()
			}

		default:
			// Keyboard shortcuts (work anytime)
			switch msg.String() {
			case "ctrl+s":
				m.mouseEnabled = !m.mouseEnabled
				if m.mouseEnabled {
					m.statusMessage = "Mouse ON (scroll mode)"
					return m, tea.Batch(
						func() tea.Msg { return tea.EnableMouseCellMotion() },
						clearStatusAfter(2*time.Second),
					)
				} else {
					m.statusMessage = "Mouse OFF (select mode)"
					return m, func() tea.Msg { return tea.DisableMouse() }
				}
			case "ctrl+h", "ctrl+/":
				m.view = ViewHelp
				m.updateViewport()
				return m, nil
			case "ctrl+a":
				m.view = ViewAbout
				m.updateViewport()
				return m, nil
			case "ctrl+p":
				m.view = ViewProjects
				m.selectedProj = ""
				m.updateViewport()
				return m, nil
			case "ctrl+r":
				m.view = ViewResume
				m.updateViewport()
				return m, nil
			case "ctrl+e":
				m.view = ViewExperience
				m.updateViewport()
				return m, nil
			case "ctrl+w":
				// Go home/welcome
				m.view = ViewChat
				m.showWelcome = len(m.chatHistory) == 0
				m.updateViewport()
				return m, nil
			case "ctrl+l":
				// Clear chat
				m.chatHistory = nil
				m.showWelcome = true
				m.view = ViewChat
				m.errorMessage = ""
				m.statusMessage = ""
				m.updateViewport()
				return m, nil
			case "ctrl+q":
				m.quitting = true
				return m, quitAfter(1500 * time.Millisecond)
			}

			// Number keys for project selection (only in projects view with empty input)
			if m.view == ViewProjects && m.input.Value() == "" {
				switch msg.String() {
				case "1", "2", "3", "4", "5", "6", "7", "8", "9":
					idx := int(msg.String()[0] - '1')
					if idx >= 0 && idx < len(m.projects.Projects) {
						m.selectedProj = m.projects.Projects[idx].ID
						m.view = ViewProjectDetail
						m.updateViewport()
						return m, nil
					}
				}
			}
		}

	case ClearStatusMsg:
		m.statusMessage = ""

	case QuitMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.themeManager.SetSize(msg.Width, msg.Height)
		m.input.Width = msg.Width - 8
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.updateViewport()

	case StreamChunkMsg:
		m.streamMu.Lock()
		m.chatResponse.WriteString(msg.Chunk)
		m.streamMu.Unlock()
		m.updateViewport()
		if m.chunkChan != nil {
			return m, listenForChunks(m.chunkChan, m.errChan)
		}

	case StreamDoneMsg:
		m.isStreaming = false
		m.streamMu.Lock()
		response := m.chatResponse.String()
		m.streamMu.Unlock()
		if msg.Error != nil {
			m.errorMessage = msg.Error.Error()
		} else if response != "" {
			m.chatHistory = append(m.chatHistory, ChatMessage{
				Role:    "assistant",
				Content: response,
			})
		}
		m.chatResponse.Reset()
		m.chunkChan = nil
		m.errChan = nil
		m.updateViewport()
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) handleInput(input string) (tea.Model, tea.Cmd) {
	if strings.HasPrefix(input, "/") {
		return m.handleSlashCommand(input)
	}
	return m.sendChatMessage(input)
}

func (m Model) handleSlashCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "/help", "/h", "/?":
		m.view = ViewHelp
		m.showWelcome = false
	case "/about", "/bio":
		m.view = ViewAbout
		m.showWelcome = false
	case "/projects", "/p":
		m.view = ViewProjects
		m.showWelcome = false
	case "/open", "/o":
		if len(args) == 0 {
			m.errorMessage = "Usage: /open <project-id>"
		} else {
			m.selectedProj = args[0]
			if m.projects.GetProjectByID(m.selectedProj) == nil {
				m.errorMessage = "Project not found: " + m.selectedProj
			} else {
				m.view = ViewProjectDetail
				m.showWelcome = false
			}
		}
	case "/resume", "/cv", "/r":
		m.view = ViewResume
		m.showWelcome = false
	case "/exp", "/experience", "/work":
		m.view = ViewExperience
		m.showWelcome = false
	case "/clear", "/cls":
		m.view = ViewChat
		m.chatHistory = nil
		m.showWelcome = true
		m.errorMessage = ""
		m.statusMessage = ""
	case "/exit", "/quit", "/q":
		m.quitting = true
		return m, quitAfter(1500 * time.Millisecond)
	case "/back", "/b":
		m.view = ViewChat
	default:
		m.errorMessage = "Unknown command: " + command
	}
	m.updateViewport()
	return m, nil
}

func (m Model) sendChatMessage(message string) (tea.Model, tea.Cmd) {
	if m.aiClient == nil {
		m.errorMessage = "AI not available"
		return m, nil
	}

	m.view = ViewChat
	m.showWelcome = false
	m.chatHistory = append(m.chatHistory, ChatMessage{Role: "user", Content: message})
	m.isStreaming = true
	m.chatResponse.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	m.streamCtx = ctx
	m.streamCancel = cancel

	chunkChan := make(chan string, 1000)
	errChan := make(chan error, 1)
	m.chunkChan = chunkChan
	m.errChan = errChan
	m.updateViewport()

	history := make([]client.Message, 0, len(m.chatHistory)-1)
	for _, msg := range m.chatHistory[:len(m.chatHistory)-1] {
		history = append(history, client.Message{Role: msg.Role, Content: msg.Content})
	}

	aiClient := m.aiClient
	sessionID := m.sessionID

	go func() {
		defer close(chunkChan)
		defer close(errChan)
		err := aiClient.ChatStream(ctx, sessionID, message, history, func(chunk string) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case chunkChan <- chunk:
				return nil
			}
		})
		if err != nil {
			errChan <- err
		}
	}()

	return m, listenForChunks(chunkChan, errChan)
}

func (m *Model) updateViewport() {
	styles := m.themeManager.Styles()
	mdRenderer := ui.NewMarkdownRenderer(styles)

	var content string
	switch m.view {
	case ViewChat:
		content = m.buildChatView(styles, mdRenderer)
	case ViewHelp:
		content = ui.Help(styles, m.width)
	case ViewAbout:
		content = ui.About(styles, m.bio, m.width)
	case ViewProjects:
		content = ui.ProjectsList(styles, m.projects, m.width)
	case ViewProjectDetail:
		content = ui.ProjectDetail(styles, m.projects.GetProjectByID(m.selectedProj), m.width)
	case ViewResume:
		content = ui.Resume(styles, m.resume, m.width)
	case ViewExperience:
		content = ui.Experience(styles, m.resume, m.width)
	}

	m.viewport.SetContent(content)
	if m.view == ViewChat {
		m.viewport.GotoBottom()
	}
}

func (m Model) buildChatView(styles theme.Styles, mdRenderer *ui.MarkdownRenderer) string {
	var b strings.Builder

	if m.showWelcome && len(m.chatHistory) == 0 {
		b.WriteString(ui.WelcomeMessage(styles, m.width))
	}

	for _, msg := range m.chatHistory {
		b.WriteString(ui.ChatMessage(styles, msg.Role, msg.Content, m.width, mdRenderer))
		b.WriteString("\n")
	}

	if m.isStreaming {
		m.streamMu.Lock()
		currentResponse := m.chatResponse.String()
		m.streamMu.Unlock()
		b.WriteString(ui.StreamingMessage(styles, currentResponse, m.width, mdRenderer))
	}

	return b.String()
}

func (m Model) View() string {
	if m.quitting {
		return m.renderQuitScreen()
	}

	styles := m.themeManager.Styles()
	var b strings.Builder

	// ╔══════════════════════════════════════════════════════════════════╗
	// ║                           HEADER                                 ║
	// ╠══════════════════════════════════════════════════════════════════╣
	b.WriteString(m.renderHeader(styles))
	b.WriteString("\n")

	// ║                          CONTENT                                 ║
	content := m.viewport.View()
	// Pad content to fill width
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := m.width - 4 - lineWidth
		if padding > 0 {
			lines[i] = styles.Dim.Render("║ ") + line + strings.Repeat(" ", padding) + styles.Dim.Render(" ║")
		} else {
			lines[i] = styles.Dim.Render("║ ") + line + styles.Dim.Render(" ║")
		}
	}
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n")

	// ╠══════════════════════════════════════════════════════════════════╣
	// ║                           FOOTER                                 ║
	// ╚══════════════════════════════════════════════════════════════════╝
	b.WriteString(m.renderFooter(styles))

	return b.String()
}

func (m Model) renderQuitScreen() string {
	styles := m.themeManager.Styles()
	var b strings.Builder

	border := styles.Muted.Render(strings.Repeat("═", m.width-2))
	b.WriteString("\n")
	b.WriteString(styles.Yellow.Render("╔") + border + styles.Yellow.Render("╗"))
	b.WriteString("\n")

	msg := styles.Neon.Bold(true).Render("CONNECTION TERMINATED")
	msgWidth := lipgloss.Width(msg)
	pad := (m.width - 4 - msgWidth) / 2
	b.WriteString(styles.Muted.Render("║ ") + strings.Repeat(" ", pad) + msg + strings.Repeat(" ", m.width-4-pad-msgWidth) + styles.Muted.Render(" ║"))
	b.WriteString("\n")

	sub := styles.Yellow.Render("// session ended")
	subWidth := lipgloss.Width(sub)
	pad2 := (m.width - 4 - subWidth) / 2
	b.WriteString(styles.Muted.Render("║ ") + strings.Repeat(" ", pad2) + sub + strings.Repeat(" ", m.width-4-pad2-subWidth) + styles.Muted.Render(" ║"))
	b.WriteString("\n")

	b.WriteString(styles.Yellow.Render("╚") + border + styles.Yellow.Render("╝"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderHeader(styles theme.Styles) string {
	var b strings.Builder
	innerWidth := m.width - 4

	// Top border - Yellow corners, Muted lines (cyberpunk!)
	topBorder := styles.Yellow.Render("╔") + styles.Muted.Render(strings.Repeat("═", innerWidth+2)) + styles.Yellow.Render("╗")
	b.WriteString(topBorder)
	b.WriteString("\n")

	// Title bar - Yellow/Neon gradient
	logo := styles.Yellow.Bold(true).Render("▓▒░") + styles.Neon.Bold(true).Render(" MOHAK.SH ") + styles.Yellow.Bold(true).Render("░▒▓")

	// View indicator
	viewName := ""
	viewStyle := styles.Muted
	switch m.view {
	case ViewChat:
		viewName = "NEURAL_LINK"
		viewStyle = styles.Green
	case ViewHelp:
		viewName = "SYS_HELP"
		viewStyle = styles.Purple
	case ViewAbout:
		viewName = "PROFILE"
		viewStyle = styles.Cyan
	case ViewProjects:
		viewName = "PROJECTS"
		viewStyle = styles.Yellow
	case ViewProjectDetail:
		viewName = "PROJECT"
		viewStyle = styles.Yellow
	case ViewResume:
		viewName = "CREDENTIALS"
		viewStyle = styles.Neon
	case ViewExperience:
		viewName = "EXPERIENCE"
		viewStyle = styles.Orange
	}

	status := ""
	if m.isStreaming {
		status = styles.Neon.Render("◉ STREAMING")
	} else {
		status = styles.Green.Render("◉ ONLINE")
	}

	viewTag := styles.Yellow.Render("[") + viewStyle.Bold(true).Render(viewName) + styles.Yellow.Render("]")

	// Calculate layout
	logoWidth := lipgloss.Width(logo)
	viewWidth := lipgloss.Width(viewTag)
	statusWidth := lipgloss.Width(status)
	totalContent := logoWidth + viewWidth + statusWidth
	spacing1 := (innerWidth - totalContent) / 2 - 2
	spacing2 := innerWidth - logoWidth - spacing1 - viewWidth - statusWidth

	headerLine := styles.Muted.Render("║ ") + logo + strings.Repeat(" ", max(1, spacing1)) + viewTag + strings.Repeat(" ", max(1, spacing2)) + status + styles.Muted.Render(" ║")
	b.WriteString(headerLine)
	b.WriteString("\n")

	// Bottom border with connectors - Yellow corners
	bottomBorder := styles.Yellow.Render("╠") + styles.Muted.Render(strings.Repeat("═", innerWidth+2)) + styles.Yellow.Render("╣")
	b.WriteString(bottomBorder)

	return b.String()
}

func (m Model) renderFooter(styles theme.Styles) string {
	var b strings.Builder
	innerWidth := m.width - 4

	// Top border - Yellow corners, Muted lines (consistent with header)
	topBorder := styles.Yellow.Render("╠") + styles.Muted.Render(strings.Repeat("═", innerWidth+2)) + styles.Yellow.Render("╣")
	b.WriteString(topBorder)
	b.WriteString("\n")

	// Input line
	prompt := styles.Yellow.Bold(true).Render("❯ ")
	inputView := m.input.View()
	inputLine := prompt + inputView
	inputWidth := lipgloss.Width(inputLine)
	inputPad := innerWidth - inputWidth
	b.WriteString(styles.Muted.Render("║ ") + inputLine + strings.Repeat(" ", max(0, inputPad)) + styles.Muted.Render(" ║"))
	b.WriteString("\n")

	// Separator - Yellow corners
	sep := styles.Yellow.Render("╟") + styles.Muted.Render(strings.Repeat("─", innerWidth+2)) + styles.Yellow.Render("╢")
	b.WriteString(sep)
	b.WriteString("\n")

	// Status/hint line
	var hint string
	if m.errorMessage != "" {
		hint = styles.Red.Bold(true).Render("⚠ ERR: " + m.errorMessage)
	} else if m.statusMessage != "" {
		hint = styles.Green.Bold(true).Render("✓ " + m.statusMessage)
	} else if m.isStreaming {
		hint = styles.Neon.Render("▓▒░") + styles.Cyan.Render(" streaming ") + styles.Neon.Render("░▒▓") + styles.Dim.Render(" │ ") + styles.Yellow.Render("ESC") + styles.Dim.Render(" abort")
	} else if m.view != ViewChat {
		hint = styles.Yellow.Render("ESC") + styles.Dim.Render(" back │ ") +
			styles.Cyan.Render("^W") + styles.Dim.Render(" home │ ") +
			styles.Purple.Render("^H") + styles.Dim.Render(" help")
	} else {
		hint = styles.Green.Render("^A") + styles.Dim.Render(" about ") +
			styles.Yellow.Render("^P") + styles.Dim.Render(" projects ") +
			styles.Orange.Render("^E") + styles.Dim.Render(" exp ") +
			styles.Neon.Render("^R") + styles.Dim.Render(" resume ") +
			styles.Purple.Render("^H") + styles.Dim.Render(" help")
	}
	hintWidth := lipgloss.Width(hint)
	hintPad := innerWidth - hintWidth
	b.WriteString(styles.Muted.Render("║ ") + hint + strings.Repeat(" ", max(0, hintPad)) + styles.Muted.Render(" ║"))
	b.WriteString("\n")

	// Bottom border - Yellow corners, Muted lines
	bottomBorder := styles.Yellow.Render("╚") + styles.Muted.Render(strings.Repeat("═", innerWidth+2)) + styles.Yellow.Render("╝")
	b.WriteString(bottomBorder)

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Unused but kept for potential future use
var _ = fmt.Sprint
