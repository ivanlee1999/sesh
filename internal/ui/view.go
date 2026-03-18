package ui

import (
	"fmt"
	"strings"
	"time"

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

	full := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, statusBar)
	if m.InputMode == app.ModeHelp {
		return overlayHelp(w, m.Height, full)
	}
	return full
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

	if m.InputMode == app.ModeSessionComplete {
		stateStr = stateStyle.Foreground(theme.FocusAccent).Render(" ✓ SESSION COMPLETE ")
		hints = "│ Enter:save   Esc:discard"
		hintStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
		line := stateStr + " " + hintStyle.Render(hints)
		bar := lipgloss.NewStyle().
			Width(w).
			Background(theme.StatusBarBG).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FocusAccent)
		return bar.Render(line)
	}
	if m.InputMode == app.ModeSessionPost {
		stateStr = stateStyle.Foreground(theme.FocusAccent).Render(" ✓ SAVED ")
		hints = "│ b:break  Enter:new session  q:quit"
		hintStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
		line := stateStr + " " + hintStyle.Render(hints)
		bar := lipgloss.NewStyle().
			Width(w).
			Background(theme.StatusBarBG).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FocusAccent)
		return bar.Render(line)
	}

	switch m.Timer.Phase {
	case state.PhaseIdle:
		stateStr = stateStyle.Foreground(theme.FGSecondary).Render(" ○ IDLE ")
		hints = "│ Enter:focus  b:break  i:intention  c:category  q:quit  ?:help"
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
	const minTotalForTimeline = 80
	const timelineW = 30

	showTimeline := w >= minTotalForTimeline

	// Determine the width for the timer panel
	timerW := w
	if showTimeline {
		timerW = w - timelineW
	}

	var sections []string

	// Clock circle
	sections = append(sections, renderClock(m, timerW))

	// Intention & category — shown when set or timer is active
	if m.Intention != "" {
		intentBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.Accent).
			Padding(0, 1).
			Foreground(theme.FG).
			Bold(true).
			Render("▸ " + m.Intention)
		sections = append(sections, center(intentBox, timerW))
	} else if m.Timer.IsActive() {
		noIntent := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("(no intention set)")
		sections = append(sections, center(noIntent, timerW))
	}
	if cat := m.SelectedCategory(); cat != nil {
		dot := lipgloss.NewStyle().Foreground(hexColor(cat.HexColor)).Render("● ")
		catLine := dot + lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(cat.Title)
		sections = append(sections, center(catLine, timerW))
	}

	// Progress bar (when active)
	if m.Timer.IsActive() && m.Timer.Phase != state.PhaseAbandoned {
		sections = append(sections, "")
		sections = append(sections, renderProgressBar(m, timerW))
	}

	// Session info
	info := renderSessionInfo(m)
	if info != "" {
		sections = append(sections, "")
		sections = append(sections, center(info, timerW))
	}

	// Today stats
	sections = append(sections, "")
	statsLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
		fmt.Sprintf("Today: %s focused │ %d sessions │ Streak: %d days",
			app.FormatFocusTime(m.TodayFocusMins), m.TodaySessions, m.Streak))
	sections = append(sections, center(statsLine, timerW))

	// Idle controls
	if m.Timer.Phase == state.PhaseIdle {
		sections = append(sections, "")
		durLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Duration: ") +
			lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(fmt.Sprintf("%d min", m.FocusDurationMins)) +
			lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("  (+/- to adjust)")
		sections = append(sections, center(durLine, timerW))
		helpLine := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			"[Enter] Start Focus  [b] Break  [i] Intention  [c] Category")
		sections = append(sections, center(helpLine, timerW))
	}

	result := strings.Join(sections, "\n")

	// Join with timeline if visible
	if showTimeline {
		// Estimate content height for the timeline panel
		tabH := 2
		statusH := 2
		contentH := m.Height - tabH - statusH
		if contentH < 6 {
			contentH = 6
		}

		timerPanel := lipgloss.NewStyle().Width(timerW).Render(result)
		timelinePanel := renderTimeline(m, timelineW, contentH)
		result = lipgloss.JoinHorizontal(lipgloss.Top, timerPanel, timelinePanel)
	}

	// Overlays (always use full width w so they cover both panels)
	if m.InputMode == app.ModeSessionComplete {
		result = overlaySessionComplete(m, w, m.Height)
	} else if m.InputMode == app.ModeSessionPost {
		result = overlaySessionPost(m, w, m.Height)
	} else if m.InputMode == app.ModeIntention {
		result = overlayIntention(m, w, m.Height)
	} else if m.InputMode == app.ModeCategory {
		result = overlayCategory(m, w, m.Height)
	}

	return result
}

