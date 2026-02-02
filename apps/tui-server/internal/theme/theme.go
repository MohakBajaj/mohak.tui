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
	Neon       string // Hot pink/magenta
	Cyan       string // Electric cyan
	Yellow     string // Warning yellow
	Green      string // Matrix green
	Orange     string // Neon orange
	Red        string // Alert red
	Purple     string // Deep purple
	Blue       string // Electric blue
	
	// UI colors
	Muted       string
	Dim         string
	Border      string
	BorderBright string
	Highlight   string
	
	// Text variants
	BodyText         string
	UserText         string
	AssistantText    string
}{
	Background: "#0d0d12",
	Foreground: "#e8f0f8", // Bright white-blue
	
	// Neon cyberpunk colors - MORE SATURATED
	Neon:       "#ff2a6d", // Hot pink
	Cyan:       "#00ffff", // Pure electric cyan
	Yellow:     "#ffff00", // Pure neon yellow
	Green:      "#00ff41", // Matrix green (brighter)
	Orange:     "#ff9500", // Bright neon orange
	Red:        "#ff0055", // Neon red
	Purple:     "#bf00ff", // Electric purple
	Blue:       "#00d4ff", // Bright neon blue
	
	Muted:        "#7a8a9a", // Brighter muted - used for borders
	Dim:          "#556677", // Cyan-tinted dim
	Border:       "#2a3040", // Dark border
	BorderBright: "#446688", // Bright cyan-tinted border
	Highlight:    "#1a2535", // Deep blue highlight
	
	// Text variants
	BodyText:       "#d0e0f0", // Soft white-blue
	UserText:       "#c0e8ff", // Light cyan tint
	AssistantText:  "#e0f0e8", // Light green tint
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
		Foreground(lipgloss.Color(Colors.BodyText))

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
		Foreground(lipgloss.Color(Colors.Muted)).
		Italic(true)

	// Chat styles
	m.styles.UserLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Cyan)).
		Bold(true)

	m.styles.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.UserText))

	m.styles.AssistantLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.Neon)).
		Bold(true)

	m.styles.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Colors.AssistantText))

	// Component styles
	m.styles.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Colors.BorderBright))

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
