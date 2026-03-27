package ui

import (
	"charm.land/lipgloss/v2"
)

type BottomBar struct {
	Width int
}

func NewBottomBar() BottomBar {
	return BottomBar{}
}

func (b BottomBar) View() string {
	keyStyle := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(ColorDim)
	sep := lipgloss.NewStyle().Foreground(ColorDimmer).Render("  ·  ")

	content := "  " +
		keyStyle.Render("/help") + sep +
		keyStyle.Render("CTRL+P") + " " + descStyle.Render("post") + sep +
		keyStyle.Render("CTRL+N") + " " + descStyle.Render("nick") + sep +
		keyStyle.Render("CTRL+C") + " " + descStyle.Render("exit")

	return BottomBarStyle.Width(b.Width).MaxWidth(b.Width).Render(content)
}