// renderTimeline renders a vertical day-planner view of today's focus sessions.
func renderTimeline(m app.Model, w int, h int) string {
	now := time.Now()

	// Determine time range
	startHour := 6
	endHour := now.Hour() + 2
	if endHour > 24 {
		endHour = 24
	}

	// Adjust range based on actual sessions
	for _, s := range m.TodayTimeline {
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", s.StartedAt, time.Local); err == nil {
			if t.Hour() < startHour {
				startHour = t.Hour()
			}
		}
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", s.EndedAt, time.Local); err == nil {
			eh := t.Hour() + 1
			if eh > endHour {
				endHour = eh
			}
		}
	}
	// Also account for active session
	if !m.StartedAtChrono.IsZero() && m.Timer.IsActive() {
		if m.StartedAtChrono.Hour() < startHour {
			startHour = m.StartedAtChrono.Hour()
		}
	}
	if endHour > 24 {
		endHour = 24
	}
	if endHour <= startHour {
		endHour = startHour + 1
	}

	totalMinutes := (endHour - startHour) * 60

	// Layout constants
	// Inner width after left border character
	innerW := w - 3 // 1 border left + 1 space + content + 1 space right
	if innerW < 10 {
		innerW = 10
	}
	labelW := 6 // " 8 AM " or " 14:00"
	blockW := innerW - labelW - 1
	if blockW < 4 {
		blockW = 4
	}

	// Title row
	titleStyle := lipgloss.NewStyle().Foreground(theme.FG).Bold(true)
	title := titleStyle.Render("Today")

	// Usable rows for the timeline content
	usableRows := h - 2 // 1 title + 1 spacing
	if usableRows < 4 {
		usableRows = 4
	}

	// Map time to row
	timeToRow := func(t time.Time) int {
		mins := (t.Hour()-startHour)*60 + t.Minute()
		if mins < 0 {
			return 0
		}
		row := mins * usableRows / totalMinutes
		if row >= usableRows {
			row = usableRows - 1
		}
		return row
	}

	// Initialize row buffers
	type rowData struct {
		label     string // hour label
		block     string // block content (styled)
		hasBlock  bool
		isNowRow  bool
	}
	rows := make([]rowData, usableRows)

	// Place hour labels
	for hr := startHour; hr < endHour; hr++ {
		t := time.Date(now.Year(), now.Month(), now.Day(), hr, 0, 0, 0, time.Local)
		row := timeToRow(t)
		if row < usableRows {
			var label string
			if hr == 0 {
				label = "12 AM"
			} else if hr < 12 {
				label = fmt.Sprintf("%2d AM", hr)
			} else if hr == 12 {
				label = "12 PM"
			} else {
				label = fmt.Sprintf("%2d PM", hr-12)
			}
			rows[row].label = label
		}
	}

	// Place "now" marker
	nowRow := timeToRow(now)
	if nowRow < usableRows {
		rows[nowRow].isNowRow = true
	}

	// Place completed session blocks
	blockStyle := func(color string) lipgloss.Style {
		c := hexColor(color)
		return lipgloss.NewStyle().Foreground(theme.BG).Background(c)
	}
	accentBlockStyle := func(color string) lipgloss.Style {
		c := hexColor(color)
		return lipgloss.NewStyle().Foreground(c)
	}

	type sessionBlock struct {
		startRow  int
		endRow    int
		color     string
		title     string
		isActive  bool
	}

	var blocks []sessionBlock

	for _, s := range m.TodayTimeline {
		st, err1 := time.ParseInLocation("2006-01-02T15:04:05", s.StartedAt, time.Local)
		et, err2 := time.ParseInLocation("2006-01-02T15:04:05", s.EndedAt, time.Local)
		if err1 != nil || err2 != nil {
			continue
		}
		sr := timeToRow(st)
		er := timeToRow(et)
		if er <= sr {
			er = sr // at least 1 row
		}
		color := "#98C379" // default FocusAccent
		if s.CategoryColor != nil {
			color = *s.CategoryColor
		}
		title := s.Title
		if title == "" {
			title = "(focus)"
		}
		blocks = append(blocks, sessionBlock{sr, er, color, title, false})
	}

	// Active session block
	if !m.StartedAtChrono.IsZero() && m.Timer.IsActive() &&
		m.Timer.Phase != state.PhaseAbandoned &&
		m.Timer.Phase != state.PhaseBreak &&
		m.Timer.Phase != state.PhaseBreakOverflow {
		sr := timeToRow(m.StartedAtChrono)
		er := timeToRow(now)
		if er <= sr {
			er = sr
		}
		color := "#98C379"
		if cat := m.SelectedCategory(); cat != nil {
			color = cat.HexColor
		}
		title := m.Intention
		if title == "" {
			title = "(focusing...)"
		}
		blocks = append(blocks, sessionBlock{sr, er, color, title, true})
	}

	// Render blocks into rows
	for _, b := range blocks {
		for row := b.startRow; row <= b.endRow && row < usableRows; row++ {
			fillChar := "█"
			if b.isActive {
				fillChar = "▓"
			}
			// First row of block gets the title
			if row == b.startRow {
				titleStr := truncate(b.title, blockW-2)
				styled := blockStyle(b.color).Render(" " + titleStr + " ")
				visW := lipgloss.Width(styled)
				if visW < blockW {
					styled += accentBlockStyle(b.color).Render(strings.Repeat(fillChar, blockW-visW))
				}
				rows[row].block = styled
			} else {
				rows[row].block = accentBlockStyle(b.color).Render(strings.Repeat(fillChar, blockW))
			}
			rows[row].hasBlock = true
		}
	}

	// Build output lines
	muted := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	nowStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	borderLine := lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("┄", blockW))

	var lines []string
	lines = append(lines, " "+title)

	for i := 0; i < usableRows; i++ {
		r := rows[i]

		// Label column
		label := r.label
		if label == "" {
			label = "     "
		} else {
			label = muted.Render(fmt.Sprintf("%-5s", label))
		}

		// Separator
		sep := muted.Render("│")

		// Block column
		var block string
		if r.isNowRow && !r.hasBlock {
			nowLabel := nowStyle.Render("◂now")
			pad := blockW - 4
			if pad < 0 {
				pad = 0
			}
			block = nowStyle.Render(strings.Repeat("─", pad)) + nowLabel
		} else if r.isNowRow && r.hasBlock {
			block = r.block
		} else if r.hasBlock {
			block = r.block
		} else if r.label != "" {
			block = borderLine
		} else {
			block = strings.Repeat(" ", blockW)
		}

		lines = append(lines, " "+label+sep+block)
	}

	// Wrap in a left-border panel
	content := strings.Join(lines, "\n")
	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.Border).
		Width(w).
		Height(h).
		MaxHeight(h).
		Render(content)

	return panel
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
	bold := func(s string) string { return lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render(s) }
	muted := func(s string) string { return lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(s) }

	lines := []string{
		"",
		bold("  Today's Summary"),
		muted(fmt.Sprintf("  Total Focus: %s  │  Sessions: %d  │  Streak: %d days",
			app.FormatFocusTime(m.TodayFocusMins), m.TodaySessions, m.Streak)),
		"",
		bold("  Focus — Last 7 Days"),
		"",
	}

	if len(m.Last7Days) > 0 {
		maxHours := 0.0
		for _, d := range m.Last7Days {
			if d.Hours > maxHours {
				maxHours = d.Hours
			}
		}
		if maxHours == 0 {
			maxHours = 1
		}
		barW := 24
		for _, d := range m.Last7Days {
			t, _ := time.Parse("2006-01-02", d.Date)
			dayLabel := t.Format("Mon 02")
			filled := int(d.Hours / maxHours * float64(barW))
			if filled > barW {
				filled = barW
			}
			empty := barW - filled
			bar := lipgloss.NewStyle().Foreground(theme.FocusAccent).Render(strings.Repeat("█", filled)) +
				lipgloss.NewStyle().Foreground(theme.ProgressEmpty).Render(strings.Repeat("░", empty))
			hoursStr := "—"
			if d.Hours > 0 {
				hoursStr = fmt.Sprintf("%.1fh", d.Hours)
			}
			lines = append(lines, fmt.Sprintf("  %s  %s  %s", dayLabel, bar, muted(hoursStr)))
		}
	} else {
		lines = append(lines, muted("  No data yet."))
	}

	lines = append(lines, "", bold("  Today — Category Breakdown"), "")

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
		if filled > barW {
			filled = barW
		}
		empty := barW - filled
		color := hexColor(b.Color)
		dot := lipgloss.NewStyle().Foreground(color).Render("●")
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(theme.ProgressEmpty).Render(strings.Repeat("░", empty))
		line := fmt.Sprintf("  %s %-14s %s %3.0f%%  %s",
			dot, b.Name, bar, pct, muted(app.FormatFocusTime(b.Minutes)))
		lines = append(lines, line)
	}
	if len(m.CatBreakdown) == 0 {
		lines = append(lines, muted("  No sessions today. Start focusing!"))
	}

	return strings.Join(lines, "\n")
}

