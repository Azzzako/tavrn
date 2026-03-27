package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Overlay composites a modal box on top of a base view, dimming the background.
// The base view is kept visible but greyed out behind the modal.
func Overlay(base, modal string, width, height int) string {
	// Split base into lines
	baseLines := strings.Split(base, "\n")

	// Pad base to fill screen
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	// Dim all base lines
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	for i, line := range baseLines {
		// Strip existing styles and dim the text
		stripped := stripAnsi(line)
		if len(stripped) < width {
			stripped += strings.Repeat(" ", width-len(stripped))
		}
		baseLines[i] = dimStyle.Render(stripped)
	}

	// Get modal dimensions
	modalLines := strings.Split(modal, "\n")
	modalHeight := len(modalLines)
	modalWidth := 0
	for _, line := range modalLines {
		w := lipgloss.Width(line)
		if w > modalWidth {
			modalWidth = w
		}
	}

	// Center modal position
	startY := (height - modalHeight) / 2
	startX := (width - modalWidth) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Composite modal onto dimmed base
	for i, mLine := range modalLines {
		row := startY + i
		if row >= len(baseLines) {
			break
		}

		baseLine := baseLines[row]
		baseRunes := []rune(stripAnsi(baseLine))

		// Build the composited line: left padding + modal + right rest
		var b strings.Builder

		// Left portion (dimmed base)
		if startX > 0 {
			leftPart := string(baseRunes[:min(startX, len(baseRunes))])
			b.WriteString(dimStyle.Render(leftPart))
		}

		// Modal content
		b.WriteString(mLine)

		// Right portion
		mw := lipgloss.Width(mLine)
		rightStart := startX + mw
		if rightStart < len(baseRunes) {
			rightPart := string(baseRunes[min(rightStart, len(baseRunes)):])
			b.WriteString(dimStyle.Render(rightPart))
		}

		baseLines[row] = b.String()
	}

	return strings.Join(baseLines[:height], "\n")
}

// stripAnsi removes ANSI escape codes from a string.
func stripAnsi(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
