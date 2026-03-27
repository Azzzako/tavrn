package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

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

// Gradient color pairs that cycle for the art animation
var artGradientPairs = [][2]color.Color{
	{lipgloss.Color("137"), lipgloss.Color("94")},
	{lipgloss.Color("172"), lipgloss.Color("137")},
	{lipgloss.Color("179"), lipgloss.Color("172")},
	{lipgloss.Color("180"), lipgloss.Color("179")},
	{lipgloss.Color("179"), lipgloss.Color("172")},
	{lipgloss.Color("172"), lipgloss.Color("137")},
}

// Pulse frames for the ENTER prompt
var enterPulse = []string{
	"[ ENTER ]",
	"[ ENTER ]",
	"[  >>>  ]",
	"[ ENTER ]",
}

type splashTickMsg time.Time

type Splash struct {
	nickname    string
	fingerprint string
	flair       bool
	width       int
	height      int
	frame       int
}

func NewSplash(nickname, fingerprint string, flair bool) Splash {
	return Splash{
		nickname:    nickname,
		fingerprint: fingerprint,
		flair:       flair,
	}
}

func splashTick() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
		return splashTickMsg(t)
	})
}

func (s Splash) Init() tea.Cmd {
	return splashTick()
}

func (s Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case splashTickMsg:
		s.frame++
		return s, splashTick()
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

	// ╱╱╱ header with animated gradient
	pair := artGradientPairs[s.frame%len(artGradientPairs)]
	diag := GradientText("╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱", pair[0], pair[1], false)
	b.WriteString(diag)
	b.WriteString("\n\n")

	// Title
	title := GradientText("TAVRN.SH", pair[1], pair[0], true)
	b.WriteString("  " + title)
	b.WriteString("\n")
	b.WriteString(SplashSubtitleStyle.Render("  a quiet place in the terminal"))
	b.WriteString("\n\n")

	// ASCII tavern art with cycling gradient
	artLines := strings.Split(tavernArt, "\n")
	for _, line := range artLines {
		colored := GradientText(line, pair[0], pair[1], false)
		b.WriteString(colored)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Identity card
	nick := s.nickname
	if s.flair {
		nick = "~" + nick
	}
	b.WriteString(SplashDescStyle.Render("  you are "))
	b.WriteString(NickStyle(0).Render(nick))
	b.WriteString("\n")
	fpShort := s.fingerprint
	if len(fpShort) > 16 {
		fpShort = fpShort[:16]
	}
	b.WriteString(SplashDescStyle.Render(fmt.Sprintf("  key: %s...", fpShort)))
	b.WriteString("\n")

	// Commands
	b.WriteString("\n")
	b.WriteString(SplashCategoryStyle.Render("  COMMANDS"))
	b.WriteString("\n")
	cmds := []struct{ cmd, desc string }{
		{"/nick NAME", "change your handle"},
		{"/who", "see who's around"},
		{"/help", "show all commands"},
	}
	for _, c := range cmds {
		b.WriteString(fmt.Sprintf("    %s  %s\n",
			SplashCommandStyle.Width(16).Render(c.cmd),
			SplashDescStyle.Render(c.desc)))
	}

	// Keys
	b.WriteString("\n")
	b.WriteString(SplashCategoryStyle.Render("  KEYS"))
	b.WriteString("\n")
	keys := []struct{ key, desc string }{
		{"ENTER", "send message"},
		{"CTRL+C", "exit tavern"},
		{"ESC", "close modals"},
	}
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("    %s  %s\n",
			SplashCommandStyle.Width(16).Render(k.key),
			SplashDescStyle.Render(k.desc)))
	}

	// Purge notice
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render(
		"  all data purged every sunday 23:59 UTC"))
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render(
		"  nothing is permanent. draw while you can."))

	// Animated action footer
	b.WriteString("\n\n")
	enterFrame := enterPulse[s.frame%len(enterPulse)]
	enterKey := SplashKeyStyle.Render(enterFrame)
	enterDesc := lipgloss.NewStyle().Foreground(ColorSand).Render(" enter the tavern")
	quitKey := SplashKeyStyle.Render("[ Q ]")
	quitDesc := lipgloss.NewStyle().Foreground(ColorDim).Render(" exit")
	b.WriteString("  " + enterKey + enterDesc + "     " + quitKey + quitDesc)

	// Bottom decorative fill (animated)
	b.WriteString("\n\n")
	bottomPair := artGradientPairs[(s.frame+3)%len(artGradientPairs)]
	bottomDiag := GradientText("╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱", bottomPair[0], bottomPair[1], false)
	b.WriteString(bottomDiag)

	// Version centered below
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Render("  [ v0.2 ]"))

	// Box and center
	box := SplashBorderStyle.Render(b.String())
	bgStyle := lipgloss.NewStyle().Background(ColorDarkBg)
	centered := lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceStyle(bgStyle))

	v := tea.NewView(centered)
	v.AltScreen = true
	return v
}

// EnterTavernMsg signals transition from splash to main tavern UI.
type EnterTavernMsg struct{}

// ShowHelpMsg signals the help overlay should be shown.
type ShowHelpMsg struct{}
