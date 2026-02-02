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

// boxWidth calculates optimal box width based on screen width
func boxWidth(screenWidth int) int {
	// Responsive box sizing
	if screenWidth < 60 {
		return screenWidth - 4
	}
	if screenWidth < 100 {
		return min(60, screenWidth-8)
	}
	return min(70, screenWidth-20)
}

// contentWidth returns usable content width inside a box
func contentWidth(boxW int) int {
	return boxW - 4 // 2 chars for borders on each side
}

func box(title string, lines []string, styles theme.Styles, width int) string {
	var b strings.Builder
	bw := boxWidth(width)
	cw := contentWidth(bw)

	// Top border with title
	titleLen := min(len(title), cw-4)
	titlePad := (cw - titleLen) / 2
	if titlePad < 1 {
		titlePad = 1
	}

	top := styles.Yellow.Render("┌") +
		styles.Muted.Render(strings.Repeat("─", titlePad)) +
		styles.Cyan.Bold(true).Render(" "+title[:min(len(title), titleLen)]+" ") +
		styles.Muted.Render(strings.Repeat("─", max(1, cw-titlePad-titleLen))) +
		styles.Yellow.Render("┐")
	b.WriteString(center(top, width))
	b.WriteString("\n")

	// Content lines
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)

		// Handle lines that are too long
		if lineWidth > cw {
			// Truncate with ellipsis for styled text
			line = TruncateText(line, cw-1)
			lineWidth = lipgloss.Width(line)
		}

		padding := cw - lineWidth
		if padding < 0 {
			padding = 0
		}

		row := styles.Muted.Render("│ ") + line + strings.Repeat(" ", padding) + styles.Muted.Render(" │")
		b.WriteString(center(row, width))
		b.WriteString("\n")
	}

	// Bottom border
	bottom := styles.Yellow.Render("└") + styles.Muted.Render(strings.Repeat("─", cw+2)) + styles.Yellow.Render("┘")
	b.WriteString(center(bottom, width))

	return b.String()
}

