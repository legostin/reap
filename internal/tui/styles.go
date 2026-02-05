package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("57"))

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2).
			Width(50)

	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	filterPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	childCellStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("243"))

	expandLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	expandValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	parentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
)

var portColorMap = map[string]lipgloss.Color{
	"green":   lipgloss.Color("82"),
	"yellow":  lipgloss.Color("220"),
	"cyan":    lipgloss.Color("87"),
	"magenta": lipgloss.Color("213"),
	"red":     lipgloss.Color("196"),
	"blue":    lipgloss.Color("75"),
	"white":   lipgloss.Color("15"),
	"dim":     lipgloss.Color("241"),
}

func portStyle(colorName string) lipgloss.Style {
	c, ok := portColorMap[colorName]
	if !ok {
		c = portColorMap["dim"]
	}
	return lipgloss.NewStyle().Foreground(c)
}

func colorLegend() string {
	items := []struct {
		color string
		label string
	}{
		{"green", "frontend"},
		{"yellow", "backend"},
		{"cyan", "flask/vite"},
		{"magenta", "postgres"},
		{"red", "redis"},
		{"blue", "mysql/mongo"},
		{"white", "http/s"},
		{"dim", "other"},
	}
	var parts []string
	for _, it := range items {
		dot := portStyle(it.color).Render("●")
		parts = append(parts, dot+" "+it.label)
	}
	return strings.Join(parts, "  ")
}
