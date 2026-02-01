package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
)

// MarkdownRenderer renders markdown text with theme styles
type MarkdownRenderer struct {
	styles theme.Styles
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(styles theme.Styles) *MarkdownRenderer {
	return &MarkdownRenderer{styles: styles}
}

// Render converts markdown text to styled terminal output
func (r *MarkdownRenderer) Render(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	inCodeBlock := false
	codeBlockLang := ""

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Code block handling
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(line, "```")
				result.WriteString(r.styles.Dim.Render("┌─"))
				if codeBlockLang != "" {
					result.WriteString(r.styles.Cyan.Render(" " + codeBlockLang + " "))
				}
				result.WriteString(r.styles.Dim.Render("─────────────────────────────────"))
				result.WriteString("\n")
			} else {
				inCodeBlock = false
				codeBlockLang = ""
				result.WriteString(r.styles.Dim.Render("└─────────────────────────────────────────"))
				result.WriteString("\n")
			}
			i++
			continue
		}

		if inCodeBlock {
			result.WriteString(r.styles.Dim.Render("│ "))
			result.WriteString(r.styles.Green.Render(line))
			result.WriteString("\n")
			i++
			continue
		}

		// Check for table (line starts with |)
		if r.isTableRow(line) && i+1 < len(lines) && r.isTableSeparator(lines[i+1]) {
			tableLines := []string{line}
			j := i + 1
			for j < len(lines) && (r.isTableRow(lines[j]) || r.isTableSeparator(lines[j])) {
				tableLines = append(tableLines, lines[j])
				j++
			}
			result.WriteString(r.renderTable(tableLines))
			result.WriteString("\n")
			i = j
			continue
		}

		// Process regular line
		rendered := r.renderLine(line)
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
	// Check if it contains mostly dashes and pipes
	cleaned := strings.ReplaceAll(trimmed, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return len(cleaned) == 0
}

func (r *MarkdownRenderer) parseTableRow(line string) []string {
	// Remove leading/trailing pipes and split
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := strings.Split(trimmed, "|")
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}
	return cells
}

func (r *MarkdownRenderer) renderTable(lines []string) string {
	if len(lines) < 2 {
		return ""
	}

	// Parse header
	header := r.parseTableRow(lines[0])
	numCols := len(header)

	// Parse data rows (skip separator at index 1)
	var dataRows [][]string
	for i := 2; i < len(lines); i++ {
		if !r.isTableSeparator(lines[i]) {
			row := r.parseTableRow(lines[i])
			// Pad row to match header columns
			for len(row) < numCols {
				row = append(row, "")
			}
			dataRows = append(dataRows, row[:numCols])
		}
	}

	// Calculate column widths
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

	// Add padding
	for i := range colWidths {
		colWidths[i] += 2
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
		cell := r.padCenter(h, colWidths[i])
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
			paddedCell := r.padCenter(cell, colWidths[i])
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

func (r *MarkdownRenderer) padCenter(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}
	leftPad := (width - textLen) / 2
	rightPad := width - textLen - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

func (r *MarkdownRenderer) renderLine(line string) string {
	// Headers
	if strings.HasPrefix(line, "#### ") {
		text := strings.TrimPrefix(line, "#### ")
		return r.styles.Yellow.Render("▸ ") + r.styles.Yellow.Render(text)
	}
	if strings.HasPrefix(line, "### ") {
		text := strings.TrimPrefix(line, "### ")
		return r.styles.Cyan.Render("◆ ") + r.styles.Cyan.Bold(true).Render(text)
	}
	if strings.HasPrefix(line, "## ") {
		text := strings.TrimPrefix(line, "## ")
		return r.styles.Neon.Render("◈ ") + r.styles.Neon.Bold(true).Render(text)
	}
	if strings.HasPrefix(line, "# ") {
		text := strings.TrimPrefix(line, "# ")
		return r.styles.Neon.Bold(true).Render("═══ " + text + " ═══")
	}

	// Blockquote
	if strings.HasPrefix(line, "> ") {
		text := strings.TrimPrefix(line, "> ")
		return r.styles.Dim.Render("┃ ") + r.styles.Muted.Italic(true).Render(text)
	}

	// Horizontal rule
	if line == "---" || line == "***" || line == "___" {
		return r.styles.Dim.Render("─────────────────────────────────────────")
	}

	// Unordered list items
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		text := line[2:]
		text = r.renderInline(text)
		return r.styles.Green.Render("  ▹ ") + text
	}

	// Nested list items
	if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  * ") {
		text := line[4:]
		text = r.renderInline(text)
		return r.styles.Cyan.Render("    ◦ ") + text
	}

	// Ordered list items
	if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
		parts := strings.SplitN(line, ". ", 2)
		if len(parts) == 2 {
			num := parts[0]
			text := r.renderInline(parts[1])
			return r.styles.Yellow.Render("  "+num+". ") + text
		}
	}

	// Regular line with inline formatting
	return r.renderInline(line)
}

func (r *MarkdownRenderer) renderInline(text string) string {
	// Process inline code first (to avoid conflicts with other formatting)
	text = r.processInlineCode(text)

	// Process bold
	text = r.processBold(text)

	// Process italic
	text = r.processItalic(text)

	// Process links
	text = r.processLinks(text)

	return text
}

func (r *MarkdownRenderer) processInlineCode(text string) string {
	// Match `code`
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.Trim(match, "`")
		return r.styles.Cyan.Render("⟨") + r.styles.Green.Render(code) + r.styles.Cyan.Render("⟩")
	})
}

func (r *MarkdownRenderer) processBold(text string) string {
	// Match **bold** or __bold__
	re := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.Trim(match, "*_")
		return r.styles.Neon.Bold(true).Render(inner)
	})
}

func (r *MarkdownRenderer) processItalic(text string) string {
	// Match *italic* - by this point bold (**) has already been processed
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
	// Match [text](url)
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

// RenderStreaming renders a partial markdown string (for streaming)
// This is a simpler version that doesn't break on incomplete blocks
func (r *MarkdownRenderer) RenderStreaming(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				lang := strings.TrimPrefix(line, "```")
				result.WriteString(r.styles.Dim.Render("┌─"))
				if lang != "" {
					result.WriteString(r.styles.Cyan.Render(" " + lang + " "))
				}
				result.WriteString(r.styles.Dim.Render("───────────────────────"))
			} else {
				result.WriteString(r.styles.Dim.Render("└───────────────────────────────"))
			}
			result.WriteString("\n")
			continue
		}

		if inCodeBlock {
			result.WriteString(r.styles.Dim.Render("│ "))
			result.WriteString(r.styles.Green.Render(line))
			result.WriteString("\n")
			continue
		}

		// Simple inline rendering for streaming (no table support during stream)
		rendered := r.renderLine(line)
		result.WriteString(rendered)
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}
