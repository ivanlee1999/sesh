package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ivanlee1999/sesh/internal/app"
	"github.com/ivanlee1999/sesh/internal/state"
)

var theme = DarkTheme()

func View(m app.Model) string {
	w := m.Width
	if w < 20 {
		w = 80
	}

	var content string
	switch m.Screen {
	case app.ScreenTimer:
		content = renderTimer(m, w)
	case app.ScreenAnalytics:
		content = renderAnalytics(m, w)
	case app.ScreenHistory:
		content = renderHistory(m, w)
	case app.ScreenSettings:
		content = renderSettings(m, w)
	}

	tabBar := renderTabBar(m, w)
	statusBar := renderStatusBar(m, w)

	// Fill content height
	tabH := lipgloss.Height(tabBar)
	statusH := lipgloss.Height(statusBar)
	contentH := m.Height - tabH - statusH
	if contentH < 1 {
		contentH = 1
	}
	content = lipgloss.NewStyle().Height(contentH).MaxHeight(contentH).Width(w).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content, statusBar)
}

func renderTabBar(m app.Model, w int) string {
	tabs := []string{"Timer", "Analytics", "History", "Settings"}
	var parts []string
	for i, t := range tabs {
		style := lipgloss.NewStyle().Padding(0, 2)
		if app.Screen(i) == m.Screen {
			style = style.Foreground(theme.Accent).Bold(true)
		} else {
			style = style.Foreground(theme.FGSecondary)
		}
		parts = append(parts, style.Render(t))
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	border := lipgloss.NewStyle().
		Width(w).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border)
	return border.Render(bar)
}

func renderStatusBar(m app.Model, w int) string {
	var stateStr, hints string
	stateStyle := lipgloss.NewStyle().Bold(true)

	switch m.Timer.Phase {
	case state.PhaseIdle:
		stateStr = stateStyle.Foreground(theme.FGSecondary).Render(" ○ IDLE ")
		hints = "│ Enter:focus  b:break  i:intention  c:category  q:quit"
	case state.PhaseFocus:
		stateStr = stateStyle.Foreground(theme.FocusAccent).Render(
			fmt.Sprintf(" ⏱ %s FOCUS ", m.Timer.DisplayTime()))
		hints = "│ space:pause  f:finish  b:break  x:abandon"
	case state.PhaseOverflow:
		stateStr = stateStyle.Foreground(theme.OverflowAccent).Render(
			fmt.Sprintf(" ◆ %s OVERFLOW ", m.Timer.DisplayTime()))
		hints = "│ f:finish  b:break  x:abandon"
	case state.PhasePaused:
		stateStr = stateStyle.Foreground(theme.PausedFG).Render(
			fmt.Sprintf(" ⏸ %s PAUSED ", m.Timer.DisplayTime()))
		hints = "│ space:resume  f:finish  x:abandon"
	case state.PhaseBreak:
		stateStr = stateStyle.Foreground(theme.BreakAccent).Render(
			fmt.Sprintf(" ☕ %s %s BREAK ", m.Timer.DisplayTime(), m.Timer.BreakType))
		hints = "│ Enter:end break"
	case state.PhaseBreakOverflow:
		stateStr = stateStyle.Foreground(theme.BreakAccent).Render(
			fmt.Sprintf(" ☕ %s BREAK OVER ", m.Timer.DisplayTime()))
		hints = "│ Enter:start focus"
	case state.PhaseAbandoned:
		stateStr = stateStyle.Foreground(theme.Error).Render(" ✖ ABANDONED ")
		hints = "│ u:undo (5s)"
	}

	hintStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	line := stateStr + " " + hintStyle.Render(hints)
	bar := lipgloss.NewStyle().
		Width(w).
		Background(theme.StatusBarBG).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(stateAccentColor(m.Timer.Phase))
	return bar.Render(line)
}

