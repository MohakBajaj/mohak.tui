package ui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// WrapText wraps text to fit within maxWidth, preserving words when possible
func WrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Handle empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		wrapped := wrapLine(line, maxWidth)
		result.WriteString(wrapped)
	}

	return result.String()
}

// wrapLine wraps a single line of text
func wrapLine(line string, maxWidth int) string {
	// Get visible width (ignoring ANSI codes)
	visibleWidth := lipgloss.Width(line)
	if visibleWidth <= maxWidth {
		return line
	}

	// For styled text, we need to be careful
	// If the line has ANSI codes, do character-level wrapping
	if strings.Contains(line, "\x1b[") {
		return wrapStyledLine(line, maxWidth)
	}

	// Simple word wrapping for plain text
	return wrapPlainLine(line, maxWidth)
}

// wrapPlainLine wraps plain text at word boundaries
func wrapPlainLine(line string, maxWidth int) string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	currentLineLen := 0

	for i, word := range words {
		wordLen := len(word)

		// If single word is longer than maxWidth, break it
		if wordLen > maxWidth {
			if currentLineLen > 0 {
				result.WriteString("\n")
				currentLineLen = 0
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
				if end < len(word) {
					result.WriteString("-")
				}
			}
			currentLineLen = len(word) % (maxWidth - 1)
			continue
		}

		// Check if word fits on current line
		spaceNeeded := wordLen
		if i > 0 && currentLineLen > 0 {
			spaceNeeded++ // for space
		}

		if currentLineLen+spaceNeeded > maxWidth {
			result.WriteString("\n")
			currentLineLen = 0
		} else if currentLineLen > 0 {
			result.WriteString(" ")
			currentLineLen++
		}

		result.WriteString(word)
		currentLineLen += wordLen
	}

	return result.String()
}

// wrapStyledLine wraps text that contains ANSI escape codes
func wrapStyledLine(line string, maxWidth int) string {
	var result strings.Builder
	var currentLineWidth int
	var inEscape bool
	var escapeSeq strings.Builder
	var activeStyles []string // Track active ANSI sequences

	runes := []rune(line)
	lastBreakPoint := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Track ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}

		if inEscape {
			escapeSeq.WriteRune(r)
			if r == 'm' {
				inEscape = false
				seq := escapeSeq.String()
				if seq == "\x1b[0m" {
					activeStyles = nil
				} else {
					activeStyles = append(activeStyles, seq)
				}
				result.WriteString(seq)
			}
			continue
		}

		// Track word boundaries
		if unicode.IsSpace(r) {
			lastBreakPoint = result.Len()
		}

		// Check if we need to wrap
		if currentLineWidth >= maxWidth {
			// Try to break at last word boundary
			if lastBreakPoint > 0 && lastBreakPoint < result.Len() {
				// This is complex with ANSI - for now just break here
				result.WriteString("\n")
				// Re-apply active styles
				for _, style := range activeStyles {
					result.WriteString(style)
				}
				currentLineWidth = 0
			} else {
				result.WriteString("\n")
				for _, style := range activeStyles {
					result.WriteString(style)
				}
				currentLineWidth = 0
			}
		}

		result.WriteRune(r)
		currentLineWidth++
	}

	return result.String()
}

// WrapTextWithPrefix wraps text and adds a prefix to continuation lines
func WrapTextWithPrefix(text string, maxWidth int, firstPrefix, contPrefix string) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	firstPrefixWidth := lipgloss.Width(firstPrefix)
	contPrefixWidth := lipgloss.Width(contPrefix)

	// Adjust maxWidth for prefixes
	firstLineWidth := maxWidth - firstPrefixWidth
	contLineWidth := maxWidth - contPrefixWidth

	if firstLineWidth < 10 {
		firstLineWidth = 10
	}
	if contLineWidth < 10 {
		contLineWidth = 10
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		if strings.TrimSpace(line) == "" {
			result.WriteString(firstPrefix)
			continue
		}

		words := strings.Fields(line)
		currentLineLen := 0
		isFirstLine := true

		for j, word := range words {
			wordLen := lipgloss.Width(word)
			currentMaxWidth := firstLineWidth
			if !isFirstLine {
				currentMaxWidth = contLineWidth
			}

			// Check if word fits
			spaceNeeded := wordLen
			if j > 0 && currentLineLen > 0 {
				spaceNeeded++
			}

			if currentLineLen+spaceNeeded > currentMaxWidth && currentLineLen > 0 {
				result.WriteString("\n")
				result.WriteString(contPrefix)
				currentLineLen = 0
				isFirstLine = false
			} else if currentLineLen == 0 {
				if isFirstLine {
					result.WriteString(firstPrefix)
				}
			} else {
				result.WriteString(" ")
				currentLineLen++
			}

			result.WriteString(word)
			currentLineLen += wordLen
		}
	}

	return result.String()
}

// TruncateText truncates text to maxWidth with ellipsis
func TruncateText(text string, maxWidth int) string {
	if maxWidth <= 3 {
		return "..."
	}

	visibleWidth := lipgloss.Width(text)
	if visibleWidth <= maxWidth {
		return text
	}

	// For plain text, simple truncation
	if !strings.Contains(text, "\x1b[") {
		if len(text) > maxWidth-3 {
			return text[:maxWidth-3] + "..."
		}
		return text
	}

	// For styled text, we need to count visible characters
	var result strings.Builder
	var visibleCount int
	var inEscape bool

	for _, r := range text {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		if visibleCount >= maxWidth-3 {
			result.WriteString("...")
			result.WriteString("\x1b[0m") // Reset styles
			break
		}

		result.WriteRune(r)
		visibleCount++
	}

	return result.String()
}
