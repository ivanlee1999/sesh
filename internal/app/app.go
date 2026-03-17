package app

import (
	"fmt"
	"time"

	"github.com/ivanlee1999/sesh/internal/calendar"
	"github.com/ivanlee1999/sesh/internal/calsync"
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
	ModeSessionComplete // notes input after timer hits 0
	ModeSessionPost     // after saving: b/enter/q
	ModeHelp            // keybinding help overlay
)

type tickMsg time.Time
type calSyncDoneMsg struct{}

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
	Intention       string
	IntentionDraft  string
	Categories      []db.Category
	CatIdx          int
	CatIdxDraft     int
	CatScrollOffset int

	// History
	HistorySelected    int
	HistoryScrollOffset int
	HistorySessions    []db.SessionRecord

	// Analytics
	TodayFocusMins float64
	TodaySessions  int64
	Streak         int64
	CatBreakdown   []db.CategoryBreakdown
	Last7Days      []db.DayFocus
	TotalFocusMins float64

	// Session completion
	CompletionNotes    string
	CompletionDuration time.Duration
	CompletedAt        time.Time

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
	last7, _ := database.GetLast7DaysFocus()
	historySessions, _ := database.GetSessions(200)

	return Model{
		Timer:             state.NewIdle(),
		Screen:            ScreenTimer,
		InputMode:         ModeNormal,
		Config:            cfg,
		FocusDurationMins: cfg.Timer.FocusDuration,
		TargetDuration:    time.Duration(cfg.Timer.FocusDuration) * time.Minute,
		Categories:        cats,
		TodayFocusMins:    focusMins,
		TodaySessions:     sessions,
		Streak:            streak,
		CatBreakdown:      breakdown,
		Last7Days:         last7,
		TotalFocusMins:    database.GetTotalFocusAllTime(),
		HistorySessions:   historySessions,
		DB:                database,
		Width:             80,
		Height:            24,
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

	case calSyncDoneMsg:
		// Calendar sync completed in background; nothing to do.
		return m, nil
	}
	return m, nil
}

