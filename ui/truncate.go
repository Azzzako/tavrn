package ui

import "charm.land/lipgloss/v2"

// truncateWidth truncates s to fit within maxW display columns,
// appending "..." if truncated. Safe for emoji, CJK, and Unicode.
func truncateWidth(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}
	runes := []rune(s)
	for lipgloss.Width(string(runes)) > maxW-3 && len(runes) > 0 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}
