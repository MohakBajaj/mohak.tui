package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk color palette
var Colors = struct {
	// Base
	Background string
	Foreground string
	
	// Neon colors
	Neon       string // Hot pink/magenta
	Cyan       string // Electric cyan
	Yellow     string // Warning yellow
	Green      string // Matrix green
	Orange     string // Neon orange
	Red        string // Alert red
	Purple     string // Deep purple
	Blue       string // Electric blue
	
	// UI colors
	Muted      string
	Dim        string
	Border     string
	Highlight  string
}{
	Background: "#0a0a0f",
	Foreground: "#c7d5e0",
	
	// Neon cyberpunk colors
	Neon:       "#ff2a6d", // Hot pink
	Cyan:       "#05d9e8", // Electric cyan
	Yellow:     "#f9f871", // Neon yellow
	Green:      "#39ff14", // Matrix green
	Orange:     "#ff6e27", // Neon orange
	Red:        "#ff073a", // Neon red
	Purple:     "#bd00ff", // Electric purple
	Blue:       "#01c8ee", // Neon blue
	
	Muted:      "#5a6270",
	Dim:        "#2d3138",
	Border:     "#1a1d23",
	Highlight:  "#162029",
}

// Styles contains all lipgloss styles for the TUI
type Styles struct {
	// Base
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// Text
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Dim         lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style

	// Neon colors
	Neon      lipgloss.Style
	Cyan      lipgloss.Style
	Yellow    lipgloss.Style
	Green     lipgloss.Style
	Orange    lipgloss.Style
	Red       lipgloss.Style
	Purple    lipgloss.Style
	Blue      lipgloss.Style

	// Interactive
	Prompt      lipgloss.Style
	Input       lipgloss.Style
	Command     lipgloss.Style
	CommandHint lipgloss.Style

	// Chat
	UserLabel        lipgloss.Style
	UserMessage      lipgloss.Style
	AssistantLabel   lipgloss.Style
	AssistantMessage lipgloss.Style

	// Components
	Border    lipgloss.Style
	Box       lipgloss.Style
	Tag       lipgloss.Style
	Link      lipgloss.Style
	Highlight lipgloss.Style
	
	// Cyberpunk specific
	Glitch    lipgloss.Style
	Scanline  lipgloss.Style
}

// Manager handles styles
type Manager struct {
	styles Styles
	width  int
	height int
}

// NewManager creates a theme manager
func NewManager(width, height int) *Manager {
	m := &Manager{
		width:  width,
		height: height,
	}
	m.buildStyles()
	return m
}

// SetSize updates dimensions and rebuilds styles
func (m *Manager) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.buildStyles()
}

// Styles returns the current styles
func (m *Manager) Styles() Styles {
	return m.styles
}

// Width returns current width
func (m *Manager) Width() int {
	return m.width
}

// Height returns current height
func (m *Manager) Height() int {
	return m.height
}

func (m *Manager) buildStyles() {
	// Base styles
	m.styles.App = lipgloss.NewStyle().
		Background(lipgloss.Color(Colors.Background)).
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.Header = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.Footer = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Muted))

	// Text styles
	m.styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.Subtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.Body = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.Muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Muted))

	m.styles.Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Dim))

	m.styles.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Red)).
		Bold(true)

	m.styles.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Green))

	m.styles.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Yellow))

	m.styles.Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Cyan))

	// Neon color styles
	m.styles.Neon = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Neon))
	m.styles.Cyan = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Cyan))
	m.styles.Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Yellow))
	m.styles.Green = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Green))
	m.styles.Orange = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Orange))
	m.styles.Red = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Red))
	m.styles.Purple = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Purple))
	m.styles.Blue = lipgloss.NewStyle().Foreground(lipgloss.Color(Colors.Blue))

	// Interactive styles
	m.styles.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.Input = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.Command = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Green)).
		Bold(true)

	m.styles.CommandHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Dim)).
		Italic(true)

	// Chat styles
	m.styles.UserLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.AssistantLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Foreground))

	// Component styles
	m.styles.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Colors.Dim))

	m.styles.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Colors.Cyan)).
		Padding(0, 1)

	m.styles.Tag = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Background)).
		Background(lipgloss.Color(Colors.Cyan)).
		Padding(0, 1).
		Bold(true)

	m.styles.Link = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Blue)).
		Underline(true)

	m.styles.Highlight = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Yellow)).
		Bold(true)
	
	// Cyberpunk specific
	m.styles.Glitch = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Background(lipgloss.Color(Colors.Highlight))
	
	m.styles.Scanline = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Dim))
}
