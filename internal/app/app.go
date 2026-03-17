package app

import (
	"fmt"
	"time"

	"github.com/ivanlee1999/sesh/internal/calendar"
	"github.com/ivanlee1999/sesh/internal/config"
	"github.com/ivanlee1999/sesh/internal/db"
	"github.com/ivanlee1999/sesh/internal/state"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenTimer Screen = iota
	ScreenAnalytics
	ScreenHistory
	ScreenSettings
)

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeIntention
	ModeCategory
)

type tickMsg time.Time

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type Model struct {
	Timer       state.TimerState
	Screen      Screen
	InputMode   InputMode
	Quitting    bool

	// Config
	Config         config.Config
	FocusDurationMins int
	TargetDuration time.Duration

	// Intention & category
	Intention  string
	Categories []db.Category
	CatIdx     int

	// History
	HistorySelected int

	// Analytics
	TodayFocusMins  float64
	TodaySessions   int64
	Streak          int64
	CatBreakdown    []db.CategoryBreakdown

	// DB
	DB *db.Database

	// Internal
	StartedAtChrono time.Time
	CumulativeFocus time.Duration
	PauseAccum      time.Duration

	// Terminal size
	Width  int
	Height int
}

func NewModel(database *db.Database, cfg config.Config) Model {
	cats, _ := database.GetCategories()
	focusMins, sessions, _ := database.GetTodayStats()
	streak := database.GetStreak()
	breakdown, _ := database.GetCategoryBreakdownToday()

	return Model{
		Timer:            state.NewIdle(),
		Screen:           ScreenTimer,
		InputMode:        ModeNormal,
		Config:           cfg,
		FocusDurationMins: cfg.Timer.FocusDuration,
		TargetDuration:   time.Duration(cfg.Timer.FocusDuration) * time.Minute,
		Categories:       cats,
		TodayFocusMins:   focusMins,
		TodaySessions:    sessions,
		Streak:           streak,
		CatBreakdown:     breakdown,
		DB:               database,
		Width:            80,
		Height:           24,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd(time.Duration(m.Config.General.TickRateMs) * time.Millisecond)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tickMsg:
		m.tick()
		return m, tickCmd(time.Duration(m.Config.General.TickRateMs) * time.Millisecond)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) tick() {
	d := time.Duration(m.Config.General.TickRateMs) * time.Millisecond
	switch m.Timer.Phase {
	case state.PhaseFocus:
		if m.Timer.Remaining <= d {
			m.Timer.Phase = state.PhaseOverflow
			m.Timer.Elapsed = 0
			m.Timer.TargetWas = m.Timer.Target
		} else {
			m.Timer.Remaining -= d
		}
	case state.PhaseOverflow:
		m.Timer.Elapsed += d
	case state.PhaseBreak:
		if m.Timer.Remaining <= d {
			m.Timer.Phase = state.PhaseBreakOverflow
			m.Timer.Elapsed = 0
		} else {
			m.Timer.Remaining -= d
		}
	case state.PhaseBreakOverflow:
		m.Timer.Elapsed += d
	case state.PhaseAbandoned:
		if time.Now().After(m.Timer.UndoDeadline) {
			m.Timer = state.NewIdle()
		}
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Input modes
	if m.InputMode == ModeIntention {
		return m.handleIntentionKey(msg)
	}
	if m.InputMode == ModeCategory {
		return m.handleCategoryKey(msg)
	}

	key := msg.String()

	// Ctrl+C always quits
	if key == "ctrl+c" {
		m.Quitting = true
		return m, tea.Quit
	}

	// Global keys
	switch key {
	case "q":
		if !m.Timer.IsActive() {
			m.Quitting = true
			return m, tea.Quit
		}
	case "1":
		m.Screen = ScreenTimer
	case "2":
		m.refreshStats()
		m.Screen = ScreenAnalytics
	case "3":
		m.Screen = ScreenHistory
	case "4":
		m.Screen = ScreenSettings
	case "tab":
		m.Screen = (m.Screen + 1) % 4
		if m.Screen == ScreenAnalytics {
			m.refreshStats()
		}
	}

	// Screen-specific
	switch m.Screen {
	case ScreenTimer:
		return m.handleTimerKey(msg)
	case ScreenHistory:
		return m.handleHistoryKey(msg)
	}
	return m, nil
}

func (m Model) handleIntentionKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.InputMode = ModeNormal
	case "enter":
		m.InputMode = ModeNormal
	case "backspace":
		if len(m.Intention) > 0 {
			m.Intention = m.Intention[:len(m.Intention)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.Intention += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleCategoryKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.InputMode = ModeNormal
	case "up", "k":
		if m.CatIdx > 0 {
			m.CatIdx--
		}
	case "down", "j":
		if m.CatIdx < len(m.Categories)-1 {
			m.CatIdx++
		}
	}
	return m, nil
}

func (m Model) handleTimerKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	switch m.Timer.Phase {
	case state.PhaseIdle:
		switch key {
		case "enter":
			m.startFocus()
		case "b":
			m.startBreak(state.BreakShort)
		case "B":
			m.startBreak(state.BreakLong)
		case "i":
			m.InputMode = ModeIntention
		case "c":
			m.InputMode = ModeCategory
		case "+", "=":
			m.adjustDuration(5)
		case "-":
			m.adjustDuration(-5)
		case ">", ".":
			m.adjustDuration(1)
		case "<", ",":
			m.adjustDuration(-1)
		}
	case state.PhaseFocus, state.PhaseOverflow:
		switch key {
		case " ":
			m.togglePause()
		case "f":
			m.finishSession()
		case "x":
			m.abandonSession()
		case "b":
			m.finishSession()
			m.startBreak(state.BreakShort)
		}
	case state.PhasePaused:
		switch key {
		case " ":
			m.togglePause()
		case "f":
			m.finishSession()
		case "x":
			m.abandonSession()
		}
	case state.PhaseBreak, state.PhaseBreakOverflow:
		switch key {
		case "enter", "f":
			m.Timer = state.NewIdle()
		}
	case state.PhaseAbandoned:
		switch key {
		case "u":
			m.undoAbandon()
		default:
			m.Timer = state.NewIdle()
		}
	}
	return m, nil
}

func (m Model) handleHistoryKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.HistorySelected > 0 {
			m.HistorySelected--
		}
	case "down", "j":
		m.HistorySelected++
	}
	return m, nil
}

