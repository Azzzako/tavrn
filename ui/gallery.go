package ui

import (
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"tavrn/internal/store"
)

type GalleryNote struct {
	ID          int
	X, Y        int
	Text        string
	Nickname    string
	Fingerprint string
	ColorIndex  int
}

type GalleryView struct {
	notes       []GalleryNote
	selected    int // index into notes, -1 = none
	dragging    bool
	dragOffsetX int
	dragOffsetY int
	width       int
	height      int
	fingerprint string
	rng         *rand.Rand
}

func NewGalleryView(fingerprint string) GalleryView {
	return GalleryView{
		notes:       make([]GalleryNote, 0),
		selected:    -1,
		fingerprint: fingerprint,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *GalleryView) SetSize(width, height int) {
	g.width = width
	g.height = height
}

func (g *GalleryView) LoadNotes(rows []store.NoteRow) {
	g.notes = make([]GalleryNote, len(rows))
	for i, r := range rows {
		g.notes[i] = GalleryNote{
			ID:          r.ID,
			X:           r.X,
			Y:           r.Y,
			Text:        r.Text,
			Nickname:    r.Nickname,
			Fingerprint: r.Fingerprint,
			ColorIndex:  r.ColorIndex,
		}
	}
}

func (g *GalleryView) AddNote(n GalleryNote) {
	g.notes = append(g.notes, n)
}

func (g *GalleryView) MoveNote(id, x, y int) {
	for i := range g.notes {
		if g.notes[i].ID == id {
			g.notes[i].X = x
			g.notes[i].Y = y
			return
		}
	}
}

func (g *GalleryView) RemoveNote(id int) {
	for i := range g.notes {
		if g.notes[i].ID == id {
			g.notes = append(g.notes[:i], g.notes[i+1:]...)
			if g.selected >= len(g.notes) {
				g.selected = -1
			}
			return
		}
	}
}

func (g *GalleryView) ClearAll() {
	g.notes = nil
	g.selected = -1
}

// RandomPosition finds a spot for a new note.
func (g *GalleryView) RandomPosition() (int, int) {
	maxX := g.width - 22
	maxY := g.height - 6
	if maxX < 2 {
		maxX = 2
	}
	if maxY < 2 {
		maxY = 2
	}
	return 2 + g.rng.Intn(maxX), 1 + g.rng.Intn(maxY)
}

func (g GalleryView) Update(msg tea.Msg) (GalleryView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "d", "delete", "backspace":
			// Delete selected note if it's yours
			if g.selected >= 0 && g.selected < len(g.notes) {
				note := g.notes[g.selected]
				if note.Fingerprint == g.fingerprint {
					return g, func() tea.Msg {
						return GalleryDeleteMsg{NoteID: note.ID}
					}
				}
			}
		case "tab":
			// Cycle selection
			if len(g.notes) > 0 {
				g.selected = (g.selected + 1) % len(g.notes)
			}
		}

	case tea.MouseClickMsg:
		mx, my := msg.Mouse().X, msg.Mouse().Y
		// Hit test — find topmost note under cursor
		hit := -1
		for i := len(g.notes) - 1; i >= 0; i-- {
			n := g.notes[i]
			nw, nh := noteSize(n.Text)
			if mx >= n.X && mx < n.X+nw && my >= n.Y && my < n.Y+nh {
				hit = i
				break
			}
		}
		if hit >= 0 {
			g.selected = hit
			g.dragging = true
			g.dragOffsetX = mx - g.notes[hit].X
			g.dragOffsetY = my - g.notes[hit].Y
		} else {
			g.selected = -1
		}

	case tea.MouseReleaseMsg:
		if g.dragging && g.selected >= 0 {
			g.dragging = false
			note := g.notes[g.selected]
			if note.Fingerprint == g.fingerprint {
				return g, func() tea.Msg {
					return GalleryMoveMsg{NoteID: note.ID, X: note.X, Y: note.Y}
				}
			}
		}
		g.dragging = false

	case tea.MouseMotionMsg:
		if g.dragging && g.selected >= 0 {
			mx, my := msg.Mouse().X, msg.Mouse().Y
			newX := mx - g.dragOffsetX
			newY := my - g.dragOffsetY
			if newX < 0 {
				newX = 0
			}
			if newY < 0 {
				newY = 0
			}
			g.notes[g.selected].X = newX
			g.notes[g.selected].Y = newY
		}
	}

	return g, nil
}

