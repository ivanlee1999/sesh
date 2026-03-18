package ui

import (
	"fmt"
	"os"
	"path/filepath"
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
	if m.InputMode == app.ModeConfirmSave {
		stateStr = stateStyle.Foreground(theme.FocusAccent).Render(" ⚠ SHORT SESSION ")
		hints = "│ y/Enter:save   n/Esc:discard"
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
	if m.InputMode == app.ModeSettingsEdit {
		stateStr = stateStyle.Foreground(theme.Accent).Render(" ✎ EDITING ")
		hints = "│ Enter:save  Esc:cancel"
		hintStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
		line := stateStr + " " + hintStyle.Render(hints)
		bar := lipgloss.NewStyle().
			Width(w).
			Background(theme.StatusBarBG).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.Accent)
		return bar.Render(line)
	}
	if m.Screen == app.ScreenSettings {
		stateStr = stateStyle.Foreground(theme.FGSecondary).Render(" ⚙ SETTINGS ")
		hints = "│ j/k:navigate  Enter:edit  ?:help"
		hintStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
		line := stateStr + " " + hintStyle.Render(hints)
		bar := lipgloss.NewStyle().
			Width(w).
			Background(theme.StatusBarBG).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FGSecondary)
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
	// Weekly calendar view takes over the full content area
	if m.WeeklyView {
		tabH := 2
		statusH := 2
		contentH := m.Height - tabH - statusH
		if contentH < 6 {
			contentH = 6
		}
		return renderWeeklyCalendar(m, w, contentH)
	}

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
	if m.InputMode == app.ModeConfirmSave {
		result = overlayConfirmSave(m, w, m.Height)
	} else if m.InputMode == app.ModeSessionComplete {
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

// sessionFillChar returns the fill character for a session type.
func sessionFillChar(sessionType string, isActive bool) string {
	if isActive {
		return "▓"
	}
	switch sessionType {
	case "rest":
		return "░"
	default:
		return "█"
	}
}

// sessionColor returns the color for a session based on its type and category.
func sessionColor(sessionType, title string, categoryColor *string) string {
	switch sessionType {
	case "rest":
		if strings.Contains(title, "Long") {
			return string(theme.LongBreakAccent)
		}
		return string(theme.BreakAccent)
	default:
		if categoryColor != nil {
			return *categoryColor
		}
		return string(theme.FocusAccent)
	}
}

// sessionTitle returns the display title for a session.
func sessionTitle(title, sessionType string) string {
	if title != "" {
		return title
	}
	if sessionType == "rest" {
		return "(break)"
	}
	return "(focus)"
}

// formatHourLabel formats an hour (0-23) as "HH AM/PM".
func formatHourLabel(hr int) string {
	if hr == 0 {
		return "12 AM"
	} else if hr < 12 {
		return fmt.Sprintf("%2d AM", hr)
	} else if hr == 12 {
		return "12 PM"
	}
	return fmt.Sprintf("%2d PM", hr-12)
}

// renderTimeline renders a vertical day-planner view of today's sessions.
func renderTimeline(m app.Model, w int, h int) string {
	now := time.Now()

	// Default time range: 8 AM to 11 PM, shifted by scroll offset
	const windowSize = 15 // hours visible
	startHour := 8 + m.TimelineScrollOffset
	endHour := startHour + windowSize
	if startHour < 0 {
		startHour = 0
		endHour = windowSize
	}
	if endHour > 24 {
		endHour = 24
		startHour = endHour - windowSize
	}
	if startHour < 0 {
		startHour = 0
	}

	totalMinutes := (endHour - startHour) * 60

	// Layout constants
	innerW := w - 3
	if innerW < 10 {
		innerW = 10
	}
	labelW := 6
	blockW := innerW - labelW - 1
	if blockW < 4 {
		blockW = 4
	}

	// Title row
	titleStyle := lipgloss.NewStyle().Foreground(theme.FG).Bold(true)
	scrollHint := ""
	if m.TimelineScrollOffset != 0 {
		scrollHint = lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(
			fmt.Sprintf(" %d–%d", startHour, endHour))
	}
	title := titleStyle.Render("Today") + scrollHint

	usableRows := h - 2
	if usableRows < 4 {
		usableRows = 4
	}

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

	type rowData struct {
		label    string
		block    string
		hasBlock bool
		isNowRow bool
	}
	rows := make([]rowData, usableRows)

	// Place hour labels
	for hr := startHour; hr < endHour; hr++ {
		t := time.Date(now.Year(), now.Month(), now.Day(), hr, 0, 0, 0, time.Local)
		row := timeToRow(t)
		if row < usableRows {
			rows[row].label = formatHourLabel(hr)
		}
	}

	// Place "now" marker
	if now.Hour() >= startHour && now.Hour() < endHour {
		nowRow := timeToRow(now)
		if nowRow < usableRows {
			rows[nowRow].isNowRow = true
		}
	}

	// Style helpers
	blockStyle := func(color string) lipgloss.Style {
		c := hexColor(color)
		return lipgloss.NewStyle().Foreground(theme.BG).Background(c)
	}
	accentBlockStyle := func(color string) lipgloss.Style {
		c := hexColor(color)
		return lipgloss.NewStyle().Foreground(c)
	}

	type sessionBlock struct {
		startRow int
		endRow   int
		color    string
		title    string
		fillChar string
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
			er = sr
		}
		color := sessionColor(s.SessionType, s.Title, s.CategoryColor)
		title := sessionTitle(s.Title, s.SessionType)
		fill := sessionFillChar(s.SessionType, false)
		blocks = append(blocks, sessionBlock{sr, er, color, title, fill})
	}

	// Active focus session block
	if !m.StartedAtChrono.IsZero() && m.Timer.IsActive() &&
		m.Timer.Phase != state.PhaseAbandoned &&
		m.Timer.Phase != state.PhaseBreak &&
		m.Timer.Phase != state.PhaseBreakOverflow {
		sr := timeToRow(m.StartedAtChrono)
		er := timeToRow(now)
		if er <= sr {
			er = sr
		}
		color := string(theme.FocusAccent)
		if cat := m.SelectedCategory(); cat != nil {
			color = cat.HexColor
		}
		title := m.Intention
		if title == "" {
			title = "(focusing...)"
		}
		blocks = append(blocks, sessionBlock{sr, er, color, title, "▓"})
	}

	// Active break session block
	if !m.BreakStartedAt.IsZero() &&
		(m.Timer.Phase == state.PhaseBreak || m.Timer.Phase == state.PhaseBreakOverflow) {
		sr := timeToRow(m.BreakStartedAt)
		er := timeToRow(now)
		if er <= sr {
			er = sr
		}
		color := string(theme.BreakAccent)
		title := "(short break...)"
		fill := "░"
		if m.Timer.BreakType == state.BreakLong {
			color = string(theme.LongBreakAccent)
			title = "(long break...)"
		}
		blocks = append(blocks, sessionBlock{sr, er, color, title, fill})
	}

	// Render blocks into rows
	for _, b := range blocks {
		for row := b.startRow; row <= b.endRow && row < usableRows; row++ {
			if row == b.startRow {
				titleStr := truncate(b.title, blockW-2)
				styled := blockStyle(b.color).Render(" " + titleStr + " ")
				visW := lipgloss.Width(styled)
				if visW < blockW {
					styled += accentBlockStyle(b.color).Render(strings.Repeat(b.fillChar, blockW-visW))
				}
				rows[row].block = styled
			} else {
				rows[row].block = accentBlockStyle(b.color).Render(strings.Repeat(b.fillChar, blockW))
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

		label := r.label
		if label == "" {
			label = "     "
		} else {
			label = muted.Render(fmt.Sprintf("%-5s", label))
		}

		sep := muted.Render("│")

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

	// Scroll indicators
	if startHour > 0 {
		lines[1] = " " + muted.Render("  ↑  ") + muted.Render("│") + muted.Render(" scroll up (k/↑)")
	}
	if endHour < 24 {
		lastIdx := len(lines) - 1
		lines[lastIdx] = " " + muted.Render("  ↓  ") + muted.Render("│") + muted.Render(" scroll down (j/↓)")
	}

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

// renderWeeklyCalendar renders a 7-day calendar grid view.
func renderWeeklyCalendar(m app.Model, w int, h int) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// 7 days: today-6 through today
	days := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		days[i] = today.AddDate(0, 0, i-6)
	}

	// Time range (same as daily timeline defaults)
	startHour := 8 + m.TimelineScrollOffset
	endHour := startHour + 15
	if startHour < 0 {
		startHour = 0
		endHour = 15
	}
	if endHour > 24 {
		endHour = 24
		startHour = endHour - 15
	}
	if startHour < 0 {
		startHour = 0
	}

	// Layout
	labelW := 7 // "  8 AM "
	sepW := 1
	availW := w - labelW - sepW - 2
	if availW < 14 {
		availW = 14
	}
	colW := availW / 7
	if colW < 2 {
		colW = 2
	}

	// Usable rows for hours (reserve: title + header + divider + summary + bottom divider = 5)
	usableRows := h - 5
	if usableRows < 4 {
		usableRows = 4
	}

	totalMinutes := (endHour - startHour) * 60

	timeToRow := func(hour, minute int) int {
		mins := (hour-startHour)*60 + minute
		if mins < 0 {
			return 0
		}
		row := mins * usableRows / totalMinutes
		if row >= usableRows {
			row = usableRows - 1
		}
		return row
	}

	// Group sessions by day index (0-6)
	type daySession struct {
		startHour, startMin int
		endHour, endMin     int
		sessionType         string
		title               string
		categoryColor       *string
	}
	daySessions := make([][]daySession, 7)
	dayFocusSecs := make([]int64, 7) // total focus seconds per day

	for _, s := range m.WeeklyTimeline {
		st, err1 := time.ParseInLocation("2006-01-02T15:04:05", s.StartedAt, time.Local)
		if err1 != nil {
			continue
		}
		sessionDay := time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, time.Local)
		dayIdx := -1
		for i, d := range days {
			if sessionDay.Equal(d) {
				dayIdx = i
				break
			}
		}
		if dayIdx < 0 {
			continue
		}

		et, err2 := time.ParseInLocation("2006-01-02T15:04:05", s.EndedAt, time.Local)
		if err2 != nil {
			continue
		}

		daySessions[dayIdx] = append(daySessions[dayIdx], daySession{
			startHour: st.Hour(), startMin: st.Minute(),
			endHour: et.Hour(), endMin: et.Minute(),
			sessionType: s.SessionType, title: s.Title,
			categoryColor: s.CategoryColor,
		})

		if s.SessionType == "full_focus" || s.SessionType == "partial_focus" {
			dayFocusSecs[dayIdx] += s.ActualSeconds - s.PauseSeconds
		}
	}

	// Build grid: rows[row][col] = color/fill info
	type cell struct {
		filled   bool
		fillChar string
		color    string
	}
	grid := make([][]cell, usableRows)
	for i := range grid {
		grid[i] = make([]cell, 7)
	}

	for dayIdx := 0; dayIdx < 7; dayIdx++ {
		for _, s := range daySessions[dayIdx] {
			sr := timeToRow(s.startHour, s.startMin)
			er := timeToRow(s.endHour, s.endMin)
			if er <= sr {
				er = sr
			}
			color := sessionColor(s.sessionType, s.title, s.categoryColor)
			fill := sessionFillChar(s.sessionType, false)
			for row := sr; row <= er && row < usableRows; row++ {
				grid[row][dayIdx] = cell{true, fill, color}
			}
		}
	}

	// Mark active session in today's column (dayIdx=6)
	if !m.StartedAtChrono.IsZero() && m.Timer.IsActive() &&
		m.Timer.Phase != state.PhaseAbandoned &&
		m.Timer.Phase != state.PhaseBreak &&
		m.Timer.Phase != state.PhaseBreakOverflow {
		sr := timeToRow(m.StartedAtChrono.Hour(), m.StartedAtChrono.Minute())
		er := timeToRow(now.Hour(), now.Minute())
		if er <= sr {
			er = sr
		}
		color := string(theme.FocusAccent)
		if cat := m.SelectedCategory(); cat != nil {
			color = cat.HexColor
		}
		for row := sr; row <= er && row < usableRows; row++ {
			grid[row][6] = cell{true, "▓", color}
		}
	}
	if !m.BreakStartedAt.IsZero() &&
		(m.Timer.Phase == state.PhaseBreak || m.Timer.Phase == state.PhaseBreakOverflow) {
		sr := timeToRow(m.BreakStartedAt.Hour(), m.BreakStartedAt.Minute())
		er := timeToRow(now.Hour(), now.Minute())
		if er <= sr {
			er = sr
		}
		color := string(theme.BreakAccent)
		fill := "░"
		if m.Timer.BreakType == state.BreakLong {
			color = string(theme.LongBreakAccent)
		}
		for row := sr; row <= er && row < usableRows; row++ {
			grid[row][6] = cell{true, fill, color}
		}
	}

	// Render
	muted := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	bold := lipgloss.NewStyle().Foreground(theme.FG).Bold(true)
	nowStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)

	var lines []string

	// Title
	titleLine := bold.Render("  Last 7 Days") + "  " + muted.Render("[w] daily view")
	lines = append(lines, titleLine)

	// Day headers
	headerLine := strings.Repeat(" ", labelW+sepW)
	for i, d := range days {
		dayStr := d.Format("Mon 02")
		if d.Equal(today) {
			dayStr = nowStyle.Render(padCenter(dayStr, colW))
		} else {
			dayStr = muted.Render(padCenter(dayStr, colW))
		}
		headerLine += dayStr
		if i < 6 {
			headerLine += " "
		}
	}
	lines = append(lines, headerLine)

	// Hour rows
	for row := 0; row < usableRows; row++ {
		// Determine hour label for this row
		label := ""
		for hr := startHour; hr < endHour; hr++ {
			r := timeToRow(hr, 0)
			if r == row {
				label = formatHourLabel(hr)
				break
			}
		}

		var labelStr string
		if label != "" {
			labelStr = muted.Render(fmt.Sprintf("%6s", label))
		} else {
			labelStr = "      "
		}

		sep := muted.Render("│")
		rowStr := labelStr + sep

		for col := 0; col < 7; col++ {
			c := grid[row][col]
			if c.filled {
				rowStr += lipgloss.NewStyle().Foreground(hexColor(c.color)).
					Render(strings.Repeat(c.fillChar, colW))
			} else {
				// Show now marker for today column
				isNowRow := col == 6 && now.Hour() >= startHour && now.Hour() < endHour &&
					timeToRow(now.Hour(), now.Minute()) == row
				if isNowRow {
					marker := nowStyle.Render(strings.Repeat("─", colW))
					rowStr += marker
				} else {
					rowStr += strings.Repeat(" ", colW)
				}
			}
			if col < 6 {
				rowStr += " "
			}
		}

		lines = append(lines, rowStr)
	}

	// Divider
	divider := muted.Render(strings.Repeat("─", w-2))
	lines = append(lines, divider)

	// Summary row: total focus per day
	summaryLine := muted.Render(fmt.Sprintf("%6s", "Total")) + muted.Render("│")
	for i := 0; i < 7; i++ {
		mins := float64(dayFocusSecs[i]) / 60.0
		var label string
		if mins <= 0 {
			label = "—"
		} else {
			label = app.FormatFocusTime(mins)
		}
		summaryLine += muted.Render(padCenter(label, colW))
		if i < 6 {
			summaryLine += " "
		}
	}
	lines = append(lines, summaryLine)

	return strings.Join(lines, "\n")
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
	const dateW = 14
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
		if t, err := time.Parse("2006-01-02T15:04:05", s.StartedAt); err == nil {
			date = t.Format("Jan 02 15:04")
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
	items := app.BuildSettingsItems()

	headerStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.FGSecondary)
	valueStyle := lipgloss.NewStyle().Foreground(theme.FG)
	cursorStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	editStyle := lipgloss.NewStyle().Foreground(theme.Accent).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Accent)

	// Flash message
	flashStr := ""
	if !m.SettingsSaveFlash.IsZero() && time.Since(m.SettingsSaveFlash) < 1500*time.Millisecond {
		flashStr = lipgloss.NewStyle().Foreground(theme.FocusAccent).Bold(true).Render("Saved!")
	}

	// Title row
	title := lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("  Settings")
	if flashStr != "" {
		pad := w - lipgloss.Width(title) - lipgloss.Width(flashStr) - 4
		if pad < 2 {
			pad = 2
		}
		title = title + strings.Repeat(" ", pad) + flashStr
	}

	var lines []string
	lines = append(lines, "", title, "")

	// Auth status message
	if m.SettingsAuthMsg != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FocusAccent).Render("  "+m.SettingsAuthMsg), "")
	}

	// Calculate visible window
	visibleRows := m.Height - 8
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Render items with scroll
	startIdx := m.SettingsScrollOff
	endIdx := startIdx + visibleRows
	if endIdx > len(items) {
		endIdx = len(items)
	}

	labelW := 24

	for i := startIdx; i < endIdx; i++ {
		item := items[i]

		if item.Kind == app.SettingHeader {
			lines = append(lines, "")
			lines = append(lines, headerStyle.Render("  "+item.Label))
			continue
		}

		prefix := "    "
		if i == m.SettingsCursor {
			prefix = cursorStyle.Render("  > ")
		}

		var valStr string
		isEditing := m.InputMode == app.ModeSettingsEdit && m.SettingsEditIdx == i

		switch item.Kind {
		case app.SettingBool:
			if item.GetBool(&m.Config) {
				valStr = valueStyle.Render("[x]")
			} else {
				valStr = valueStyle.Render("[ ]")
			}
		case app.SettingInt:
			if isEditing {
				valStr = editStyle.Render(m.SettingsEditBuf+"▌") + labelStyle.Render(item.Suffix)
			} else {
				valStr = valueStyle.Render(fmt.Sprintf("%d", item.GetInt(&m.Config))) + labelStyle.Render(item.Suffix)
			}
		case app.SettingString:
			if isEditing {
				valStr = editStyle.Render(m.SettingsEditBuf + "▌")
			} else {
				sv := item.GetString(&m.Config)
				if sv == "" {
					sv = "(not set)"
				} else if item.Sensitive {
					sv = "••••••••"
				}
				valStr = valueStyle.Render(sv)
			}
		case app.SettingAction:
			var status string
			switch item.Key {
			case "calendar.google.auth":
				if calsyncHasToken("google") {
					status = "✓ Authenticated"
				} else {
					status = "Press Enter to auth"
				}
			case "calendar.outlook.auth":
				if calsyncHasToken("outlook") {
					status = "✓ Authenticated"
				} else {
					status = "Press Enter to auth"
				}
			}
			valStr = lipgloss.NewStyle().Foreground(theme.FocusAccent).Render(status)
		}

		label := item.Label
		if len(label) < labelW {
			label = label + strings.Repeat(" ", labelW-len(label))
		}

		line := prefix + labelStyle.Render(label) + valStr
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render(fmt.Sprintf("  Config: %s", config_path())))

	return strings.Join(lines, "\n")
}

func calsyncHasToken(provider string) bool {
	path := filepath.Join(data_dir(), provider+"_token.json")
	_, err := os.Stat(path)
	return err == nil
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

func overlayConfirmSave(m app.Model, w, h int) string {
	durStr := state.FormatDuration(m.CompletionDuration)
	boxW := 48
	if boxW > w-4 {
		boxW = w - 4
	}
	content := lipgloss.NewStyle().Foreground(theme.FocusAccent).Bold(true).Render("⚠  Short Session") +
		"\n" +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render("Session was only "+durStr) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(theme.FG).Render("Save it?") +
		"  " +
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("[y]") +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(" Yes  ") +
		lipgloss.NewStyle().Foreground(theme.FG).Bold(true).Render("[n]") +
		lipgloss.NewStyle().Foreground(theme.FGSecondary).Render(" No")
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
		row("w", "toggle weekly view"),
		row("j/k or ↑/↓", "scroll timeline"),
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