func renderTimer(m app.Model, w int) string {
	var sections []string

	// Clock circle
	sections = append(sections, renderClock(m, w))

	// Intention & category — shown when set or timer is active
	if m.Intention != "" {
		intentBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.Accent).
			Padding(0, 1).
			Foreground(theme.FG).
			Bold(true).
			Render("▸ " + m.Intention)
		sections = append(sections, center(intentBox, w))
	} else if m.Timer.IsActive() {
		noIntent := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("(no intention set)")
		sections = append(sections, center(noIntent, w))
	}
	if cat := m.SelectedCategory(); cat != nil {
		dot := lipgloss.NewStyle().Foreground(hexColor(cat.HexColor)).Render("● ")
		catLine := dot + lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(cat.Title)
		sections = append(sections, center(catLine, w))
	}

	// Progress bar (when active)
	if m.Timer.IsActive() && m.Timer.Phase != state.PhaseAbandoned {
		sections = append(sections, "")
		sections = append(sections, renderProgressBar(m, w))
	}

	// Session info
	info := renderSessionInfo(m)
	if info != "" {
		sections = append(sections, "")
		sections = append(sections, center(info, w))
	}

	// Today stats
	sections = append(sections, "")
	statsLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
		fmt.Sprintf("Today: %s focused │ %d sessions │ Streak: %d days",
			app.FormatFocusTime(m.TodayFocusMins), m.TodaySessions, m.Streak))
	sections = append(sections, center(statsLine, w))

	// Idle controls
	if m.Timer.Phase == state.PhaseIdle {
		sections = append(sections, "")
		durLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Duration: ") +
			lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(fmt.Sprintf("%d min", m.FocusDurationMins)) +
			lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("  (+/- to adjust)")
		sections = append(sections, center(durLine, w))
		helpLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			"[Enter] Start Focus  [b] Break  [i] Intention  [c] Category")
		sections = append(sections, center(helpLine, w))
	}

	result := strings.Join(sections, "\n")

	// Overlays
	if m.InputMode == app.ModeIntention {
		result = overlayIntention(m, w, m.Height)
	} else if m.InputMode == app.ModeCategory {
		result = overlayCategory(m, w, m.Height)
	}

	return result
}

// renderClock draws the timer clock face surrounded by a 20-segment progress ring.
// Ring fills clockwise: 6 top (left→right), 4 right (top→bottom),
// 6 bottom (right→left), 4 left (bottom→top) = 20 total.
func renderClock(m app.Model, w int) string {
	var timeStr string
	var color lipgloss.Color

	switch m.Timer.Phase {
	case state.PhaseIdle:
		timeStr = fmt.Sprintf("%02d:00", m.FocusDurationMins)
		color = theme.FGSecondary
	case state.PhaseFocus:
		timeStr = m.Timer.DisplayTime()
		color = theme.FocusAccent
	case state.PhaseOverflow:
		timeStr = m.Timer.DisplayTime()
		color = theme.OverflowAccent
	case state.PhasePaused:
		timeStr = m.Timer.DisplayTime()
		color = theme.PausedFG
	case state.PhaseBreak:
		timeStr = m.Timer.DisplayTime()
		color = theme.BreakAccent
	case state.PhaseBreakOverflow:
		timeStr = m.Timer.DisplayTime()
		color = theme.BreakAccent
	case state.PhaseAbandoned:
		timeStr = "ABANDONED"
		color = theme.Error
	}

	progress := m.Timer.Progress()
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	const totalSegs = 20
	filled := int(float64(totalSegs) * progress)
	filledDot := lipgloss.NewStyle().Foreground(color).Render("◉")
	emptyDot := lipgloss.NewStyle().Foreground(theme.ProgressEmpty).Render("○")
	seg := func(pos int) string {
		if pos < filled {
			return filledDot
		}
		return emptyDot
	}

	label := m.Timer.Phase.String()
	timeStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(color)
	bdr := lipgloss.NewStyle().Foreground(color)

	// Box: inner content width=14, total box width=18 (╭+16×─+╮ or │+space+14+space+│)
	const innerW = 14
	topBorder := bdr.Render("╭" + strings.Repeat("─", innerW+2) + "╮")
	botBorder := bdr.Render("╰" + strings.Repeat("─", innerW+2) + "╯")
	pipe := bdr.Render("│")
	timeCell := pipe + " " + padCenter(timeStyle.Render(timeStr), innerW) + " " + pipe
	labelCell := pipe + " " + padCenter(labelStyle.Render(label), innerW) + " " + pipe

	// Ring positions (clockwise):
	//   Top    : 0-5   displayed left→right
	//   Right  : 6-9   displayed top→bottom
	//   Bottom : 10-15 displayed right→left (seg(15)..seg(10) left→right)
	//   Left   : 16-19 displayed bottom→top (seg(19)..seg(16) top→bottom)
	//
	// Ring row visible width = 22 (matches side-dot rows: 1+1+18+1+1)
	// 6 dots with 2-space gaps = 16 chars; pad 3 each side → 3+16+3 = 22 ✓
	topRing := "   " + seg(0) + "  " + seg(1) + "  " + seg(2) + "  " + seg(3) + "  " + seg(4) + "  " + seg(5) + "   "
	botRing := "   " + seg(15) + "  " + seg(14) + "  " + seg(13) + "  " + seg(12) + "  " + seg(11) + "  " + seg(10) + "   "

	lines := []string{
		"",
		topRing,
		seg(19) + " " + topBorder + " " + seg(6),
		seg(18) + " " + timeCell + " " + seg(7),
		seg(17) + " " + labelCell + " " + seg(8),
		seg(16) + " " + botBorder + " " + seg(9),
		botRing,
		"",
	}
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, strings.Join(lines, "\n"))
}