func renderHistory(m app.Model, w int) string {
	sessions := m.HistorySessions
	if len(sessions) == 0 {
		return "\n" + lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			"  No sessions yet. Start focusing!")
	}

	// Column widths
	const dateW = 10
	const durW = 7
	const catW = 14
	sepW := 2
	intentW := w - dateW - durW - catW - sepW*3 - 4
	if intentW < 8 {
		intentW = 8
	}

	// Header
	hs := lipgloss.NewStyle().Foreground(theme.FGSecondary).Bold(true)
	sep := "  "
	header := "  " +
		hs.Render(padRight("Date", dateW)) + sep +
		hs.Render(padRight("Dur", durW)) + sep +
		hs.Render(padRight("Category", catW)) + sep +
		hs.Render("Intention")
	divider := "  " + lipgloss.NewStyle().Foreground(theme.Border).Render(
		strings.Repeat("─", w-4))

	// Scroll window
	visibleRows := m.Height - 8
	if visibleRows < 5 {
		visibleRows = 5
	}
	start := m.HistoryScrollOffset
	end := start + visibleRows
	if end > len(sessions) {
		end = len(sessions)
	}

	lines := []string{
		"",
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Session History"),
		"",
		header,
		divider,
	}

	// Scroll-up indicator
	if start > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).
			Width(w).Align(lipgloss.Center).Render("↑ more"))
	}

	for i := start; i < end; i++ {
		s := sessions[i]

		date := ""
		if idx := strings.Index(s.StartedAt, "T"); idx > 0 {
			date = s.StartedAt[:idx]
		}
		dur := fmt.Sprintf("%d:%02d", s.ActualSeconds/60, s.ActualSeconds%60)

		catName := "—"
		rowColor := theme.FGSecondary
		if s.CategoryTitle != nil {
			catName = *s.CategoryTitle
		}
		if s.CategoryColor != nil {
			rowColor = hexColor(*s.CategoryColor)
		}

		title := s.Title
		if title == "" {
			title = "(no intention)"
		}

		marker := "  "
		if i == m.HistorySelected {
			marker = lipgloss.NewStyle().Foreground(theme.Accent).Render("> ")
		}

		rc := lipgloss.NewStyle().Foreground(rowColor)
		catStr := lipgloss.NewStyle().Foreground(rowColor).Bold(true).Render(padRight(catName, catW))
		dateStr := rc.Render(padRight(date, dateW))
		durStr := rc.Render(padRight(dur, durW))
		intentStr := rc.Render(truncate(title, intentW))

		line := marker + dateStr + sep + durStr + sep + catStr + sep + intentStr
		lines = append(lines, line)
	}

	// Scroll-down indicator
	if end < len(sessions) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).
			Width(w).Align(lipgloss.Center).Render("↓ more"))
	}

	// Total focus time
	lines = append(lines, divider)
	totalStr := fmt.Sprintf("  Total: %s across %d sessions",
		app.FormatFocusTime(m.TotalFocusMins), len(sessions))
	lines = append(lines, lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(totalStr))

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

