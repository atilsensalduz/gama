package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	DocStyle             = lipgloss.NewStyle().Padding(1, 1, 1, 1)
	HighlightColorCyan   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#00AAFF"}
	HighlightColorOrange = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#FFAA00"}
	WindowStyleCyan      = lipgloss.NewStyle().BorderForeground(HighlightColorCyan).Padding(0, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder())
	WindowStyleOrange    = lipgloss.NewStyle().BorderForeground(HighlightColorOrange).Padding(0, 0).Border(lipgloss.RoundedBorder())
	WindowStyleRed       = lipgloss.NewStyle().BorderForeground(lipgloss.Color("9")).Padding(0, 0).Border(lipgloss.RoundedBorder())
	WindowStyleGreen     = lipgloss.NewStyle().BorderForeground(lipgloss.Color("10")).Padding(0, 0).Border(lipgloss.RoundedBorder())
	WindowStyleGray      = lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")).Padding(0, 0).Border(lipgloss.NormalBorder())
	WindowStyleWhite     = lipgloss.NewStyle().BorderForeground(lipgloss.Color("255")).Padding(0, 0).Border(lipgloss.NormalBorder())
	WindowStyleYellow    = lipgloss.NewStyle().BorderForeground(lipgloss.Color("11")).Padding(0, 0).Border(lipgloss.NormalBorder())

	WindowStyleHelp     = WindowStyleGray.Copy().Margin(0, 0, 0, 1).Padding(0, 2, 0, 2)
	WindowStyleError    = WindowStyleRed.Copy().Margin(0, 0, 0, 1).Padding(0, 2, 0, 2)
	WindowStyleProgress = WindowStyleOrange.Copy().Margin(0, 0, 0, 1).Padding(0, 2, 0, 2)
	WindowStyleSuccess  = WindowStyleGreen.Copy().Margin(0, 0, 0, 1).Padding(0, 2, 0, 2)
	WindowStyleDefault  = WindowStyleWhite.Copy().Margin(0, 0, 0, 1).Padding(0, 2, 0, 2)
)

var (
	TitleStyleActive = func() lipgloss.Style {
		b := lipgloss.DoubleBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 2).BorderForeground(lipgloss.Color("205"))
	}()

	TitleStyleDisable = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 2).BorderForeground(lipgloss.Color("255"))
	}()
)
