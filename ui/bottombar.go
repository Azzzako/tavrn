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
		keyStyle.Render("^H") + " " + descStyle.Render("help") + sep +
		keyStyle.Render("^J") + " " + descStyle.Render("rooms") + sep +
		keyStyle.Render("^P") + " " + descStyle.Render("post") + sep +
		keyStyle.Render("^N") + " " + descStyle.Render("nick") + sep +
		keyStyle.Render("^C") + " " + descStyle.Render("exit")

	return BottomBarStyle.Width(b.Width).MaxWidth(b.Width).Render(content)
}
