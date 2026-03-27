package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type RoomInfo struct {
	Name  string
	Count int
}

type Sidebar struct {
	Rooms       []RoomInfo
	OnlineUsers []string
	Width       int
	Height      int
}

func NewSidebar() Sidebar {
	return Sidebar{
		Rooms: []RoomInfo{{Name: "lounge", Count: 0}},
	}
}

func (s Sidebar) View() string {
	sectionHeader := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		MarginBottom(1)

	dimText := lipgloss.NewStyle().Foreground(ColorDim)

	// NOW ONLINE section
	content := sectionHeader.Render("NOW ONLINE") + "\n"
	if len(s.OnlineUsers) == 0 {
		content += dimText.Render("  (empty)") + "\n"
	} else {
		for _, u := range s.OnlineUsers {
			bullet := lipgloss.NewStyle().Foreground(lipgloss.Color("108")).Render("*")
			content += fmt.Sprintf("  %s %s\n", bullet, u)
		}
	}

	content += "\n"

	// ROOMS section
	content += sectionHeader.Render("ROOMS") + "\n"
	for _, r := range s.Rooms {
		count := lipgloss.NewStyle().Foreground(ColorDim).Render(fmt.Sprintf("%d", r.Count))
		name := lipgloss.NewStyle().Foreground(ColorSand).Render(fmt.Sprintf("#%s", r.Name))
		content += fmt.Sprintf("  %s  %s\n", name, count)
	}

	content += "\n"

	// UP NEXT section
	content += sectionHeader.Render("UP NEXT") + "\n"
	content += dimText.Render("  (coming soon)") + "\n"

	return SidebarStyle.
		Width(s.Width).
		Height(s.Height).
		MaxHeight(s.Height).
		Padding(1, 1).
		Render(content)
}
