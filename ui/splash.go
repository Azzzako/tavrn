package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const tavernArt = `       _____
      /     \
     | () () |
      \ ___ /
    __|_____|__
   /  |     |  \
  |   | TAV |   |
  |   | ERN |   |
  |   |_____|   |
  |  /       \  |
  |_/  [] []  \_|
    |  []  [] |
    |_________|
    |_|_|_|_|_|`

type Splash struct {
	nickname    string
	fingerprint string
	flair       bool
	width       int
	height      int
}

func NewSplash(nickname, fingerprint string, flair bool) Splash {
	return Splash{
		nickname:    nickname,
		fingerprint: fingerprint,
		flair:       flair,
	}
}

func (s Splash) Init() tea.Cmd {
	return nil
}

func (s Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "y":
			return s, func() tea.Msg { return EnterTavernMsg{} }
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s Splash) View() tea.View {
	if s.width == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	var b strings.Builder

	// /// TAVRN.SH /// gradient header
	diag := lipgloss.NewStyle().Foreground(ColorBorder).Bold(true).Render("///")
	title := GradientText(" TAVRN.SH ", ColorHighlight, ColorAccent, true)
	header := diag + title + diag
	b.WriteString("\n")
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(SplashSubtitleStyle.Render("a quiet place in the terminal"))
	b.WriteString("\n\n")

	// ASCII tavern art with gradient coloring
	artLines := strings.Split(tavernArt, "\n")
	for _, line := range artLines {
		colored := GradientText(line, ColorAccent, ColorBorder, false)
		b.WriteString(colored)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Identity card
	nick := s.nickname
	if s.flair {
		nick = "~" + nick
	}
	identLabel := SplashDescStyle.Render("you are ")
	identNick := NickStyle(0).Render(nick)
	b.WriteString(identLabel + identNick)
	b.WriteString("\n")

	fpShort := s.fingerprint
	if len(fpShort) > 16 {
		fpShort = fpShort[:16]
	}
	b.WriteString(SplashDescStyle.Render(fmt.Sprintf("key: %s...", fpShort)))
	b.WriteString("\n")

	// Commands section
	b.WriteString("\n")
	b.WriteString(SplashCategoryStyle.Render("COMMANDS"))
	b.WriteString("\n")
	cmds := []struct{ cmd, desc string }{
		{"/nick NAME", "change your handle"},
		{"/who", "see who's around"},
		{"/help", "show all commands"},
	}
	for _, c := range cmds {
		cmd := SplashCommandStyle.Width(18).Render(c.cmd)
		desc := SplashDescStyle.Render(c.desc)
		b.WriteString(fmt.Sprintf("  %s %s\n", cmd, desc))
	}

	// Keys section
	b.WriteString("\n")
	b.WriteString(SplashCategoryStyle.Render("KEYS"))
	b.WriteString("\n")
	keys := []struct{ key, desc string }{
		{"ENTER", "send message"},
		{"CTRL+C", "exit tavern"},
		{"UP / DOWN", "scroll chat"},
	}
	for _, k := range keys {
		key := SplashCommandStyle.Width(18).Render(k.key)
		desc := SplashDescStyle.Render(k.desc)
		b.WriteString(fmt.Sprintf("  %s %s\n", key, desc))
	}

	// Purge notice
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render(
		"all data purged every sunday 23:59 UTC"))
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render(
		"nothing is permanent. draw while you can."))

	// Action footer
	b.WriteString("\n\n")
	enterKey := SplashKeyStyle.Render("[ ENTER ]")
	enterDesc := lipgloss.NewStyle().Foreground(ColorSand).Render(" enter the tavern")
	quitKey := SplashKeyStyle.Render("[ Q ]")
	quitDesc := lipgloss.NewStyle().Foreground(ColorDim).Render(" exit")
	b.WriteString(enterKey + enterDesc + "     " + quitKey + quitDesc)

	// Version
	b.WriteString("\n\n")
	b.WriteString(SplashDescStyle.Render("[ v0.2 ]"))

	// Box it up and center
	box := SplashBorderStyle.Render(b.String())
	bgStyle := lipgloss.NewStyle().Background(ColorDarkBg)
	centered := lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceStyle(bgStyle),
	)

	v := tea.NewView(centered)
	v.AltScreen = true
	return v
}

// EnterTavernMsg signals transition from splash to main tavern UI.
type EnterTavernMsg struct{}

// ShowHelpMsg signals the help overlay should be shown.
type ShowHelpMsg struct{}
