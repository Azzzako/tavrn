package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type TopBar struct {
	Room        string
	OnlineCount int
	WeeklyCount int
	NowPlaying  string
	Width       int
}

func NewTopBar() TopBar {
	return TopBar{Room: "lounge"}
}

func (t TopBar) View() string {
	// /// TAVRN /// style header with gradient
	diag := lipgloss.NewStyle().Foreground(ColorBorder).Render("///")
	title := GradientText(" TAVRN.SH ", ColorHighlight, ColorAccent, true)
	header := fmt.Sprintf(" %s%s%s", diag, title, diag)

	// Stats section
	onlineBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("108")).Bold(true).Render(
		fmt.Sprintf("[ %02d online ]", t.OnlineCount),
	)
	weeklyBadge := lipgloss.NewStyle().Foreground(ColorDim).Render(
		fmt.Sprintf("[ %d this week ]", t.WeeklyCount),
	)
	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(" * ")
	room := lipgloss.NewStyle().Foreground(ColorAmber).Render(fmt.Sprintf("#%s", t.Room))

	stats := onlineBadge + "  " + weeklyBadge + sep + room

	// Now playing (right aligned)
	right := ""
	if t.NowPlaying != "" {
		note := lipgloss.NewStyle().Foreground(ColorAmber).Render("~")
		right = fmt.Sprintf(" %s %s ", note, t.NowPlaying)
	}

	// Compose: header  stats  [now playing]
	left := header + "  " + stats
	gap := t.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}

	content := left + padding + right
	return TopBarStyle.Width(t.Width).MaxWidth(t.Width).Render(content)
}
