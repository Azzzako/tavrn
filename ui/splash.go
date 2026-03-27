package ui

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const tavernArt = `      _____
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

var artGradientPairs = [][2]color.Color{
	{lipgloss.Color("137"), lipgloss.Color("94")},
	{lipgloss.Color("172"), lipgloss.Color("137")},
	{lipgloss.Color("179"), lipgloss.Color("172")},
	{lipgloss.Color("180"), lipgloss.Color("179")},
	{lipgloss.Color("179"), lipgloss.Color("172")},
	{lipgloss.Color("172"), lipgloss.Color("137")},
}

var enterPulse = []string{
	"[ ENTER ]",
	"[ ENTER ]",
	"[  >>>  ]",
	"[ ENTER ]",
}

// Background particle characters and their dim colors
var particleChars = []string{"·", ".", "*", ":", "°", "+", "·"}
var particleColors = []color.Color{
	lipgloss.Color("236"),
	lipgloss.Color("237"),
	lipgloss.Color("238"),
	lipgloss.Color("239"),
	lipgloss.Color("240"),
}

type particle struct {
	x, y  int
	char  int // index into particleChars
	color int // index into particleColors
	speed int // frames between moves
	age   int
}

type splashTickMsg time.Time

type Splash struct {
	nickname    string
	fingerprint string
	flair       bool
	width       int
	height      int
	frame       int
	particles   []particle
	rng         *rand.Rand
}

func NewSplash(nickname, fingerprint string, flair bool) Splash {
	return Splash{
		nickname:    nickname,
		fingerprint: fingerprint,
		flair:       flair,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Splash) initParticles() {
	if s.width == 0 || s.height == 0 {
		return
	}
	count := (s.width * s.height) / 40 // ~2.5% density
	if count > 200 {
		count = 200
	}
	s.particles = make([]particle, count)
	for i := range s.particles {
		s.particles[i] = particle{
			x:     s.rng.Intn(s.width),
			y:     s.rng.Intn(s.height),
			char:  s.rng.Intn(len(particleChars)),
			color: s.rng.Intn(len(particleColors)),
			speed: 2 + s.rng.Intn(4),
			age:   s.rng.Intn(20),
		}
	}
}

func (s *Splash) tickParticles() {
	for i := range s.particles {
		p := &s.particles[i]
		p.age++
		if p.age%p.speed == 0 {
			// Drift upward and slightly sideways
			p.y--
			if s.rng.Intn(3) == 0 {
				p.x += s.rng.Intn(3) - 1 // -1, 0, or 1
			}
			// Cycle character occasionally
			if s.rng.Intn(8) == 0 {
				p.char = s.rng.Intn(len(particleChars))
			}
			// Respawn at bottom if off screen
			if p.y < 0 || p.x < 0 || p.x >= s.width {
				p.y = s.height - 1 - s.rng.Intn(3)
				p.x = s.rng.Intn(s.width)
				p.color = s.rng.Intn(len(particleColors))
			}
		}
	}
}

func splashTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
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
		s.initParticles()
		return s, nil
	case splashTickMsg:
		s.frame++
		s.tickParticles()
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

	// Build the splash card content (centered text)
	card := s.renderCard()

	// Render card into a bordered box
	box := SplashBorderStyle.Render(card)
	boxLines := strings.Split(box, "\n")
	boxH := len(boxLines)
	boxW := 0
	for _, l := range boxLines {
		w := lipgloss.Width(l)
		if w > boxW {
			boxW = w
		}
	}

	// Build the full background with particles
	bg := s.renderBackground()
	bgLines := strings.Split(bg, "\n")

	// Center the box on the background
	startY := (s.height - boxH) / 2
	startX := (s.width - boxW) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Composite box onto background
	for i, bLine := range boxLines {
		row := startY + i
		if row >= len(bgLines) {
			break
		}
		// Replace the center of the bg line with the box line
		bgRunes := []rune(bgLines[row])
		leftPad := ""
		if startX > 0 && startX < len(bgRunes) {
			leftPad = string(bgRunes[:startX])
		} else if startX > 0 {
			leftPad = strings.Repeat(" ", startX)
		}
		rightStart := startX + boxW
		rightPad := ""
		if rightStart < len(bgRunes) {
			rightPad = string(bgRunes[rightStart:])
		}
		bgLines[row] = leftPad + bLine + rightPad
	}

	result := strings.Join(bgLines, "\n")
	v := tea.NewView(result)
	v.AltScreen = true
	return v
}

func (s Splash) renderBackground() string {
	// Build a lookup map: y*width+x → particle index
	type pInfo struct {
		char  string
		color color.Color
	}
	lookup := make(map[int]pInfo, len(s.particles))
	for _, p := range s.particles {
		if p.y >= 0 && p.y < s.height && p.x >= 0 && p.x < s.width {
			lookup[p.y*s.width+p.x] = pInfo{
				char:  particleChars[p.char%len(particleChars)],
				color: particleColors[p.color%len(particleColors)],
			}
		}
	}

	var lines []string
	for y := 0; y < s.height; y++ {
		var b strings.Builder
		for x := 0; x < s.width; x++ {
			if pi, ok := lookup[y*s.width+x]; ok {
				b.WriteString(lipgloss.NewStyle().Foreground(pi.color).Render(pi.char))
			} else {
				b.WriteRune(' ')
			}
		}
		lines = append(lines, b.String())
	}
	return strings.Join(lines, "\n")
}

func (s Splash) renderCard() string {
	pair := artGradientPairs[s.frame%len(artGradientPairs)]

	var b strings.Builder

	// Top decorative fill
	diag := GradientText(strings.Repeat("╱", 44), pair[0], pair[1], false)
	b.WriteString(diag)
	b.WriteString("\n\n")

	// Title centered
	title := GradientText("TAVRN.SH", pair[1], pair[0], true)
	b.WriteString(centerText(title, 8, 44))
	b.WriteString("\n")
	sub := SplashSubtitleStyle.Render("a quiet place in the terminal")
	b.WriteString(centerText(sub, 29, 44))
	b.WriteString("\n\n")

	// Art centered
	artLines := strings.Split(tavernArt, "\n")
	for _, line := range artLines {
		colored := GradientText(line, pair[0], pair[1], false)
		b.WriteString(colored)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Identity
	nick := s.nickname
	if s.flair {
		nick = "~" + nick
	}
	b.WriteString(SplashDescStyle.Render("you are ") + NickStyle(0).Render(nick))
	b.WriteString("\n")
	fpShort := s.fingerprint
	if len(fpShort) > 16 {
		fpShort = fpShort[:16]
	}
	b.WriteString(SplashDescStyle.Render(fmt.Sprintf("key: %s...", fpShort)))

	// Commands
	b.WriteString("\n\n")
	b.WriteString(SplashCategoryStyle.Render("COMMANDS"))
	b.WriteString("\n")
	for _, c := range []struct{ cmd, desc string }{
		{"/nick NAME", "change your handle"},
		{"/who", "see who's around"},
		{"/help", "show all commands"},
	} {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			SplashCommandStyle.Width(16).Render(c.cmd),
			SplashDescStyle.Render(c.desc)))
	}

	b.WriteString("\n")
	b.WriteString(SplashCategoryStyle.Render("KEYS"))
	b.WriteString("\n")
	for _, k := range []struct{ key, desc string }{
		{"ENTER", "send message"},
		{"CTRL+C", "exit tavern"},
		{"ESC", "close modals"},
	} {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			SplashCommandStyle.Width(16).Render(k.key),
			SplashDescStyle.Render(k.desc)))
	}

	// Purge
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render("all data purged every sunday 23:59 UTC"))
	b.WriteString("\n")
	b.WriteString(SplashDescStyle.Italic(true).Render("nothing is permanent. draw while you can."))

	// Animated enter prompt
	b.WriteString("\n\n")
	enterFrame := enterPulse[s.frame%len(enterPulse)]
	enterKey := SplashKeyStyle.Render(enterFrame)
	enterDesc := lipgloss.NewStyle().Foreground(ColorSand).Render(" enter the tavern")
	quitKey := SplashKeyStyle.Render("[ Q ]")
	quitDesc := lipgloss.NewStyle().Foreground(ColorDim).Render(" exit")
	b.WriteString(enterKey + enterDesc + "    " + quitKey + quitDesc)

	// Bottom fill
	b.WriteString("\n\n")
	bottomPair := artGradientPairs[(s.frame+3)%len(artGradientPairs)]
	b.WriteString(GradientText(strings.Repeat("╱", 44), bottomPair[0], bottomPair[1], false))
	b.WriteString("\n")
	b.WriteString(centerText(SplashDescStyle.Render("[ v0.2 ]"), 8, 44))

	return b.String()
}

func centerText(rendered string, rawLen, totalWidth int) string {
	pad := (totalWidth - rawLen) / 2
	if pad <= 0 {
		return rendered
	}
	return strings.Repeat(" ", pad) + rendered
}

type EnterTavernMsg struct{}
type ShowHelpMsg struct{}