func (m *Model) startFocus() {
	now := time.Now()
	m.StartedAtChrono = now
	m.PauseAccum = 0
	m.Timer = state.TimerState{
		Phase:     state.PhaseFocus,
		Remaining: m.TargetDuration,
		Target:    m.TargetDuration,
		StartedAt: now,
	}
}

func (m *Model) startBreak(bt state.BreakType) {
	dur := time.Duration(m.Config.Timer.ShortBreakDuration) * time.Minute
	if bt == state.BreakLong {
		dur = time.Duration(m.Config.Timer.LongBreakDuration) * time.Minute
	}
	m.Timer = state.TimerState{
		Phase:     state.PhaseBreak,
		Remaining: dur,
		Target:    dur,
		BreakType: bt,
		StartedAt: time.Now(),
	}
}

func (m *Model) togglePause() {
	switch m.Timer.Phase {
	case state.PhaseFocus, state.PhaseOverflow:
		m.Timer.PausedPhase = m.Timer.Phase
		m.Timer.Phase = state.PhasePaused
		m.Timer.PausedAt = time.Now()
	case state.PhasePaused:
		pauseDur := time.Since(m.Timer.PausedAt)
		m.PauseAccum += pauseDur
		m.Timer.TotalPaused = m.PauseAccum
		m.Timer.Phase = m.Timer.PausedPhase
	}
}

func (m *Model) finishSession() {
	if m.StartedAtChrono.IsZero() {
		m.Timer = state.NewIdle()
		return
	}

	now := time.Now()
	totalElapsed := now.Sub(m.StartedAtChrono)
	pauseSecs := int64(m.PauseAccum.Seconds())
	actualSecs := int64(totalElapsed.Seconds()) - pauseSecs
	targetSecs := int64(m.TargetDuration.Seconds())
	overflowSecs := actualSecs - targetSecs
	if overflowSecs < 0 {
		overflowSecs = 0
	}

	sessionType := "full_focus"
	if actualSecs < targetSecs {
		sessionType = "partial_focus"
	}

	var catID *string
	var catTitle, catColor *string
	if m.CatIdx < len(m.Categories) {
		catID = &m.Categories[m.CatIdx].ID
		catTitle = &m.Categories[m.CatIdx].Title
		catColor = &m.Categories[m.CatIdx].HexColor
	}

	startedStr := m.StartedAtChrono.Format("2006-01-02T15:04:05")
	endedStr := now.Format("2006-01-02T15:04:05")

	m.DB.SaveSession(
		m.Intention, catID, sessionType,
		targetSecs, actualSecs, pauseSecs, overflowSecs,
		startedStr, endedStr, nil,
	)

	// Auto-export to ICS
	if m.Config.Calendar.Enabled && m.Config.Calendar.AutoExport {
		rec := db.SessionRecord{
			ID: "auto", Title: m.Intention,
			CategoryID: catID, CategoryTitle: catTitle, CategoryColor: catColor,
			SessionType: sessionType, TargetSeconds: targetSecs,
			ActualSeconds: actualSecs, PauseSeconds: pauseSecs,
			OverflowSeconds: overflowSecs, StartedAt: startedStr, EndedAt: endedStr,
		}
		calendar.AutoExportSession(&rec, m.Config.Calendar.ICSPath)
	}

	if actualSecs > 0 {
		m.CumulativeFocus += time.Duration(actualSecs) * time.Second
	}
	m.refreshStats()
	m.Timer = state.NewIdle()
	m.StartedAtChrono = time.Time{}
}

func (m *Model) abandonSession() {
	prev := m.Timer
	m.Timer = state.TimerState{
		Phase:         state.PhaseAbandoned,
		PreviousState: &prev,
		UndoDeadline:  time.Now().Add(5 * time.Second),
	}
}

func (m *Model) undoAbandon() {
	if m.Timer.PreviousState != nil && time.Now().Before(m.Timer.UndoDeadline) {
		m.Timer = *m.Timer.PreviousState
	}
}

func (m *Model) adjustDuration(deltaMins int) {
	if m.Timer.Phase != state.PhaseIdle {
		return
	}
	newMins := m.FocusDurationMins + deltaMins
	if newMins < 1 {
		newMins = 1
	}
	m.FocusDurationMins = newMins
	m.TargetDuration = time.Duration(newMins) * time.Minute
}

func (m *Model) refreshStats() {
	if mins, count, err := m.DB.GetTodayStats(); err == nil {
		m.TodayFocusMins = mins
		m.TodaySessions = count
	}
	m.Streak = m.DB.GetStreak()
	if b, err := m.DB.GetCategoryBreakdownToday(); err == nil {
		m.CatBreakdown = b
	}
}

func (m *Model) SelectedCategory() *db.Category {
	if m.CatIdx < len(m.Categories) {
		return &m.Categories[m.CatIdx]
	}
	return nil
}

func FormatFocusTime(mins float64) string {
	h := int(mins) / 60
	min := int(mins) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, min)
	}
	return fmt.Sprintf("%dm", min)
}