func overlaySessionComplete(m app.Model, w, h int) string {
	durStr := state.FormatDuration(m.CompletionDuration)
	boxW := 52
	if boxW > w-4 {
		boxW = w - 4
	}
	inputInnerW := boxW - 8
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.BorderFocus).
		Width(inputInnerW)
	inputText := lipgloss.NewStyle().Foreground(theme.Accent).Render("> ") +
		lipgloss.NewStyle().Foreground(theme.FG).Render(m.CompletionNotes) +
		lipgloss.NewStyle().Background(theme.Accent).Foreground(theme.BG).Render(" ")
	inputField := inputStyle.Render(inputText)

	content := lipgloss.NewStyle().Foreground(theme.FocusAccent).Bold(true).Render("✓  Session Complete") +
		"\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Duration: "+durStr) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(theme.FG).Render("Notes (optional)") +
		"\n" + inputField + "\n\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Enter ↵ save   Esc discard")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.FocusAccent).
		Width(boxW).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(w, h-4, lipgloss.Center, lipgloss.Center, box)
}

func overlaySessionPost(m app.Model, w, h int) string {
	durStr := state.FormatDuration(m.CompletionDuration)
	boxW := 48
	if boxW > w-4 {
		boxW = w - 4
	}
	content := lipgloss.NewStyle().Foreground(theme.FocusAccent).Bold(true).Render("✓  Session Saved") +
		"\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Duration: "+durStr) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("[b]") +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(" Break   ") +
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("[Enter]") +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(" New Session   ") +
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("[q]") +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(" Quit")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.FocusAccent).
		Width(boxW).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(w, h-4, lipgloss.Center, lipgloss.Center, box)
}

