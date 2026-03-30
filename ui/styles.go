package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Cantina palette — ANSI 256 colors.
var (
	ColorBackground = lipgloss.Color("235")
	ColorDarkBg     = lipgloss.Color("233")
	ColorPanelBg    = lipgloss.Color("234")
	ColorSand       = lipgloss.Color("180")
	ColorDim        = lipgloss.Color("243")
	ColorDimmer     = lipgloss.Color("239")
	ColorBorder     = lipgloss.Color("94")
	ColorHighlight  = lipgloss.Color("179")
	ColorAmber      = lipgloss.Color("172")
	ColorTitle      = lipgloss.Color("180")
	ColorCommand    = lipgloss.Color("179")
	ColorDesc       = lipgloss.Color("243")
	ColorAccent     = lipgloss.Color("137")
	ColorGreen      = lipgloss.Color("108")
	ColorTyping     = lipgloss.Color("109")
	ColorMention    = lipgloss.Color("179") // gold highlight for @mentions

	// 12 muted cantina tones for nicknames
	NickColors = []color.Color{
		lipgloss.Color("174"), // dusty rose
		lipgloss.Color("109"), // faded teal
		lipgloss.Color("137"), // aged copper
		lipgloss.Color("138"), // soft clay
		lipgloss.Color("108"), // pale sage
		lipgloss.Color("179"), // weathered gold
		lipgloss.Color("140"), // dim lavender
		lipgloss.Color("67"),  // smoky blue
		lipgloss.Color("131"), // muted coral
		lipgloss.Color("144"), // warm stone
		lipgloss.Color("136"), // quiet amber
		lipgloss.Color("97"),  // dusk violet
	}
)

var (
	TopBarBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.Border{Bottom: "─"}, false, false, true, false).
				BorderForeground(ColorBorder)

	BottomBarStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Border(lipgloss.Border{Top: "─"}, true, false, false, false).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ChatBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Foreground(ColorSand)

	// Left sidebar: border on right side
	LeftSidebarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, true, false, false).
				BorderForeground(ColorBorder).
				Foreground(ColorSand).
				PaddingTop(1).
				PaddingLeft(1).
				PaddingRight(1)

	// Right sidebar: border on left side
	RightSidebarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(ColorBorder).
				Foreground(ColorSand).
				PaddingTop(1).
				PaddingLeft(1).
				PaddingRight(1)

	SystemMsgStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorSand)

	// Splash screen styles
	SplashBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Foreground(ColorSand).
				Padding(1, 3)

	SplashTitleStyle = lipgloss.NewStyle().
				Foreground(ColorTitle).
				Bold(true)

	SplashSubtitleStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	SplashKeyStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	SplashDescStyle = lipgloss.NewStyle().
			Foreground(ColorDesc)

	SplashCategoryStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true).
				MarginTop(1)

	SplashCommandStyle = lipgloss.NewStyle().
				Foreground(ColorCommand).
				Bold(true)

	// Chat message styles
	MsgTimeStyle = lipgloss.NewStyle().
			Foreground(ColorDimmer)

	TypingStyle = lipgloss.NewStyle().
			Foreground(ColorTyping).
			Italic(true)

	// Mention autocomplete popup styles
	MentionPopupStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	MentionSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true)

	MentionItemStyle = lipgloss.NewStyle().
				Foreground(ColorSand)

	MentionSelfStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true)

	MentionHighlightStyle = lipgloss.NewStyle().
				Foreground(ColorMention).
				Bold(true)
)

func NickStyle(colorIndex int) lipgloss.Style {
	idx := colorIndex % len(NickColors)
	return lipgloss.NewStyle().Foreground(NickColors[idx]).Bold(true)
}

func NickBarColor(colorIndex int) color.Color {
	return NickColors[colorIndex%len(NickColors)]
}
