package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
)

// MarkdownRenderer renders markdown text with theme styles
type MarkdownRenderer struct {
	styles   theme.Styles
	maxWidth int
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(styles theme.Styles) *MarkdownRenderer {
	return &MarkdownRenderer{styles: styles, maxWidth: 80}
}

// NewMarkdownRendererWithWidth creates a renderer with specific width
func NewMarkdownRendererWithWidth(styles theme.Styles, width int) *MarkdownRenderer {
	if width < 20 {
		width = 80
	}
	return &MarkdownRenderer{styles: styles, maxWidth: width}
}

// SetWidth updates the max width for rendering
func (r *MarkdownRenderer) SetWidth(width int) {
	if width >= 20 {
		r.maxWidth = width
	}
}

// Render converts markdown text to styled terminal output
func (r *MarkdownRenderer) Render(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	inCodeBlock := false
	codeBlockLang := ""

	// Calculate content width (leave room for borders/prefix)
	contentWidth := r.maxWidth - 4

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Code block handling
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(line, "```")
				borderLen := min(contentWidth-4, 40)
				result.WriteString(r.styles.Dim.Render("┌─"))
				if codeBlockLang != "" {
					result.WriteString(r.styles.Cyan.Render(" " + codeBlockLang + " "))
					borderLen -= len(codeBlockLang) + 2
				}
				result.WriteString(r.styles.Dim.Render(strings.Repeat("─", max(borderLen, 10))))
				result.WriteString("\n")
			} else {
				inCodeBlock = false
				codeBlockLang = ""
				borderLen := min(contentWidth, 44)
				result.WriteString(r.styles.Dim.Render("└" + strings.Repeat("─", borderLen)))
				result.WriteString("\n")
			}
			i++
			continue
		}

		if inCodeBlock {
			// Code blocks: truncate if too long, don't wrap
			codeLine := line
			if len(codeLine) > contentWidth-4 {
				codeLine = codeLine[:contentWidth-7] + "..."
			}
			result.WriteString(r.styles.Dim.Render("│ "))
			result.WriteString(r.styles.Green.Render(codeLine))
			result.WriteString("\n")
			i++
			continue
		}

		// Check for table
		if r.isTableRow(line) && i+1 < len(lines) && r.isTableSeparator(lines[i+1]) {
			tableLines := []string{line}
			j := i + 1
			for j < len(lines) && (r.isTableRow(lines[j]) || r.isTableSeparator(lines[j])) {
				tableLines = append(tableLines, lines[j])
				j++
			}
			result.WriteString(r.renderTable(tableLines, contentWidth))
			result.WriteString("\n")
			i = j
			continue
		}

		// Process regular line with wrapping
		rendered := r.renderLine(line, contentWidth)
		result.WriteString(rendered)
		result.WriteString("\n")
		i++
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func (r *MarkdownRenderer) isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
}

func (r *MarkdownRenderer) isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	cleaned := strings.ReplaceAll(trimmed, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return len(cleaned) == 0
}

func (r *MarkdownRenderer) parseTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := strings.Split(trimmed, "|")
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}
	return cells
}

