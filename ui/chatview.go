package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"tavrn/internal/chat"
)

type ChatView struct {
	viewport viewport.Model
	input    textinput.Model
	messages []chat.Message
	width    int
	height   int
}

func NewChatView() ChatView {
	ti := textinput.New()
	ti.Placeholder = "say something..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Prompt = "> "

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))

	return ChatView{
		viewport: vp,
		input:    ti,
		messages: make([]chat.Message, 0),
	}
}

func (c *ChatView) SetSize(width, height int) {
	c.width = width
	c.height = height
	inputHeight := 3 // prompt line + top/bottom padding
	borderHeight := 2
	vpW := width - borderHeight - 2 // border + padding
	vpH := height - inputHeight - borderHeight
	if vpW < 1 {
		vpW = 1
	}
	if vpH < 1 {
		vpH = 1
	}
	c.viewport.SetWidth(vpW)
	c.viewport.SetHeight(vpH)
	c.input.SetWidth(width - 6)
}

func (c *ChatView) AddMessage(msg chat.Message) {
	// Ensure timestamp is set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	c.messages = append(c.messages, msg)
	c.renderMessages()
	c.viewport.GotoBottom()
}

func (c *ChatView) renderMessages() {
	var lines []string

	for i, msg := range c.messages {
		if msg.IsSystem {
			// System messages: dimmed with star
			line := SystemMsgStyle.Render("  * " + msg.Text)
			lines = append(lines, line)
			// Add spacing after system messages
			if i < len(c.messages)-1 {
				lines = append(lines, "")
			}
		} else {
			// User messages: Crush-style with colored left bar
			barColor := NickBarColor(msg.ColorIndex)
			bar := lipgloss.NewStyle().Foreground(barColor).Render("  |")
			barDim := lipgloss.NewStyle().Foreground(barColor).Render("  |")

			// Header line: bar  nick  timestamp
			nick := NickStyle(msg.ColorIndex).Render(msg.Nickname)
			ts := formatTimestamp(msg.Timestamp)
			timeStr := MsgTimeStyle.Render(ts)
			header := fmt.Sprintf("%s %s  %s", bar, nick, timeStr)
			lines = append(lines, header)

			// Body line(s): bar  message text
			// Word-wrap long messages
			msgLines := wordWrap(msg.Text, c.viewport.Width()-6)
			for _, ml := range msgLines {
				body := fmt.Sprintf("%s %s", barDim, ml)
				lines = append(lines, body)
			}

			// Spacing between messages
			if i < len(c.messages)-1 {
				lines = append(lines, "")
			}
		}
	}
	c.viewport.SetContent(strings.Join(lines, "\n"))
}

func formatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case diff < 24*time.Hour:
		return t.Format("15:04")
	default:
		return t.Format("Jan 02 15:04")
	}
}

func wordWrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func (c ChatView) Update(msg tea.Msg) (ChatView, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.input, cmd = c.input.Update(msg)
	cmds = append(cmds, cmd)

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c ChatView) View() string {
	chatContent := c.viewport.View()

	// Input area with separator line
	sep := lipgloss.NewStyle().Foreground(ColorBorder).
		Width(c.width - 4).
		Render(strings.Repeat("─", c.width-6))
	inputLine := "  " + c.input.View()

	inner := lipgloss.JoinVertical(lipgloss.Left, chatContent, sep, inputLine)
	return ChatBorderStyle.Width(c.width).Height(c.height).Padding(0, 1).Render(inner)
}

// InputValue returns current input text and clears it.
func (c *ChatView) InputValue() string {
	val := c.input.Value()
	c.input.Reset()
	return val
}
