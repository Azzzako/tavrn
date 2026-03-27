package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// GradientText renders text with a horizontal color gradient using HCL blending.
func GradientText(text string, c1, c2 color.Color, bold bool) string {
	if len(text) == 0 {
		return ""
	}

	runes := []rune(text)
	colors := blendColors(len(runes), c1, c2)

	var b strings.Builder
	for i, r := range runes {
		style := lipgloss.NewStyle().Foreground(colors[i])
		if bold {
			style = style.Bold(true)
		}
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}

func blendColors(n int, stops ...color.Color) []color.Color {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []color.Color{stops[0]}
	}

	result := make([]color.Color, n)
	c1, _ := colorful.MakeColor(stops[0])
	c2, _ := colorful.MakeColor(stops[len(stops)-1])

	for i := range n {
		t := float64(i) / float64(n-1)
		blended := c1.BlendHcl(c2, t).Clamped()
		result[i] = blended
	}
	return result
}
