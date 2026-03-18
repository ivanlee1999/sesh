package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	BG, FG               lipgloss.Color
	BGSecondary          lipgloss.Color
	FGSecondary          lipgloss.Color
	Border, BorderFocus  lipgloss.Color
	Accent               lipgloss.Color
	Error                lipgloss.Color
	FocusAccent          lipgloss.Color
	OverflowAccent       lipgloss.Color
	BreakAccent          lipgloss.Color
	LongBreakAccent      lipgloss.Color
	PausedFG             lipgloss.Color
	ProgressEmpty        lipgloss.Color
	StatusBarBG          lipgloss.Color
}

func DarkTheme() Theme {
	return Theme{
		BG:             lipgloss.Color("#1E1E2E"),
		FG:             lipgloss.Color("#CDD6F4"),
		BGSecondary:    lipgloss.Color("#313244"),
		FGSecondary:    lipgloss.Color("#9399B2"),
		Border:         lipgloss.Color("#585B70"),
		BorderFocus:    lipgloss.Color("#A3E635"),
		Accent:         lipgloss.Color("#A3E635"),
		Error:          lipgloss.Color("#E06C75"),
		FocusAccent:    lipgloss.Color("#98C379"),
		OverflowAccent: lipgloss.Color("#E5C07B"),
		BreakAccent:    lipgloss.Color("#61AFEF"),
		LongBreakAccent: lipgloss.Color("#C678DD"),
		PausedFG:       lipgloss.Color("#C678DD"),
		ProgressEmpty:  lipgloss.Color("#45475A"),
		StatusBarBG:    lipgloss.Color("#181825"),
	}
}