func overlayHelp(w, h int, behind string) string {
	boxW := 52
	if boxW > w-4 {
		boxW = w - 4
	}

	kw := boxW - 10 // inner content width
	bold := func(s string) string {
		return lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(s)
	}
	sec := func(s string) string {
		return lipgloss.NewStyle().Foreground(theme.FGSecondary).Bold(true).Render(s)
	}
	row := func(key, action string) string {
		keyW := 18
		keyStr := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render(padRight(key, keyW))
		actStr := lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(action)
		_ = kw
		return keyStr + actStr
	}

	lines := []string{
		bold("Keybindings"),
		"",
		sec("Global"),
		row("1-4 / Tab", "switch tabs"),
		row("?", "toggle this help"),
		row("q", "quit (when idle)"),
		row("Ctrl+C", "force quit"),
		"",
		sec("Timer — Idle"),
		row("Enter", "start focus"),
		row("b / B", "short / long break"),
		row("i", "set intention"),
		row("c", "pick category"),
		row("+/- or >/<", "duration ±5 / ±1 min"),
		"",
		sec("Timer — Focus"),
		row("Space", "pause / resume"),
		row("f", "finish session"),
		row("b", "finish + start break"),
		row("x", "abandon session"),
		"",
		sec("Timer — Overflow"),
		row("f", "finish (enter notes)"),
		row("b", "finish + start break"),
		row("Space", "pause / resume"),
		row("x", "abandon"),
		"",
		sec("Break"),
		row("Enter / f", "end break"),
		"",
		sec("History"),
		row("j / k  or  ↑ / ↓", "scroll"),
		"",
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Esc or ? to close"),
	}

	content := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Width(boxW).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
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

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func config_path() string {
	return "~/.config/sesh/config.toml"
}

func data_dir() string {
	return "~/.local/share/sesh/"
}