func renderProgressBar(m app.Model, w int) string {
	progress := m.Timer.Progress()
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	barW := 30
	if barW > w-10 {
		barW = w - 10
	}
	filled := int(float64(barW) * progress)
	if filled > barW {
		filled = barW
	}
	empty := barW - filled

	var color lipgloss.Color
	switch m.Timer.Phase {
	case state.PhaseFocus:
		color = theme.FocusAccent
	case state.PhaseOverflow:
		color = theme.OverflowAccent
	case state.PhaseBreak, state.PhaseBreakOverflow:
		color = theme.BreakAccent
	case state.PhasePaused:
		color = theme.PausedFG
	default:
		color = theme.FGSecondary
	}

	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(theme.ProgressEmpty).Render(strings.Repeat("░", empty))

	label := fmt.Sprintf(" %d%%", int(progress*100))
	return center(bar+lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(label), w)
}

func renderSessionInfo(m app.Model) string {
	sec := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	switch m.Timer.Phase {
	case state.PhaseFocus:
		elapsed := m.Timer.Target - m.Timer.Remaining
		return sec.Render(fmt.Sprintf("Started: %s  │  Elapsed: %s  │  Target: %s",
			m.Timer.StartedAt.Format("15:04"),
			state.FormatDuration(elapsed),
			state.FormatDuration(m.Timer.Target)))
	case state.PhaseOverflow:
		total := m.Timer.TargetWas + m.Timer.Elapsed
		return sec.Render(fmt.Sprintf("Target: %s  │  Overflow: +%s  │  Total: %s",
			state.FormatDuration(m.Timer.TargetWas),
			state.FormatDuration(m.Timer.Elapsed),
			state.FormatDuration(total)))
	case state.PhasePaused:
		return lipgloss.NewStyle().Foreground(theme.PausedFG).Bold(true).Render(
			"⏸  PAUSED — press Space to resume")
	}
	return ""
}

func renderAnalytics(m app.Model, w int) string {
	title := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Today's Summary")
	summary := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
		fmt.Sprintf("  Total Focus: %s  │  Sessions: %d  │  Streak: %d days",
			app.FormatFocusTime(m.TodayFocusMins), m.TodaySessions, m.Streak))

	lines := []string{"", title, summary, ""}

	catTitle := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Category Breakdown")
	lines = append(lines, catTitle, "")

	var total float64
	for _, b := range m.CatBreakdown {
		total += b.Minutes
	}

	for _, b := range m.CatBreakdown {
		pct := 0.0
		if total > 0 {
			pct = b.Minutes / total * 100
		}
		barW := 20
		filled := int(pct / 100 * float64(barW))
		empty := barW - filled
		color := hexColor(b.Color)

		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(theme.ProgressEmpty).Render(strings.Repeat("░", empty))

		line := fmt.Sprintf("  %s %3.0f%% %-16s %s",
			bar, pct, b.Name, app.FormatFocusTime(b.Minutes))
		lines = append(lines, line)
	}

	if len(m.CatBreakdown) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			"  No sessions today. Start focusing!"))
	}

	return strings.Join(lines, "\n")
}

func renderHistory(m app.Model, w int) string {
	sessions, _ := m.DB.GetSessions(50)
	if len(sessions) == 0 {
		return "\n" + lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			"  No sessions yet. Start focusing!")
	}

	lines := []string{
		"",
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Session History"),
		"",
	}

	currentDate := ""
	for i, s := range sessions {
		datePart := ""
		if idx := strings.Index(s.StartedAt, "T"); idx > 0 {
			datePart = s.StartedAt[:idx]
		}
		if datePart != currentDate {
			currentDate = datePart
			lines = append(lines, "")
			lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).Bold(true).Render(
				fmt.Sprintf("  ── %s ──", currentDate)))
			lines = append(lines, "")
		}

		startTime := extractTime(s.StartedAt)
		endTime := extractTime(s.EndedAt)
		durMins := s.ActualSeconds / 60
		durSecs := s.ActualSeconds % 60

		catName := "—"
		if s.CategoryTitle != nil {
			catName = *s.CategoryTitle
		}

		marker := "  "
		if i == m.HistorySelected {
			marker = lipgloss.NewStyle().Foreground(theme.Accent).Render("> ")
		}

		typeIcon := "●"
		if s.SessionType == "partial_focus" {
			typeIcon = "◐"
		}

		title := s.Title
		if title == "" {
			title = "(no intention)"
		}

		line := fmt.Sprintf("%s%s %s - %s  %-28s %-14s %d:%02d",
			marker, typeIcon, startTime, endTime, truncate(title, 26), catName, durMins, durSecs)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderSettings(m app.Model, w int) string {
	sec := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	lines := []string{
		"",
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Settings"),
		"",
		sec.Render(fmt.Sprintf("  Theme:              %s", m.Config.General.Theme)),
		sec.Render(fmt.Sprintf("  Focus Duration:     %d min", m.Config.Timer.FocusDuration)),
		sec.Render(fmt.Sprintf("  Short Break:        %d min", m.Config.Timer.ShortBreakDuration)),
		sec.Render(fmt.Sprintf("  Long Break:         %d min", m.Config.Timer.LongBreakDuration)),
		sec.Render(fmt.Sprintf("  Long Break After:   %d min cumulative", m.Config.Timer.LongBreakAfter)),
		"",
		sec.Render(fmt.Sprintf("  Config: %s", config_path())),
		sec.Render(fmt.Sprintf("  Data:   %s", data_dir())),
	}
	return strings.Join(lines, "\n")
}