func (r *MarkdownRenderer) renderTable(lines []string, maxWidth int) string {
	if len(lines) < 2 {
		return ""
	}

	header := r.parseTableRow(lines[0])
	numCols := len(header)

	var dataRows [][]string
	for i := 2; i < len(lines); i++ {
		if !r.isTableSeparator(lines[i]) {
			row := r.parseTableRow(lines[i])
			for len(row) < numCols {
				row = append(row, "")
			}
			dataRows = append(dataRows, row[:numCols])
		}
	}

	// Calculate column widths, respecting maxWidth
	colWidths := make([]int, numCols)
	for i, h := range header {
		if len(h) > colWidths[i] {
			colWidths[i] = len(h)
		}
	}
	for _, row := range dataRows {
		for i, cell := range row {
			if i < numCols && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add padding and enforce max column width
	totalWidth := numCols + 1 // borders
	maxColWidth := (maxWidth - totalWidth) / numCols
	if maxColWidth < 8 {
		maxColWidth = 8
	}

	for i := range colWidths {
		colWidths[i] += 2
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
		if colWidths[i] < 5 {
			colWidths[i] = 5
		}
	}

	var result strings.Builder

	// Top border
	result.WriteString(r.styles.Cyan.Render("┌"))
	for i, w := range colWidths {
		result.WriteString(r.styles.Dim.Render(strings.Repeat("─", w)))
		if i < numCols-1 {
			result.WriteString(r.styles.Cyan.Render("┬"))
		}
	}
	result.WriteString(r.styles.Cyan.Render("┐"))
	result.WriteString("\n")

	// Header row
	result.WriteString(r.styles.Cyan.Render("│"))
	for i, h := range header {
		cell := r.padCenter(r.truncateCell(h, colWidths[i]-2), colWidths[i])
		result.WriteString(r.styles.Neon.Bold(true).Render(cell))
		if i < numCols-1 {
			result.WriteString(r.styles.Cyan.Render("│"))
		}
	}
	result.WriteString(r.styles.Cyan.Render("│"))
	result.WriteString("\n")

	// Header separator
	result.WriteString(r.styles.Cyan.Render("├"))
	for i, w := range colWidths {
		result.WriteString(r.styles.Dim.Render(strings.Repeat("─", w)))
		if i < numCols-1 {
			result.WriteString(r.styles.Cyan.Render("┼"))
		}
	}
	result.WriteString(r.styles.Cyan.Render("┤"))
	result.WriteString("\n")

	// Data rows
	colorCycle := []lipgloss.Style{r.styles.Body, r.styles.Muted}
	for rowIdx, row := range dataRows {
		result.WriteString(r.styles.Dim.Render("│"))
		rowStyle := colorCycle[rowIdx%2]
		for i, cell := range row {
			paddedCell := r.padCenter(r.truncateCell(cell, colWidths[i]-2), colWidths[i])
			result.WriteString(rowStyle.Render(paddedCell))
			if i < numCols-1 {
				result.WriteString(r.styles.Dim.Render("│"))
			}
		}
		result.WriteString(r.styles.Dim.Render("│"))
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(r.styles.Cyan.Render("└"))
	for i, w := range colWidths {
		result.WriteString(r.styles.Dim.Render(strings.Repeat("─", w)))
		if i < numCols-1 {
			result.WriteString(r.styles.Cyan.Render("┴"))
		}
	}
	result.WriteString(r.styles.Cyan.Render("┘"))

	return result.String()
}

func (r *MarkdownRenderer) truncateCell(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return "..."
	}
	return text[:maxLen-3] + "..."
}

func (r *MarkdownRenderer) padCenter(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}
	leftPad := (width - textLen) / 2
	rightPad := width - textLen - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

func (r *MarkdownRenderer) renderLine(line string, maxWidth int) string {
	// Headers - don't wrap, truncate if needed
	if strings.HasPrefix(line, "#### ") {
		text := strings.TrimPrefix(line, "#### ")
		if len(text) > maxWidth-4 {
			text = text[:maxWidth-7] + "..."
		}
		return r.styles.Yellow.Render("▸ ") + r.styles.Yellow.Render(text)
	}
	if strings.HasPrefix(line, "### ") {
		text := strings.TrimPrefix(line, "### ")
		if len(text) > maxWidth-4 {
			text = text[:maxWidth-7] + "..."
		}
		return r.styles.Cyan.Render("◆ ") + r.styles.Cyan.Bold(true).Render(text)
	}
	if strings.HasPrefix(line, "## ") {
		text := strings.TrimPrefix(line, "## ")
		if len(text) > maxWidth-4 {
			text = text[:maxWidth-7] + "..."
		}
		return r.styles.Neon.Render("◈ ") + r.styles.Neon.Bold(true).Render(text)
	}
	if strings.HasPrefix(line, "# ") {
		text := strings.TrimPrefix(line, "# ")
		headerWidth := maxWidth - 8
		if len(text) > headerWidth {
			text = text[:headerWidth-3] + "..."
		}
		return r.styles.Neon.Bold(true).Render("═══ " + text + " ═══")
	}

	// Blockquote
	if strings.HasPrefix(line, "> ") {
		text := strings.TrimPrefix(line, "> ")
		wrapped := r.wrapText(text, maxWidth-4)
		lines := strings.Split(wrapped, "\n")
		var result strings.Builder
		for i, l := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(r.styles.Dim.Render("┃ ") + r.styles.Muted.Italic(true).Render(l))
		}
		return result.String()
	}

	// Horizontal rule
	if line == "---" || line == "***" || line == "___" {
		ruleLen := min(maxWidth, 44)
		return r.styles.Dim.Render(strings.Repeat("─", ruleLen))
	}

	// Unordered list items - wrap with indent
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		text := line[2:]
		text = r.renderInline(text)
		wrapped := r.wrapText(text, maxWidth-6)
		lines := strings.Split(wrapped, "\n")
		var result strings.Builder
		for i, l := range lines {
			if i == 0 {
				result.WriteString(r.styles.Green.Render("  ▹ ") + l)
			} else {
				result.WriteString("\n" + r.styles.Dim.Render("    ") + l)
			}
		}
		return result.String()
	}

	// Nested list items
	if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  * ") {
		text := line[4:]
		text = r.renderInline(text)
		wrapped := r.wrapText(text, maxWidth-8)
		lines := strings.Split(wrapped, "\n")
		var result strings.Builder
		for i, l := range lines {
			if i == 0 {
				result.WriteString(r.styles.Cyan.Render("    ◦ ") + l)
			} else {
				result.WriteString("\n" + r.styles.Dim.Render("      ") + l)
			}
		}
		return result.String()
	}

	// Ordered list items
	if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
		parts := strings.SplitN(line, ". ", 2)
		if len(parts) == 2 {
			num := parts[0]
			text := r.renderInline(parts[1])
			wrapped := r.wrapText(text, maxWidth-6)
			lines := strings.Split(wrapped, "\n")
			var result strings.Builder
			for i, l := range lines {
				if i == 0 {
					result.WriteString(r.styles.Yellow.Render("  "+num+". ") + l)
				} else {
					result.WriteString("\n" + r.styles.Dim.Render("     ") + l)
				}
			}
			return result.String()
		}
	}

	// Regular paragraph - wrap and apply inline formatting
	text := r.renderInline(line)
	return r.wrapText(text, maxWidth)
}