func (m *Model) tick() {
	if m.InputMode == ModeSessionComplete || m.InputMode == ModeSessionPost {
		return
	}
	d := time.Duration(m.Config.General.TickRateMs) * time.Millisecond
	switch m.Timer.Phase {
	case state.PhaseFocus:
		if m.Timer.Remaining <= d {
			fmt.Print("\a")
			m.Timer.Phase = state.PhaseOverflow
			m.Timer.Elapsed = 0
			m.Timer.TargetWas = m.Timer.Target
			m.Timer.Remaining = 0
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
	// Help overlay swallows all keys
	if m.InputMode == ModeHelp {
		if msg.String() == "esc" || msg.String() == "?" {
			m.InputMode = ModeNormal
		}
		return m, nil
	}

	// Input modes
	if m.InputMode == ModeIntention {
		return m.handleIntentionKey(msg)
	}
	if m.InputMode == ModeCategory {
		return m.handleCategoryKey(msg)
	}
	if m.InputMode == ModeSessionComplete {
		return m.handleSessionCompleteKey(msg)
	}
	if m.InputMode == ModeSessionPost {
		return m.handleSessionPostKey(msg)
	}

	key := msg.String()

	// Ctrl+C always quits
	if key == "ctrl+c" {
		m.Quitting = true
		return m, tea.Quit
	}

	// Help overlay
	if key == "?" {
		m.InputMode = ModeHelp
		return m, nil
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
		m.refreshHistory()
		m.Screen = ScreenHistory
	case "4":
		m.Screen = ScreenSettings
	case "tab":
		m.Screen = (m.Screen + 1) % 4
		if m.Screen == ScreenAnalytics {
			m.refreshStats()
		}
		if m.Screen == ScreenHistory {
			m.refreshHistory()
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
		m.Intention = m.IntentionDraft
		m.InputMode = ModeNormal
	case "enter":
		m.InputMode = ModeNormal
	case "backspace":
		if len(m.Intention) > 0 {
			// Remove last rune (not byte) for correct unicode handling
			runes := []rune(m.Intention)
			m.Intention = string(runes[:len(runes)-1])
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.Intention += string(msg.Runes)
		}
	}
	return m, nil
}

const CatMaxVisible = 8

func (m Model) handleCategoryKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.CatIdx = m.CatIdxDraft
		m.InputMode = ModeNormal
	case "enter":
		m.InputMode = ModeNormal
	case "up", "k":
		if m.CatIdx > 0 {
			m.CatIdx--
			if m.CatIdx < m.CatScrollOffset {
				m.CatScrollOffset = m.CatIdx
			}
		}
	case "down", "j":
		if m.CatIdx < len(m.Categories)-1 {
			m.CatIdx++
			if m.CatIdx >= m.CatScrollOffset+CatMaxVisible {
				m.CatScrollOffset = m.CatIdx - CatMaxVisible + 1
			}
		}
	}
	return m, nil
}

func clampScrollOffset(idx, offset, total int) int {
	if idx < offset {
		return idx
	}
	if idx >= offset+CatMaxVisible {
		return idx - CatMaxVisible + 1
	}
	return offset
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
			m.IntentionDraft = m.Intention
			m.InputMode = ModeIntention
		case "c":
			m.CatIdxDraft = m.CatIdx
			m.CatScrollOffset = clampScrollOffset(m.CatIdx, m.CatScrollOffset, len(m.Categories))
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
	case state.PhaseFocus:
		switch key {
		case " ":
			m.togglePause()
		case "f":
			cmd := m.finishSession()
			return m, cmd
		case "x":
			m.abandonSession()
		case "b":
			cmd := m.finishSession()
			m.startBreak(state.BreakShort)
			return m, cmd
		}
	case state.PhaseOverflow:
		switch key {
		case " ":
			m.togglePause()
		case "f":
			m.triggerSessionComplete()
		case "x":
			m.abandonSession()
		case "b":
			cmd := m.finishSession()
			m.startBreak(state.BreakShort)
			return m, cmd
		}
	case state.PhasePaused:
		switch key {
		case " ":
			m.togglePause()
		case "f":
			cmd := m.finishSession()
			return m, cmd
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
	n := len(m.HistorySessions)
	visibleRows := m.Height - 8
	if visibleRows < 5 {
		visibleRows = 5
	}
	switch msg.String() {
	case "up", "k":
		if m.HistorySelected > 0 {
			m.HistorySelected--
			if m.HistorySelected < m.HistoryScrollOffset {
				m.HistoryScrollOffset = m.HistorySelected
			}
		}
	case "down", "j":
		if m.HistorySelected < n-1 {
			m.HistorySelected++
			if m.HistorySelected >= m.HistoryScrollOffset+visibleRows {
				m.HistoryScrollOffset = m.HistorySelected - visibleRows + 1
			}
		}
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

func (m *Model) finishSession() tea.Cmd {
	if m.StartedAtChrono.IsZero() {
		m.Timer = state.NewIdle()
		return nil
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

	rec := db.SessionRecord{
		ID: "auto", Title: m.Intention,
		CategoryID: catID, CategoryTitle: catTitle, CategoryColor: catColor,
		SessionType: sessionType, TargetSeconds: targetSecs,
		ActualSeconds: actualSecs, PauseSeconds: pauseSecs,
		OverflowSeconds: overflowSecs, StartedAt: startedStr, EndedAt: endedStr,
	}

	// Auto-export to ICS
	if m.Config.Calendar.Enabled && m.Config.Calendar.AutoExport {
		calendar.AutoExportSession(&rec, m.Config.Calendar.ICSPath)
	}

	if actualSecs > 0 {
		m.CumulativeFocus += time.Duration(actualSecs) * time.Second
	}
	m.refreshStats()
	m.Timer = state.NewIdle()
	m.StartedAtChrono = time.Time{}

	return m.calSyncCmd(rec)
}

func (m *Model) triggerSessionComplete() {
	now := time.Now()
	totalElapsed := now.Sub(m.StartedAtChrono)
	actual := totalElapsed - m.PauseAccum
	if actual < 0 {
		actual = 0
	}
	m.CompletionDuration = actual
	m.CompletedAt = now
	m.CompletionNotes = ""
	m.InputMode = ModeSessionComplete
}

func (m Model) handleSessionCompleteKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "esc":
		m.InputMode = ModeNormal
		m.Timer = state.NewIdle()
		m.StartedAtChrono = time.Time{}
	case "enter":
		cmd := m.saveCompletedSession()
		m.InputMode = ModeSessionPost
		return m, cmd
	case "backspace":
		if len(m.CompletionNotes) > 0 {
			runes := []rune(m.CompletionNotes)
			m.CompletionNotes = string(runes[:len(runes)-1])
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.CompletionNotes += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleSessionPostKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "b":
		m.InputMode = ModeNormal
		m.StartedAtChrono = time.Time{}
		m.startBreak(state.BreakShort)
	case "enter":
		m.InputMode = ModeNormal
		m.Timer = state.NewIdle()
		m.StartedAtChrono = time.Time{}
	case "q":
		m.Quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) saveCompletedSession() tea.Cmd {
	actualSecs := int64(m.CompletionDuration.Seconds())
	targetSecs := int64(m.TargetDuration.Seconds())
	pauseSecs := int64(m.PauseAccum.Seconds())

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
	endedStr := m.CompletedAt.Format("2006-01-02T15:04:05")

	var notesPtr *string
	if m.CompletionNotes != "" {
		notesPtr = &m.CompletionNotes
	}

	m.DB.SaveSession(m.Intention, catID, sessionType, targetSecs, actualSecs, pauseSecs, 0, startedStr, endedStr, notesPtr)

	rec := db.SessionRecord{
		ID: "auto", Title: m.Intention,
		CategoryID: catID, CategoryTitle: catTitle, CategoryColor: catColor,
		SessionType: sessionType, TargetSeconds: targetSecs,
		ActualSeconds: actualSecs, PauseSeconds: pauseSecs,
		StartedAt: startedStr, EndedAt: endedStr, Notes: notesPtr,
	}

	if m.Config.Calendar.Enabled && m.Config.Calendar.AutoExport {
		calendar.AutoExportSession(&rec, m.Config.Calendar.ICSPath)
	}

	if actualSecs > 0 {
		m.CumulativeFocus += time.Duration(actualSecs) * time.Second
	}
	m.refreshStats()
	m.refreshHistory()

	return m.calSyncCmd(rec)
}

func (m *Model) calSyncCmd(rec db.SessionRecord) tea.Cmd {
	if !m.Config.Calendar.Google.Enabled && !m.Config.Calendar.Outlook.Enabled {
		return nil
	}
	cfg := m.Config.Calendar
	return func() tea.Msg {
		calsync.SyncSession(cfg, &rec)
		return calSyncDoneMsg{}
	}
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
	if days, err := m.DB.GetLast7DaysFocus(); err == nil {
		m.Last7Days = days
	}
	m.TotalFocusMins = m.DB.GetTotalFocusAllTime()
}

func (m *Model) refreshHistory() {
	sessions, _ := m.DB.GetSessions(200)
	m.HistorySessions = sessions
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