func overlayIntention(m app.Model, w, h int) string {
	boxW := 52
	if boxW > w-4 {
		boxW = w - 4
	}

	// Input field with underline-style border
	inputInnerW := boxW - 8 // account for outer padding + prompt chars
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.BorderFocus).
		Width(inputInnerW)
	inputText := lipgloss.NewStyle().Foreground(theme.Accent).Render("> ") +
		lipgloss.NewStyle().Foreground(theme.FG).Render(m.Intention) +
		lipgloss.NewStyle().Background(theme.Accent).Foreground(theme.BG).Render(" ")
	inputField := inputStyle.Render(inputText)

	content := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("What are you working on?") +
		"\n\n" + inputField + "\n\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Enter ↵ confirm   Esc cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderFocus).
		Width(boxW).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(w, h-4, lipgloss.Center, lipgloss.Center, box)
}

func overlayCategory(m app.Model, w, h int) string {
	const maxVisible = app.CatMaxVisible
	boxW := 44
	if boxW > w-4 {
		boxW = w - 4
	}
	rowW := boxW - 4 // inner width after padding

	start := m.CatScrollOffset
	end := start + maxVisible
	if end > len(m.Categories) {
		end = len(m.Categories)
	}

	var lines []string

	// Scroll-up indicator
	if start > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).
			Width(rowW).Align(lipgloss.Center).Render("↑ more"))
	}

	for i := start; i < end; i++ {
		cat := m.Categories[i]
		dot := lipgloss.NewStyle().Foreground(hexColor(cat.HexColor)).Render("● ")
		if i == m.CatIdx {
			row := lipgloss.NewStyle().
				Background(theme.BGSecondary).
				Foreground(theme.FG).
				Bold(true).
				Width(rowW).
				Render(dot + cat.Title)
			lines = append(lines, row)
		} else {
			row := lipgloss.NewStyle().
				Foreground(theme.FGSecondary).
				Width(rowW).
				Render(dot + cat.Title)
			lines = append(lines, row)
		}
	}

	// Scroll-down indicator
	if end < len(m.Categories) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).
			Width(rowW).Align(lipgloss.Center).Render("↓ more"))
	}

	countHint := ""
	if len(m.Categories) > 0 {
		countHint = fmt.Sprintf(" (%d/%d)", m.CatIdx+1, len(m.Categories))
	}

	content := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("Select Category"+countHint) +
		"\n\n" + strings.Join(lines, "\n") + "\n\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("↑/↓ navigate   Enter ↵ confirm   Esc cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderFocus).
		Width(boxW).
		Padding(1, 2).
		Render(content)
	return lipgloss.Place(w, h-4, lipgloss.Center, lipgloss.Center, box)
}

// Helpers

func center(s string, w int) string {
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, s)
}

// padCenter centers s (which may contain ANSI codes) within a field of visible width w.
func padCenter(s string, w int) string {
	sW := lipgloss.Width(s)
	pad := w - sW
	if pad <= 0 {
		return s
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// stateAccentColor returns the accent color for the given timer phase.
func stateAccentColor(phase state.TimerPhase) lipgloss.Color {
	switch phase {
	case state.PhaseFocus:
		return theme.FocusAccent
	case state.PhaseOverflow:
		return theme.OverflowAccent
	case state.PhaseBreak, state.PhaseBreakOverflow:
		return theme.BreakAccent
	case state.PhasePaused:
		return theme.PausedFG
	case state.PhaseAbandoned:
		return theme.Error
	default:
		return theme.Border
	}
}

func hexColor(hex string) lipgloss.Color {
	return lipgloss.Color(hex)
}

func extractTime(dt string) string {
	if idx := strings.Index(dt, "T"); idx > 0 && len(dt) >= idx+6 {
		return dt[idx+1 : idx+6]
	}
	return "??:??"
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func config_path() string {
	return "~/.config/sesh/config.toml"
}

func data_dir() string {
	return "~/.local/share/sesh/"
}
