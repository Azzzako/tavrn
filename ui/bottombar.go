package ui

import (
	"charm.land/lipgloss/v2"
)

type BottomBar struct {
	Width     int
	IsGallery bool
}

func NewBottomBar() BottomBar {
	return BottomBar{}
}

func (b BottomBar) View() string {
	k := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)
	d := lipgloss.NewStyle().Foreground(ColorDim)
	sep := lipgloss.NewStyle().Foreground(ColorDimmer).Render("  ·  ")

	var content string
	if b.IsGallery {
		content = "  " +
			k.Render("P") + " " + d.Render("post") + sep +
			k.Render("J") + " " + d.Render("rooms") + sep +
			k.Render("D") + " " + d.Render("delete") + sep +
			k.Render("TAB") + " " + d.Render("select") + sep +
			k.Render("H") + " " + d.Render("help")
	} else {
		content = "  " +
			k.Render("F1") + " " + d.Render("help") + sep +
			k.Render("F3") + " " + d.Render("rooms") + sep +
			k.Render("SHIFT+↑↓") + " " + d.Render("scroll")
	}

	return BottomBarStyle.Width(b.Width).MaxWidth(b.Width).Render(content)
}