// wrapText wraps plain text to maxWidth
func (r *MarkdownRenderer) wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 60
	}

	// For styled text, use visible width
	visibleWidth := lipgloss.Width(text)
	if visibleWidth <= maxWidth {
		return text
	}

	// If it contains ANSI codes, try simpler approach
	if strings.Contains(text, "\x1b[") {
		return r.wrapStyledText(text, maxWidth)
	}

	// Plain text word wrapping
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	currentLen := 0

	for i, word := range words {
		wordLen := len(word)

		// Word too long - break it
		if wordLen > maxWidth {
			if currentLen > 0 {
				result.WriteString("\n")
				currentLen = 0
			}
			for j := 0; j < len(word); j += maxWidth - 1 {
				end := j + maxWidth - 1
				if end > len(word) {
					end = len(word)
				}
				if j > 0 {
					result.WriteString("\n")
				}
				result.WriteString(word[j:end])
			}
			currentLen = len(word) % (maxWidth - 1)
			continue
		}

		spaceNeeded := wordLen
		if i > 0 && currentLen > 0 {
			spaceNeeded++
		}

		if currentLen+spaceNeeded > maxWidth {
			result.WriteString("\n")
			currentLen = 0
		} else if currentLen > 0 {
			result.WriteString(" ")
			currentLen++
		}

		result.WriteString(word)
		currentLen += wordLen
	}

	return result.String()
}

// wrapStyledText attempts to wrap text with ANSI codes
func (r *MarkdownRenderer) wrapStyledText(text string, maxWidth int) string {
	// Simplified: just return as-is if it's styled
	// The caller should handle overflow via viewport scrolling
	return text
}

func (r *MarkdownRenderer) renderInline(text string) string {
	text = r.processInlineCode(text)
	text = r.processBold(text)
	text = r.processItalic(text)
	text = r.processLinks(text)
	return text
}

func (r *MarkdownRenderer) processInlineCode(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.Trim(match, "`")
		return r.styles.Cyan.Render("⟨") + r.styles.Green.Render(code) + r.styles.Cyan.Render("⟩")
	})
}

func (r *MarkdownRenderer) processBold(text string) string {
	re := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.Trim(match, "*_")
		return r.styles.Neon.Bold(true).Render(inner)
	})
}

func (r *MarkdownRenderer) processItalic(text string) string {
	re := regexp.MustCompile(`\*([^*]+)\*`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.Trim(match, "*")
		if inner == "" {
			return match
		}
		return r.styles.Muted.Italic(true).Render(inner)
	})
}

func (r *MarkdownRenderer) processLinks(text string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		re2 := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
		matches := re2.FindStringSubmatch(match)
		if len(matches) == 3 {
			linkText := matches[1]
			url := matches[2]
			return r.styles.Blue.Underline(true).Render(linkText) + r.styles.Dim.Render(" ("+url+")")
		}
		return match
	})
}

// RenderStreaming renders partial markdown (for streaming)
func (r *MarkdownRenderer) RenderStreaming(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	inCodeBlock := false
	contentWidth := r.maxWidth - 4

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				lang := strings.TrimPrefix(line, "```")
				borderLen := min(contentWidth-4, 30)
				result.WriteString(r.styles.Dim.Render("┌─"))
				if lang != "" {
					result.WriteString(r.styles.Cyan.Render(" " + lang + " "))
					borderLen -= len(lang) + 2
				}
				result.WriteString(r.styles.Dim.Render(strings.Repeat("─", max(borderLen, 5))))
			} else {
				result.WriteString(r.styles.Dim.Render("└" + strings.Repeat("─", min(contentWidth, 34))))
			}
			result.WriteString("\n")
			continue
		}

		if inCodeBlock {
			codeLine := line
			if len(codeLine) > contentWidth-4 {
				codeLine = codeLine[:contentWidth-7] + "..."
			}
			result.WriteString(r.styles.Dim.Render("│ "))
			result.WriteString(r.styles.Green.Render(codeLine))
			result.WriteString("\n")
			continue
		}

		rendered := r.renderLine(line, contentWidth)
		result.WriteString(rendered)
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}
