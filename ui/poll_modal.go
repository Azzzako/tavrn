package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PollCreateMsg carries the new poll data from the creation modal.
type PollCreateMsg struct {
	Title   string
	Options []string
}

// ─────────────────────────────────────
// Poll Creation Modal
// ─────────────────────────────────────

type PollModal struct {
	title   textinput.Model
	options [4]textinput.Model
	count   int // number of visible option fields (2-4)
	focus   int // 0=title, 1-4=options
	err     string
}

func NewPollModal() PollModal {
	ti := textinput.New()
	ti.Placeholder = "poll question..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Prompt = "> "

	var opts [4]textinput.Model
	for i := range opts {
		o := textinput.New()
		o.Placeholder = fmt.Sprintf("option %d...", i+1)
		o.CharLimit = 30
		o.Prompt = "> "
		opts[i] = o
	}

	return PollModal{
		title:   ti,
		options: opts,
		count:   2,
		focus:   0,
	}
}

func (p PollModal) Update(msg tea.Msg) (PollModal, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return p, func() tea.Msg { return CloseModalMsg{} }
		case "tab":
			// Cycle focus: title → opt1 → opt2 → ... → title
			p.blurAll()
			p.focus++
			if p.focus > p.count {
				p.focus = 0
			}
			p.focusCurrent()
			return p, nil
		case "shift+tab":
			p.blurAll()
			p.focus--
			if p.focus < 0 {
				p.focus = p.count
			}
			p.focusCurrent()
			return p, nil
		case "ctrl+n":
			if p.count < 4 {
				p.count++
			}
			return p, nil
		case "enter":
			title := strings.TrimSpace(p.title.Value())
			if title == "" {
				p.err = "title required"
				return p, nil
			}
			var opts []string
			for i := 0; i < p.count; i++ {
				v := strings.TrimSpace(p.options[i].Value())
				if v != "" {
					opts = append(opts, v)
				}
			}
			if len(opts) < 2 {
				p.err = "need at least 2 options"
				return p, nil
			}
			return p, func() tea.Msg {
				return PollCreateMsg{Title: title, Options: opts}
			}
		}
	}

	// Forward to focused input
	var cmd tea.Cmd
	if p.focus == 0 {
		p.title, cmd = p.title.Update(msg)
	} else {
		idx := p.focus - 1
		p.options[idx], cmd = p.options[idx].Update(msg)
	}
	p.err = ""
	return p, cmd
}

func (p *PollModal) blurAll() {
	p.title.Blur()
	for i := range p.options {
		p.options[i].Blur()
	}
}

func (p *PollModal) focusCurrent() {
	if p.focus == 0 {
		p.title.Focus()
	} else {
		p.options[p.focus-1].Focus()
	}
}

func (p PollModal) View(width, height int) string {
	modalW := 44
	headerText := " Create Poll "
	fillLen := modalW - len(headerText)
	if fillLen < 4 {
		fillLen = 4
	}
	leftFill := strings.Repeat("╱", fillLen/2)
	rightFill := strings.Repeat("╱", fillLen-fillLen/2)

	headerFill := lipgloss.NewStyle().Foreground(ColorBorder).Render(leftFill)
	headerTitle := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true).Render(headerText)
	headerFillR := lipgloss.NewStyle().Foreground(ColorBorder).Render(rightFill)
	header := headerFill + headerTitle + headerFillR

	dim := lipgloss.NewStyle().Foreground(ColorDim)
	label := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")

	b.WriteString("  " + label.Render("Title"))
	b.WriteString("\n")
	b.WriteString("  " + p.title.View())
	b.WriteString("\n\n")

	for i := 0; i < p.count; i++ {
		b.WriteString("  " + label.Render(fmt.Sprintf("Option %d", i+1)))
		b.WriteString("\n")
		b.WriteString("  " + p.options[i].View())
		b.WriteString("\n")
		if i < p.count-1 {
			b.WriteString("\n")
		}
	}

	if p.count < 4 {
		b.WriteString("\n")
		b.WriteString(dim.Render("  ctrl+n to add option"))
	}

	if p.err != "" {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("131")).Render("  " + p.err))
	}

	b.WriteString("\n\n")
	footerFill := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("╱", modalW))
	b.WriteString(footerFill)
	b.WriteString("\n")

	enter := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true).Render("ENTER")
	tab := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true).Render("TAB")
	esc := lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true).Render("ESC")
	b.WriteString(dim.Render(
		fmt.Sprintf("  %s submit  ·  %s next  ·  %s cancel", enter, tab, esc)))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2).
		Width(modalW + 6).
		Render(b.String())
}
