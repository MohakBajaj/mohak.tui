package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk color palette - vibrant neon on dark
var Colors = struct {
	// Base
	Background string
	Foreground string

	// Neon colors
	Neon   string // Hot pink/magenta
	Cyan   string // Electric cyan
	Yellow string // Warning yellow
	Green  string // Matrix green
	Orange string // Neon orange
	Red    string // Alert red
	Purple string // Deep purple
	Blue   string // Electric blue

	// UI colors
	Muted        string
	Dim          string
	Border       string
	BorderBright string
	Highlight    string

	// Text variants
	BodyText      string
	UserText      string
	AssistantText string
}{
	Background: "#0d0d12",
	Foreground: "#e8f0f8", // Bright white-blue

	// Neon cyberpunk colors - MORE SATURATED
	Neon:   "#ff2a6d", // Hot pink
	Cyan:   "#00ffff", // Pure electric cyan
	Yellow: "#ffff00", // Pure neon yellow
	Green:  "#00ff41", // Matrix green (brighter)
	Orange: "#ff9500", // Bright neon orange
	Red:    "#ff0055", // Neon red
	Purple: "#bf00ff", // Electric purple
	Blue:   "#00d4ff", // Bright neon blue

	Muted:        "#7a8a9a", // Brighter muted - used for borders
	Dim:          "#556677", // Cyan-tinted dim
	Border:       "#2a3040", // Dark border
	BorderBright: "#446688", // Bright cyan-tinted border
	Highlight:    "#1a2535", // Deep blue highlight

	// Text variants
	BodyText:      "#d0e0f0", // Soft white-blue
	UserText:      "#c0e8ff", // Light cyan tint
	AssistantText: "#e0f0e8", // Light green tint
}

// Styles contains all lipgloss styles for the TUI
type Styles struct {
	// Base
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// Text
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Body     lipgloss.Style
	Muted    lipgloss.Style
	Dim      lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Info     lipgloss.Style

	// Neon colors
	Neon   lipgloss.Style
	Cyan   lipgloss.Style
	Yellow lipgloss.Style
	Green  lipgloss.Style
	Orange lipgloss.Style
	Red    lipgloss.Style
	Purple lipgloss.Style
	Blue   lipgloss.Style

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
	Glitch   lipgloss.Style
	Scanline lipgloss.Style
}

// Manager handles styles
type Manager struct {
	styles   Styles
	width    int
	height   int
	renderer *lipgloss.Renderer
}

// NewManager creates a theme manager with an optional renderer
// If renderer is nil, uses the default lipgloss renderer
func NewManager(width, height int, renderer *lipgloss.Renderer) *Manager {
	m := &Manager{
		width:    width,
		height:   height,
		renderer: renderer,
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

// newStyle creates a new style using the session renderer if available
func (m *Manager) newStyle() lipgloss.Style {
	if m.renderer != nil {
		return m.renderer.NewStyle()
	}
	return lipgloss.NewStyle()
}

func (m *Manager) buildStyles() {
	// Base styles
	m.styles.App = m.newStyle().
		Background(lipgloss.Color(Colors.Background)).
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.Header = m.newStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.Footer = m.newStyle().
		Foreground(lipgloss.Color(Colors.Muted))

	// Text styles
	m.styles.Title = m.newStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.Subtitle = m.newStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.Body = m.newStyle().
		Foreground(lipgloss.Color(Colors.BodyText))

	m.styles.Muted = m.newStyle().
		Foreground(lipgloss.Color(Colors.Muted))

	m.styles.Dim = m.newStyle().
		Foreground(lipgloss.Color(Colors.Dim))

	m.styles.Error = m.newStyle().
		Foreground(lipgloss.Color(Colors.Red)).
		Bold(true)

	m.styles.Success = m.newStyle().
		Foreground(lipgloss.Color(Colors.Green))

	m.styles.Warning = m.newStyle().
		Foreground(lipgloss.Color(Colors.Yellow))

	m.styles.Info = m.newStyle().
		Foreground(lipgloss.Color(Colors.Cyan))

	// Neon color styles
	m.styles.Neon = m.newStyle().Foreground(lipgloss.Color(Colors.Neon))
	m.styles.Cyan = m.newStyle().Foreground(lipgloss.Color(Colors.Cyan))
	m.styles.Yellow = m.newStyle().Foreground(lipgloss.Color(Colors.Yellow))
	m.styles.Green = m.newStyle().Foreground(lipgloss.Color(Colors.Green))
	m.styles.Orange = m.newStyle().Foreground(lipgloss.Color(Colors.Orange))
	m.styles.Red = m.newStyle().Foreground(lipgloss.Color(Colors.Red))
	m.styles.Purple = m.newStyle().Foreground(lipgloss.Color(Colors.Purple))
	m.styles.Blue = m.newStyle().Foreground(lipgloss.Color(Colors.Blue))

	// Interactive styles
	m.styles.Prompt = m.newStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.Input = m.newStyle().
		Foreground(lipgloss.Color(Colors.Foreground))

	m.styles.Command = m.newStyle().
		Foreground(lipgloss.Color(Colors.Green)).
		Bold(true)

	m.styles.CommandHint = m.newStyle().
		Foreground(lipgloss.Color(Colors.Muted)).
		Italic(true)

	// Chat styles
	m.styles.UserLabel = m.newStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.UserMessage = m.newStyle().
		Foreground(lipgloss.Color(Colors.UserText))

	m.styles.AssistantLabel = m.newStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.AssistantMessage = m.newStyle().
		Foreground(lipgloss.Color(Colors.AssistantText))

	// Component styles
	m.styles.Border = m.newStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Colors.BorderBright))

	m.styles.Box = m.newStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Colors.Cyan)).
		Padding(0, 1)

	m.styles.Tag = m.newStyle().
		Foreground(lipgloss.Color(Colors.Background)).
		Background(lipgloss.Color(Colors.Cyan)).
		Padding(0, 1).
		Bold(true)

	m.styles.Link = m.newStyle().
		Foreground(lipgloss.Color(Colors.Blue)).
		Underline(true)

	m.styles.Highlight = m.newStyle().
		Foreground(lipgloss.Color(Colors.Yellow)).
		Bold(true)

	// Cyberpunk specific
	m.styles.Glitch = m.newStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Background(lipgloss.Color(Colors.Highlight))

	m.styles.Scanline = m.newStyle().
		Foreground(lipgloss.Color(Colors.Dim))
}
