package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
)

func center(text string, width int) string {
	w := lipgloss.Width(text)
	if w >= width {
		return text
	}
	pad := (width - w) / 2
	return strings.Repeat(" ", pad) + text
}

func box(title string, content []string, styles theme.Styles, width int) string {
	var b strings.Builder
	boxWidth := min(60, width-8)
	innerWidth := boxWidth - 4

	// Top border with title
	titleLen := min(len(title), innerWidth-4)
	titlePad := (innerWidth - titleLen) / 2
	if titlePad < 1 {
		titlePad = 1
	}
	top := styles.Cyan.Render("┌") +
		styles.Dim.Render(strings.Repeat("─", titlePad)) +
		styles.Cyan.Bold(true).Render(" "+title[:min(len(title), titleLen)]+" ") +
		styles.Dim.Render(strings.Repeat("─", max(1, innerWidth-titlePad-titleLen))) +
		styles.Cyan.Render("┐")
	b.WriteString(center(top, width))
	b.WriteString("\n")

	// Content lines
	for _, line := range content {
		lineWidth := lipgloss.Width(line)
		padding := innerWidth - lineWidth
		if padding < 0 {
			// Line too long, it will overflow but we can't easily truncate styled text
			padding = 0
		}
		row := styles.Dim.Render("│ ") + line + strings.Repeat(" ", padding) + styles.Dim.Render(" │")
		b.WriteString(center(row, width))
		b.WriteString("\n")
	}

	// Bottom border
	bottom := styles.Cyan.Render("└") + styles.Dim.Render(strings.Repeat("─", innerWidth+2)) + styles.Cyan.Render("┘")
	b.WriteString(center(bottom, width))

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

// WelcomeMessage renders centered welcome screen
func WelcomeMessage(styles theme.Styles, width int) string {
	var b strings.Builder

	// Glitch-style ASCII banner
	banner := []string{
		"╔╦╗╔═╗╦ ╦╔═╗╦╔═ ╔═╗╦ ╦",
		"║║║║ ║╠═╣╠═╣╠╩╗ ╚═╗╠═╣",
		"╩ ╩╚═╝╩ ╩╩ ╩╩ ╩o╚═╝╩ ╩",
	}

	b.WriteString("\n\n")
	for i, line := range banner {
		var style lipgloss.Style
		switch i % 3 {
		case 0:
			style = styles.Neon
		case 1:
			style = styles.Cyan
		case 2:
			style = styles.Neon
		}
		b.WriteString(center(style.Bold(true).Render(line), width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(center(styles.Dim.Render("▓▒░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▒▓"), width))
	b.WriteString("\n\n")

	// Command hints in a box
	cmdLines := []string{
		styles.Green.Render("/about") + styles.Dim.Render("    ─ ") + styles.Muted.Render("view profile data"),
		styles.Yellow.Render("/projects") + styles.Dim.Render(" ─ ") + styles.Muted.Render("browse projects"),
		styles.Neon.Render("/resume") + styles.Dim.Render("   ─ ") + styles.Muted.Render("credentials file"),
		styles.Purple.Render("/help") + styles.Dim.Render("     ─ ") + styles.Muted.Render("system commands"),
		"",
		styles.Cyan.Render("or just type to chat with AI"),
	}
	b.WriteString(box("COMMANDS", cmdLines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// Help renders help screen
func Help(styles theme.Styles, width int) string {
	var b strings.Builder

	b.WriteString("\n")

	commands := []string{
		styles.Purple.Bold(true).Render("/help") + styles.Dim.Render("     ─── ") + styles.Muted.Render("show this help"),
		styles.Green.Bold(true).Render("/about") + styles.Dim.Render("    ─── ") + styles.Muted.Render("about me"),
		styles.Yellow.Bold(true).Render("/projects") + styles.Dim.Render(" ─── ") + styles.Muted.Render("list projects"),
		styles.Yellow.Bold(true).Render("/open <id>") + styles.Dim.Render("─── ") + styles.Muted.Render("view project"),
		styles.Orange.Bold(true).Render("/exp") + styles.Dim.Render("      ─── ") + styles.Muted.Render("work experience"),
		styles.Neon.Bold(true).Render("/resume") + styles.Dim.Render("   ─── ") + styles.Muted.Render("full resume"),
		styles.Cyan.Bold(true).Render("/clear") + styles.Dim.Render("    ─── ") + styles.Muted.Render("clear chat"),
		styles.Red.Bold(true).Render("/exit") + styles.Dim.Render("     ─── ") + styles.Muted.Render("disconnect"),
		"",
		styles.Dim.Render("─────────────────────────────────"),
		"",
		styles.Cyan.Render("ESC") + styles.Dim.Render("       ─── ") + styles.Muted.Render("go back / cancel"),
		styles.Cyan.Render("ENTER") + styles.Dim.Render("     ─── ") + styles.Muted.Render("send message"),
		styles.Green.Render("Ctrl+S") + styles.Dim.Render("    ─── ") + styles.Muted.Render("toggle mouse/select"),
	}

	b.WriteString(box("SYS_HELP", commands, styles, width))
	b.WriteString("\n\n")

	b.WriteString(center(styles.Dim.Render("// type anything to chat with AI"), width))
	b.WriteString("\n")

	return b.String()
}

// About renders about screen
func About(styles theme.Styles, bio string, width int) string {
	var b strings.Builder
	maxLineLen := 48 // Max chars per line in box

	b.WriteString("\n")

	var lines []string
	bioLines := strings.Split(bio, "\n")
	for _, line := range bioLines {
		if strings.HasPrefix(line, "# ") {
			// Title - skip, we use box title
			continue
		} else if strings.HasPrefix(line, "## ") {
			title := strings.TrimPrefix(line, "## ")
			lines = append(lines, "")
			lines = append(lines, styles.Cyan.Bold(true).Render("◈ "+title))
		} else if strings.HasPrefix(line, "- **") && strings.Contains(line, "**") {
			// Bold item: - **key** — value
			parts := strings.SplitN(line, "**", 3)
			if len(parts) >= 3 {
				rest := parts[2]
				if len(rest) > maxLineLen-len(parts[1])-6 {
					rest = rest[:maxLineLen-len(parts[1])-9] + "..."
				}
				lines = append(lines, styles.Green.Render("▸ ")+styles.Neon.Bold(true).Render(parts[1])+styles.Body.Render(rest))
			}
		} else if strings.HasPrefix(line, "- ") {
			text := strings.TrimPrefix(line, "- ")
			// Handle inline **bold**
			text = renderInlineBold(text, styles)
			if lipgloss.Width(text) > maxLineLen-2 {
				// Wrap it
				lines = append(lines, styles.Green.Render("▸ ")+text[:maxLineLen-5]+"...")
			} else {
				lines = append(lines, styles.Green.Render("▸ ")+text)
			}
		} else if line != "" {
			// Handle inline **bold**
			processed := renderInlineBold(line, styles)
			// Wrap long lines
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word) > maxLineLen {
					lines = append(lines, renderInlineBold(currentLine, styles))
					currentLine = word
				} else {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += word
				}
			}
			if currentLine != "" {
				// Check if we already rendered inline or need to
				if processed != line {
					lines = append(lines, renderInlineBold(currentLine, styles))
				} else {
					lines = append(lines, styles.Body.Render(currentLine))
				}
			}
		}
	}

	b.WriteString(box("PROFILE_DATA", lines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// renderInlineBold handles **bold** text inline
func renderInlineBold(text string, styles theme.Styles) string {
	result := text
	for {
		start := strings.Index(result, "**")
		if start == -1 {
			break
		}
		end := strings.Index(result[start+2:], "**")
		if end == -1 {
			break
		}
		end += start + 2
		boldText := result[start+2 : end]
		result = result[:start] + styles.Neon.Bold(true).Render(boldText) + result[end+2:]
	}
	if result == text {
		return styles.Body.Render(text)
	}
	return result
}

// ProjectsList renders projects list
func ProjectsList(styles theme.Styles, projects *content.Projects, width int) string {
	var b strings.Builder

	b.WriteString("\n")

	var lines []string
	for i, p := range projects.Projects {
		// Status indicator
		var statusStyle lipgloss.Style
		var statusIcon string
		switch p.Status {
		case "active":
			statusStyle = styles.Green
			statusIcon = "●"
		case "completed":
			statusStyle = styles.Cyan
			statusIcon = "◈"
		default:
			statusStyle = styles.Yellow
			statusIcon = "○"
		}

		header := styles.Dim.Render(fmt.Sprintf("[%d] ", i+1)) +
			styles.Neon.Bold(true).Render(p.Name) + " " +
			statusStyle.Render(statusIcon)
		lines = append(lines, header)

		lines = append(lines, styles.Dim.Render("    ID: ")+styles.Muted.Render(p.ID))

		desc := p.Description
		if len(desc) > 45 {
			desc = desc[:42] + "..."
		}
		lines = append(lines, styles.Dim.Render("    ")+styles.Body.Render(desc))

		// Tech tags
		var tags string
		colorCycle := []lipgloss.Style{styles.Cyan, styles.Neon, styles.Green, styles.Yellow}
		for j, tech := range p.Tech {
			if j < 3 {
				tags += colorCycle[j%4].Render("⟨"+tech+"⟩") + " "
			}
		}
		lines = append(lines, styles.Dim.Render("    ")+tags)
		lines = append(lines, "")
	}

	lines = append(lines, styles.Dim.Render("─────────────────────────────────────"))
	lines = append(lines, styles.Muted.Render("use /open <id> to view details"))

	b.WriteString(box("PROJECTS", lines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// ProjectDetail renders project details
func ProjectDetail(styles theme.Styles, project *content.Project, width int) string {
	if project == nil {
		return center(styles.Red.Render("⚠ PROJECT_NOT_FOUND"), width)
	}

	var b strings.Builder

	b.WriteString("\n")

	var lines []string

	// Status
	var statusStyle lipgloss.Style
	var statusText string
	switch project.Status {
	case "active":
		statusStyle = styles.Green
		statusText = "● ACTIVE"
	case "completed":
		statusStyle = styles.Cyan
		statusText = "◈ ARCHIVED"
	default:
		statusStyle = styles.Yellow
		statusText = "○ IN_PROGRESS"
	}
	lines = append(lines, styles.Dim.Render("STATUS: ")+statusStyle.Bold(true).Render(statusText))
	lines = append(lines, "")

	// Description
	lines = append(lines, styles.Cyan.Bold(true).Render("◈ DESCRIPTION"))
	words := strings.Fields(project.Description)
	line := ""
	for _, word := range words {
		if len(line)+len(word) > 45 {
			lines = append(lines, styles.Body.Render("  "+line))
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, styles.Body.Render("  "+line))
	}
	lines = append(lines, "")

	// Tech
	lines = append(lines, styles.Green.Bold(true).Render("◈ TECH_STACK"))
	var tags string
	colorCycle := []lipgloss.Style{styles.Cyan, styles.Neon, styles.Green, styles.Yellow}
	for i, tech := range project.Tech {
		tags += colorCycle[i%4].Render("⟨"+tech+"⟩") + " "
	}
	lines = append(lines, "  "+tags)
	lines = append(lines, "")

	// Links
	if project.Links.Demo != "" || project.Links.Github != "" {
		lines = append(lines, styles.Yellow.Bold(true).Render("◈ LINKS"))
		if project.Links.Demo != "" {
			lines = append(lines, styles.Dim.Render("  DEMO:   ")+styles.Link.Render(project.Links.Demo))
		}
		if project.Links.Github != "" {
			lines = append(lines, styles.Dim.Render("  SOURCE: ")+styles.Link.Render(project.Links.Github))
		}
	}

	b.WriteString(box(project.Name, lines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// Resume renders resume
func Resume(styles theme.Styles, resume *content.Resume, width int) string {
	var b strings.Builder

	b.WriteString("\n")

	var lines []string

	// Header
	lines = append(lines, center(styles.Neon.Bold(true).Render(resume.Name), 50))
	lines = append(lines, center(styles.Cyan.Render(resume.Title), 50))
	if resume.Tagline != "" {
		lines = append(lines, center(styles.Muted.Italic(true).Render("\""+resume.Tagline+"\""), 50))
	}
	lines = append(lines, "")

	// Contact
	contact := styles.Green.Render("✉ ") + styles.Body.Render(resume.Contact.Email)
	lines = append(lines, center(contact, 50))
	if resume.Contact.Website != "" {
		web := styles.Cyan.Render("⚡ ") + styles.Link.Render(resume.Contact.Website)
		lines = append(lines, center(web, 50))
	}
	github := styles.Purple.Render("◈ ") + styles.Body.Render(resume.Contact.Github)
	lines = append(lines, center(github, 50))
	lines = append(lines, "")

	lines = append(lines, styles.Dim.Render("─────────────────────────────────────────"))
	lines = append(lines, "")

	// Summary
	lines = append(lines, styles.Purple.Bold(true).Render("◈ SUMMARY"))
	words := strings.Fields(resume.Summary)
	line := ""
	for _, word := range words {
		if len(line)+len(word) > 45 {
			lines = append(lines, styles.Body.Render("  "+line))
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, styles.Body.Render("  "+line))
	}
	lines = append(lines, "")

	// Skills
	lines = append(lines, styles.Cyan.Bold(true).Render("◈ SKILLS"))
	skillLine := func(skills []string, style lipgloss.Style, limit int) string {
		var s string
		for i, skill := range skills {
			if i < limit {
				s += style.Render("⟨"+skill+"⟩") + " "
			}
		}
		return s
	}
	lines = append(lines, "  "+skillLine(resume.Skills.Languages, styles.Neon, 5))
	lines = append(lines, "  "+skillLine(resume.Skills.Frontend, styles.Cyan, 4))
	lines = append(lines, "  "+skillLine(resume.Skills.Backend, styles.Green, 4))
	lines = append(lines, "  "+skillLine(resume.Skills.DevOps, styles.Yellow, 4))
	lines = append(lines, "")

	// Education
	lines = append(lines, styles.Yellow.Bold(true).Render("◈ EDUCATION"))
	for _, edu := range resume.Education {
		lines = append(lines, "  "+styles.Neon.Bold(true).Render(edu.Degree))
		lines = append(lines, "  "+styles.Cyan.Render(edu.Institution)+styles.Dim.Render(", "+edu.Location))
		lines = append(lines, "  "+styles.Dim.Render(edu.Period)+" │ "+styles.Green.Render(edu.Score))
		lines = append(lines, "")
	}

	// Achievements
	if len(resume.Achievements) > 0 {
		lines = append(lines, styles.Green.Bold(true).Render("◈ ACHIEVEMENTS"))
		for i, ach := range resume.Achievements {
			if i < 3 {
				a := ach
				if len(a) > 45 {
					a = a[:42] + "..."
				}
				lines = append(lines, styles.Neon.Render("  ▸ ")+styles.Body.Render(a))
			}
		}
	}

	b.WriteString(box("CREDENTIALS", lines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// Experience renders work experience
func Experience(styles theme.Styles, resume *content.Resume, width int) string {
	var b strings.Builder

	b.WriteString("\n")

	var lines []string

	lines = append(lines, center(styles.Neon.Bold(true).Render("WORK EXPERIENCE"), 50))
	lines = append(lines, center(styles.Muted.Render(resume.Name), 50))
	lines = append(lines, "")
	lines = append(lines, styles.Dim.Render("─────────────────────────────────────────"))
	lines = append(lines, "")

	for i, exp := range resume.Experience {
		// Company & Role header
		roleTag := styles.Neon.Bold(true).Render(exp.Role)
		companyTag := styles.Cyan.Bold(true).Render(exp.Company)
		lines = append(lines, roleTag)
		lines = append(lines, styles.Dim.Render("@ ")+companyTag)
		lines = append(lines, styles.Muted.Render("  "+exp.Period))
		lines = append(lines, "")

		// Highlights
		for _, h := range exp.Highlights {
			hl := h
			if len(hl) > 45 {
				hl = hl[:42] + "..."
			}
			lines = append(lines, styles.Green.Render("  ▸ ")+styles.Body.Render(hl))
		}

		// Separator between experiences
		if i < len(resume.Experience)-1 {
			lines = append(lines, "")
			lines = append(lines, styles.Dim.Render("  ─────────────────────────────────"))
			lines = append(lines, "")
		}
	}

	b.WriteString(box("EXPERIENCE", lines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// ChatMessage renders a chat message
func ChatMessage(styles theme.Styles, role, content string, width int, mdRenderer *MarkdownRenderer) string {
	var b strings.Builder

	if role == "user" {
		b.WriteString(styles.Cyan.Bold(true).Render("┌─ YOU ────────────────────────────"))
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render("│ ") + styles.Body.Render(content))
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render("└──────────────────────────────────"))
	} else {
		b.WriteString(styles.Neon.Bold(true).Render("┌─ MOHAK.AI ───────────────────────"))
		b.WriteString("\n")
		rendered := mdRenderer.Render(content)
		lines := strings.Split(rendered, "\n")
		for _, line := range lines {
			b.WriteString(styles.Dim.Render("│ ") + line)
			b.WriteString("\n")
		}
		b.WriteString(styles.Dim.Render("└──────────────────────────────────"))
	}
	b.WriteString("\n")

	return b.String()
}

// StreamingMessage renders streaming AI response
func StreamingMessage(styles theme.Styles, content string, width int, mdRenderer *MarkdownRenderer) string {
	var b strings.Builder

	b.WriteString(styles.Neon.Bold(true).Render("┌─ MOHAK.AI ") + styles.Neon.Render("▓▒░ streaming ░▒▓"))
	b.WriteString("\n")

	if content != "" {
		rendered := mdRenderer.RenderStreaming(content)
		lines := strings.Split(rendered, "\n")
		for _, line := range lines {
			b.WriteString(styles.Dim.Render("│ ") + line)
			b.WriteString("\n")
		}
		b.WriteString(styles.Dim.Render("│ ") + styles.Neon.Render("▌"))
	} else {
		b.WriteString(styles.Dim.Render("│ ") + styles.Neon.Render("▓▒░ ") + styles.Muted.Render("initializing...") + styles.Neon.Render(" ░▒▓"))
	}
	b.WriteString("\n")
	b.WriteString(styles.Dim.Render("└──────────────────────────────────"))
	b.WriteString("\n")

	return b.String()
}

// Error renders error
func Error(styles theme.Styles, message string) string {
	return styles.Red.Render("⚠ ERR: " + message)
}

// Success renders success
func Success(styles theme.Styles, message string) string {
	return styles.Green.Render("✓ " + message)
}