// wrapTextForBox wraps text to fit within box content width
func wrapTextForBox(text string, maxWidth int, styles theme.Styles) []string {
	var result []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return result
	}

	var currentLine strings.Builder
	currentLen := 0

	for _, word := range words {
		wordLen := len(word)

		// Word too long - truncate it
		if wordLen > maxWidth {
			if currentLen > 0 {
				result = append(result, styles.Body.Render(currentLine.String()))
				currentLine.Reset()
				currentLen = 0
			}
			result = append(result, styles.Body.Render(word[:maxWidth-3]+"..."))
			continue
		}

		spaceNeeded := wordLen
		if currentLen > 0 {
			spaceNeeded++
		}

		if currentLen+spaceNeeded > maxWidth {
			result = append(result, styles.Body.Render(currentLine.String()))
			currentLine.Reset()
			currentLen = 0
		}

		if currentLen > 0 {
			currentLine.WriteString(" ")
			currentLen++
		}
		currentLine.WriteString(word)
		currentLen += wordLen
	}

	if currentLen > 0 {
		result = append(result, styles.Body.Render(currentLine.String()))
	}

	return result
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

	// "WELCOME TO" text
	welcomeText := styles.Yellow.Render("░▒▓") + styles.Muted.Render(" WELCOME TO ") + styles.Yellow.Render("▓▒░")

	// ASCII banner - scale based on width
	var banner []string
	if width >= 60 {
		banner = []string{
			"███╗   ███╗ ██████╗ ██╗  ██╗ █████╗ ██╗  ██╗",
			"████╗ ████║██╔═══██╗██║  ██║██╔══██╗██║ ██╔╝",
			"██╔████╔██║██║   ██║███████║███████║█████╔╝ ",
			"██║╚██╔╝██║██║   ██║██╔══██║██╔══██║██╔═██╗ ",
			"██║ ╚═╝ ██║╚██████╔╝██║  ██║██║  ██║██║  ██╗",
			"╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝",
		}
	} else {
		// Smaller banner for narrow terminals
		banner = []string{
			"╔╦╗╔═╗╦ ╦╔═╗╦╔═",
			"║║║║ ║╠═╣╠═╣╠╩╗",
			"╩ ╩╚═╝╩ ╩╩ ╩╩ ╩",
		}
	}

	bannerStyles := []lipgloss.Style{
		styles.Yellow,
		styles.Neon,
		styles.Neon,
		styles.Cyan,
		styles.Cyan,
		styles.Yellow,
	}

	b.WriteString("\n\n")
	b.WriteString(center(welcomeText, width))
	b.WriteString("\n\n")

	for i, line := range banner {
		styleIdx := i % len(bannerStyles)
		b.WriteString(center(bannerStyles[styleIdx].Bold(true).Render(line), width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	tagline := styles.Yellow.Render("▓▒░") + styles.Cyan.Render(" FULL STACK · SYSTEMS · AI · DEVOPS ") + styles.Yellow.Render("░▒▓")
	b.WriteString(center(tagline, width))
	b.WriteString("\n\n")

	// Shortcuts box - responsive to width
	bw := boxWidth(width)
	cw := contentWidth(bw)

	var cmdLines []string
	if cw >= 45 {
		cmdLines = []string{
			styles.Green.Bold(true).Render("Alt+A") + styles.Dim.Render(" about") + styles.Yellow.Render(" │ ") + styles.Yellow.Bold(true).Render("Alt+P") + styles.Dim.Render(" projects"),
			styles.Neon.Bold(true).Render("Alt+R") + styles.Dim.Render(" resume") + styles.Yellow.Render("│ ") + styles.Orange.Bold(true).Render("Alt+E") + styles.Dim.Render(" experience"),
			styles.Purple.Bold(true).Render("Alt+H") + styles.Dim.Render(" help") + styles.Yellow.Render("  │ ") + styles.Red.Bold(true).Render("Alt+Q") + styles.Dim.Render(" quit"),
			"",
			styles.Cyan.Render("just type to chat with AI"),
		}
	} else {
		cmdLines = []string{
			styles.Green.Bold(true).Render("Alt+A") + styles.Dim.Render(" about"),
			styles.Yellow.Bold(true).Render("Alt+P") + styles.Dim.Render(" projects"),
			styles.Neon.Bold(true).Render("Alt+R") + styles.Dim.Render(" resume"),
			styles.Purple.Bold(true).Render("Alt+H") + styles.Dim.Render(" help"),
			"",
			styles.Cyan.Render("type to chat"),
		}
	}
	b.WriteString(box("SHORTCUTS", cmdLines, styles, width))
	b.WriteString("\n")

	return b.String()
}

// Help renders help screen
func Help(styles theme.Styles, width int) string {
	var b strings.Builder
	b.WriteString("\n")

	bw := boxWidth(width)
	cw := contentWidth(bw)

	// Adjust content based on available width
	if cw >= 40 {
		shortcuts := []string{
			styles.Yellow.Bold(true).Render("NAVIGATION"),
			"",
			styles.Purple.Bold(true).Render("Alt+H") + styles.Dim.Render(" ") + styles.Muted.Render("help"),
			styles.Green.Bold(true).Render("Alt+A") + styles.Dim.Render(" ") + styles.Muted.Render("about"),
			styles.Yellow.Bold(true).Render("Alt+P") + styles.Dim.Render(" ") + styles.Muted.Render("projects"),
			styles.Orange.Bold(true).Render("Alt+E") + styles.Dim.Render(" ") + styles.Muted.Render("experience"),
			styles.Neon.Bold(true).Render("Alt+R") + styles.Dim.Render(" ") + styles.Muted.Render("resume"),
			styles.Cyan.Bold(true).Render("Alt+W") + styles.Dim.Render(" ") + styles.Muted.Render("home"),
			styles.Cyan.Bold(true).Render("Alt+C") + styles.Dim.Render(" ") + styles.Muted.Render("clear chat"),
			styles.Red.Bold(true).Render("Alt+Q") + styles.Dim.Render(" ") + styles.Muted.Render("quit"),
		}
		b.WriteString(box("ALT+KEY", shortcuts, styles, width))
		b.WriteString("\n")

		commands := []string{
			styles.Yellow.Bold(true).Render("COMMANDS"),
			"",
			styles.Purple.Bold(true).Render("/help") + styles.Muted.Render(" show help"),
			styles.Green.Bold(true).Render("/about") + styles.Muted.Render(" profile"),
			styles.Yellow.Bold(true).Render("/projects") + styles.Muted.Render(" list"),
			styles.Yellow.Bold(true).Render("/open <id>") + styles.Muted.Render(" view"),
			styles.Red.Bold(true).Render("/exit") + styles.Muted.Render(" quit"),
		}
		b.WriteString(box("SLASH", commands, styles, width))
		b.WriteString("\n")
	} else {
		// Compact view for narrow screens
		compact := []string{
			styles.Cyan.Bold(true).Render("Alt+") + styles.Muted.Render(" shortcuts"),
			"A about, P projects",
			"R resume, E exp",
			"H help, Q quit",
			"",
			styles.Cyan.Bold(true).Render("Commands:"),
			"/help /about /exit",
		}
		b.WriteString(box("HELP", compact, styles, width))
		b.WriteString("\n")
	}

	return b.String()
}

// About renders about screen
func About(styles theme.Styles, bio string, width int) string {
	var b strings.Builder
	b.WriteString("\n")

	bw := boxWidth(width)
	cw := contentWidth(bw)

	var lines []string
	bioLines := strings.Split(bio, "\n")

	for _, line := range bioLines {
		if strings.HasPrefix(line, "# ") {
			continue // Skip title
		} else if strings.HasPrefix(line, "## ") {
			title := strings.TrimPrefix(line, "## ")
			lines = append(lines, "")
			lines = append(lines, styles.Cyan.Bold(true).Render("◈ "+title))
		} else if strings.HasPrefix(line, "- **") && strings.Contains(line, "**") {
			parts := strings.SplitN(line, "**", 3)
			if len(parts) >= 3 {
				key := parts[1]
				value := parts[2]
				// Truncate value if too long
				maxVal := cw - len(key) - 6
				if maxVal < 10 {
					maxVal = 10
				}
				if len(value) > maxVal {
					value = value[:maxVal-3] + "..."
				}
				lines = append(lines, styles.Green.Render("▸ ")+styles.Neon.Bold(true).Render(key)+styles.Body.Render(value))
			}
		} else if strings.HasPrefix(line, "- ") {
			text := strings.TrimPrefix(line, "- ")
			text = renderInlineBold(text, styles)
			// Wrap long list items
			if lipgloss.Width(text) > cw-4 {
				text = TruncateText(text, cw-4)
			}
			lines = append(lines, styles.Green.Render("▸ ")+text)
		} else if line != "" {
			// Wrap paragraph text
			wrapped := wrapTextForBox(line, cw-2, styles)
			lines = append(lines, wrapped...)
		}
	}

	b.WriteString(box("PROFILE", lines, styles, width))
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

	bw := boxWidth(width)
	cw := contentWidth(bw)

	var lines []string
	for i, p := range projects.Projects {
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

		// Project header
		header := styles.Dim.Render(fmt.Sprintf("[%d] ", i+1)) +
			styles.Neon.Bold(true).Render(p.Name) + " " +
			statusStyle.Render(statusIcon)
		lines = append(lines, header)

		lines = append(lines, styles.Dim.Render("    ID: ")+styles.Muted.Render(p.ID))

		// Description - truncate to fit
		desc := p.Description
		maxDesc := cw - 6
		if maxDesc < 20 {
			maxDesc = 20
		}
		if len(desc) > maxDesc {
			desc = desc[:maxDesc-3] + "..."
		}
		lines = append(lines, styles.Dim.Render("    ")+styles.Body.Render(desc))

		// Tech tags - limit based on width
		var tags string
		colorCycle := []lipgloss.Style{styles.Cyan, styles.Neon, styles.Green, styles.Yellow}
		maxTags := 3
		if cw < 40 {
			maxTags = 2
		}
		for j, tech := range p.Tech {
			if j < maxTags {
				tags += colorCycle[j%4].Render("⟨"+tech+"⟩") + " "
			}
		}
		lines = append(lines, styles.Dim.Render("    ")+tags)
		lines = append(lines, "")
	}

	sepLen := min(cw-2, 40)
	lines = append(lines, styles.Dim.Render(strings.Repeat("─", sepLen)))
	lines = append(lines, styles.Muted.Render("/open <id> to view details"))

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

	bw := boxWidth(width)
	cw := contentWidth(bw)

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

	// Description - wrap to fit
	lines = append(lines, styles.Cyan.Bold(true).Render("◈ DESCRIPTION"))
	descLines := wrapTextForBox(project.Description, cw-4, styles)
	for _, dl := range descLines {
		lines = append(lines, "  "+dl)
	}
	lines = append(lines, "")

	// Tech
	lines = append(lines, styles.Green.Bold(true).Render("◈ TECH_STACK"))
	var tags string
	colorCycle := []lipgloss.Style{styles.Cyan, styles.Neon, styles.Green, styles.Yellow}
	currentTagLen := 0
	for i, tech := range project.Tech {
		tag := colorCycle[i%4].Render("⟨"+tech+"⟩") + " "
		tagLen := len(tech) + 3
		if currentTagLen+tagLen > cw-4 {
			lines = append(lines, "  "+tags)
			tags = ""
			currentTagLen = 0
		}
		tags += tag
		currentTagLen += tagLen
	}
	if tags != "" {
		lines = append(lines, "  "+tags)
	}
	lines = append(lines, "")

	// Links
	if project.Links.Demo != "" || project.Links.Github != "" {
		lines = append(lines, styles.Yellow.Bold(true).Render("◈ LINKS"))
		if project.Links.Demo != "" {
			demo := project.Links.Demo
			if len(demo) > cw-12 {
				demo = demo[:cw-15] + "..."
			}
			lines = append(lines, styles.Dim.Render("  DEMO:   ")+styles.Link.Render(demo))
		}
		if project.Links.Github != "" {
			gh := project.Links.Github
			if len(gh) > cw-12 {
				gh = gh[:cw-15] + "..."
			}
			lines = append(lines, styles.Dim.Render("  SOURCE: ")+styles.Link.Render(gh))
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

	bw := boxWidth(width)
	cw := contentWidth(bw)

	var lines []string

	// Header
	lines = append(lines, center(styles.Neon.Bold(true).Render(resume.Name), cw))
	lines = append(lines, center(styles.Cyan.Render(resume.Title), cw))
	if resume.Tagline != "" {
		tagline := resume.Tagline
		if len(tagline) > cw-4 {
			tagline = tagline[:cw-7] + "..."
		}
		lines = append(lines, center(styles.Muted.Italic(true).Render("\""+tagline+"\""), cw))
	}
	lines = append(lines, "")

	// Contact
	contact := styles.Green.Render("✉ ") + styles.Body.Render(resume.Contact.Email)
	lines = append(lines, center(contact, cw))
	if resume.Contact.Website != "" {
		web := styles.Cyan.Render("⚡ ") + styles.Link.Render(resume.Contact.Website)
		lines = append(lines, center(web, cw))
	}
	github := styles.Purple.Render("◈ ") + styles.Body.Render(resume.Contact.Github)
	lines = append(lines, center(github, cw))
	lines = append(lines, "")

	sepLen := min(cw-2, 44)
	lines = append(lines, styles.Dim.Render(strings.Repeat("─", sepLen)))
	lines = append(lines, "")

	// Summary - wrap text
	lines = append(lines, styles.Purple.Bold(true).Render("◈ SUMMARY"))
	summaryLines := wrapTextForBox(resume.Summary, cw-4, styles)
	for _, sl := range summaryLines {
		lines = append(lines, "  "+sl)
	}
	lines = append(lines, "")

	// Skills
	lines = append(lines, styles.Cyan.Bold(true).Render("◈ SKILLS"))
	skillLine := func(skills []string, style lipgloss.Style, maxSkills int) string {
		var s string
		currentLen := 0
		for i, skill := range skills {
			if i >= maxSkills {
				break
			}
			tag := style.Render("⟨"+skill+"⟩") + " "
			tagLen := len(skill) + 3
			if currentLen+tagLen > cw-4 {
				break
			}
			s += tag
			currentLen += tagLen
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
		degree := edu.Degree
		if len(degree) > cw-4 {
			degree = degree[:cw-7] + "..."
		}
		lines = append(lines, "  "+styles.Neon.Bold(true).Render(degree))

		inst := edu.Institution + ", " + edu.Location
		if len(inst) > cw-4 {
			inst = inst[:cw-7] + "..."
		}
		lines = append(lines, "  "+styles.Cyan.Render(inst))
		lines = append(lines, "  "+styles.Dim.Render(edu.Period)+" │ "+styles.Green.Render(edu.Score))
		lines = append(lines, "")
	}

	// Achievements
	if len(resume.Achievements) > 0 {
		lines = append(lines, styles.Green.Bold(true).Render("◈ ACHIEVEMENTS"))
		for i, ach := range resume.Achievements {
			if i < 3 {
				a := ach
				maxAch := cw - 6
				if len(a) > maxAch {
					a = a[:maxAch-3] + "..."
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

	bw := boxWidth(width)
	cw := contentWidth(bw)

	var lines []string

	lines = append(lines, center(styles.Neon.Bold(true).Render("WORK EXPERIENCE"), cw))
	lines = append(lines, center(styles.Muted.Render(resume.Name), cw))
	lines = append(lines, "")

	sepLen := min(cw-2, 44)
	lines = append(lines, styles.Dim.Render(strings.Repeat("─", sepLen)))
	lines = append(lines, "")

	for i, exp := range resume.Experience {
		role := exp.Role
		if len(role) > cw-2 {
			role = role[:cw-5] + "..."
		}
		lines = append(lines, styles.Neon.Bold(true).Render(role))

		company := exp.Company
		if len(company) > cw-4 {
			company = company[:cw-7] + "..."
		}
		lines = append(lines, styles.Dim.Render("@ ")+styles.Cyan.Bold(true).Render(company))
		lines = append(lines, styles.Muted.Render("  "+exp.Period))
		lines = append(lines, "")

		// Highlights - truncate each
		for _, h := range exp.Highlights {
			hl := h
			maxHL := cw - 6
			if len(hl) > maxHL {
				hl = hl[:maxHL-3] + "..."
			}
			lines = append(lines, styles.Green.Render("  ▸ ")+styles.Body.Render(hl))
		}

		if i < len(resume.Experience)-1 {
			lines = append(lines, "")
			innerSep := min(cw-6, 36)
			lines = append(lines, styles.Dim.Render("  "+strings.Repeat("─", innerSep)))
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

	// Calculate border width based on screen
	borderLen := min(width-8, 40)
	if borderLen < 20 {
		borderLen = 20
	}

	if role == "user" {
		b.WriteString(styles.Cyan.Bold(true).Render("┌─ YOU " + strings.Repeat("─", borderLen-6)))
		b.WriteString("\n")

		// Wrap user message
		maxMsgWidth := width - 8
		wrapped := WrapText(content, maxMsgWidth)
		for _, line := range strings.Split(wrapped, "\n") {
			b.WriteString(styles.Dim.Render("│ ") + styles.Body.Render(line))
			b.WriteString("\n")
		}

		b.WriteString(styles.Dim.Render("└" + strings.Repeat("─", borderLen)))
	} else {
		b.WriteString(styles.Neon.Bold(true).Render("┌─ MOHAK.AI " + strings.Repeat("─", borderLen-11)))
		b.WriteString("\n")

		// Set markdown renderer width
		mdRenderer.SetWidth(width - 6)
		rendered := mdRenderer.Render(content)
		lines := strings.Split(rendered, "\n")
		for _, line := range lines {
			b.WriteString(styles.Dim.Render("│ ") + line)
			b.WriteString("\n")
		}
		b.WriteString(styles.Dim.Render("└" + strings.Repeat("─", borderLen)))
	}
	b.WriteString("\n")

	return b.String()
}

// StreamingMessage renders streaming AI response
func StreamingMessage(styles theme.Styles, content string, width int, mdRenderer *MarkdownRenderer) string {
	var b strings.Builder

	borderLen := min(width-8, 40)
	if borderLen < 20 {
		borderLen = 20
	}

	b.WriteString(styles.Neon.Bold(true).Render("┌─ MOHAK.AI ") + styles.Neon.Render("▓▒░ streaming ░▒▓"))
	b.WriteString("\n")

	if content != "" {
		mdRenderer.SetWidth(width - 6)
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
	b.WriteString(styles.Dim.Render("└" + strings.Repeat("─", borderLen)))
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