func (g GalleryView) View() string {
	if g.width == 0 || g.height == 0 {
		return ""
	}

	// Build canvas manually — create a grid and stamp notes onto it
	grid := make([][]rune, g.height)
	colors := make([][]int, g.height) // -1 = no color, 0-11 = nick color
	for y := range grid {
		grid[y] = make([]rune, g.width)
		colors[y] = make([]int, g.width)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = -1
		}
	}

	// Empty state
	if len(g.notes) == 0 {
		hint := "the board is empty. press CTRL+P to leave a note."
		hintStyle := lipgloss.NewStyle().Foreground(ColorDim).Italic(true)
		cx := (g.width - len(hint)) / 2
		cy := g.height / 2
		if cx < 0 {
			cx = 0
		}

		// Render as simple centered text
		var lines []string
		for y := 0; y < g.height; y++ {
			if y == cy {
				pad := strings.Repeat(" ", cx)
				lines = append(lines, pad+hintStyle.Render(hint))
			} else {
				lines = append(lines, strings.Repeat(" ", g.width))
			}
		}
		return strings.Join(lines, "\n")
	}

	// Render each note as a bordered box
	for idx, note := range g.notes {
		isSelected := idx == g.selected
		lines := renderNote(note, isSelected, note.Fingerprint == g.fingerprint)
		for dy, line := range lines {
			ry := note.Y + dy
			if ry < 0 || ry >= g.height {
				continue
			}
			rx := note.X
			for _, ch := range line {
				if rx >= 0 && rx < g.width {
					grid[ry][rx] = ch
					colors[ry][rx] = note.ColorIndex
				}
				rx++
			}
		}
	}

	// Render grid to styled string
	var result []string
	for y := 0; y < g.height; y++ {
		var b strings.Builder
		prevColor := -2
		run := ""
		for x := 0; x < g.width; x++ {
			c := colors[y][x]
			ch := string(grid[y][x])
			if c != prevColor {
				// Flush previous run
				if run != "" {
					b.WriteString(colorRun(run, prevColor))
				}
				run = ch
				prevColor = c
			} else {
				run += ch
			}
		}
		if run != "" {
			b.WriteString(colorRun(run, prevColor))
		}
		result = append(result, b.String())
	}
	return strings.Join(result, "\n")
}

func colorRun(text string, colorIdx int) string {
	if colorIdx < 0 {
		return text // no color, plain
	}
	return lipgloss.NewStyle().Foreground(NickColors[colorIdx%len(NickColors)]).Render(text)
}

func renderNote(note GalleryNote, selected, isOwn bool) []string {
	// Note dimensions
	textW := len(note.Text)
	if textW > 18 {
		textW = 18
	}
	innerW := textW + 2
	if innerW < len(note.Nickname)+2 {
		innerW = len(note.Nickname) + 2
	}
	if innerW < 10 {
		innerW = 10
	}

	// Border characters
	tl, tr, bl, br := "╭", "╮", "╰", "╯"
	h, v := "─", "│"
	if selected {
		tl, tr, bl, br = "┏", "┓", "┗", "┛"
		h, v = "━", "┃"
	}

	var lines []string

	// Top border
	top := tl + strings.Repeat(h, innerW) + tr
	lines = append(lines, top)

	// Text lines (word wrap)
	wrapped := simpleWrap(note.Text, innerW-2)
	for _, wl := range wrapped {
		pad := innerW - len(wl)
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, v+" "+wl+strings.Repeat(" ", pad-1)+v)
	}

	// Nick line
	nick := note.Nickname
	if isOwn {
		nick = "~" + nick
	}
	nickPad := innerW - len(nick) - 1
	if nickPad < 0 {
		nickPad = 0
	}
	lines = append(lines, v+" "+nick+strings.Repeat(" ", nickPad)+v)

	// Bottom border
	bot := bl + strings.Repeat(h, innerW) + br
	lines = append(lines, bot)

	return lines
}

func simpleWrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if len(text) <= width {
		return []string{text}
	}
	var lines []string
	for len(text) > width {
		// Find last space within width
		cut := width
		for i := width; i > 0; i-- {
			if text[i] == ' ' {
				cut = i
				break
			}
		}
		lines = append(lines, text[:cut])
		text = strings.TrimLeft(text[cut:], " ")
	}
	if text != "" {
		lines = append(lines, text)
	}
	return lines
}

func noteSize(text string) (int, int) {
	textW := len(text)
	if textW > 18 {
		textW = 18
	}
	innerW := textW + 2
	if innerW < 10 {
		innerW = 10
	}
	w := innerW + 2 // borders
	wrapped := simpleWrap(text, innerW-2)
	h := len(wrapped) + 3 // top border + text lines + nick + bottom border
	return w, h
}

// Messages from gallery to app
type GalleryDeleteMsg struct{ NoteID int }
type GalleryMoveMsg struct{ NoteID, X, Y int }
