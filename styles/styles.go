package styles

import "github.com/charmbracelet/lipgloss"

var (
	Green     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render
	LightBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#ADD8E6")).Render
	DarkRed   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B0000")).Render
	DarkPink  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1493")).Render
)
