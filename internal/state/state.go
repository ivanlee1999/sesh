package state

import (
	"fmt"
	"time"
)

type TimerPhase int

const (
	PhaseIdle TimerPhase = iota
	PhaseFocus
	PhaseOverflow
	PhasePaused
	PhaseBreak
	PhaseBreakOverflow
	PhaseAbandoned
)

func (p TimerPhase) String() string {
	switch p {
	case PhaseIdle:
		return "IDLE"
	case PhaseFocus:
		return "FOCUS"
	case PhaseOverflow:
		return "OVERFLOW"
	case PhasePaused:
		return "PAUSED"
	case PhaseBreak:
		return "BREAK"
	case PhaseBreakOverflow:
		return "BREAK OVERFLOW"
	case PhaseAbandoned:
		return "ABANDONED"
	}
	return "UNKNOWN"
}

type BreakType int

const (
	BreakShort BreakType = iota
	BreakLong
)

func (b BreakType) String() string {
	if b == BreakLong {
		return "Long"
	}
	return "Short"
}

type TimerState struct {
	Phase     TimerPhase
	Remaining time.Duration
	Elapsed   time.Duration
	Target    time.Duration
	TargetWas time.Duration // for overflow: what the original target was
	StartedAt time.Time
	PausedAt  time.Time
	BreakType BreakType

	// For paused state: remember what phase we were in
	PausedPhase  TimerPhase
	TotalPaused  time.Duration

	// For abandoned state
	PreviousState *TimerState
	UndoDeadline  time.Time
}

func NewIdle() TimerState {
	return TimerState{Phase: PhaseIdle}
}

func (s *TimerState) IsActive() bool {
	return s.Phase != PhaseIdle
}

func (s *TimerState) DisplayTime() string {
	switch s.Phase {
	case PhaseIdle:
		return ""
	case PhaseFocus, PhaseBreak:
		return FormatDuration(s.Remaining)
	case PhaseOverflow, PhaseBreakOverflow:
		return "+" + FormatDuration(s.Elapsed)
	case PhasePaused:
		if s.PausedPhase == PhaseOverflow {
			return "+" + FormatDuration(s.Elapsed)
		}
		return FormatDuration(s.Remaining)
	case PhaseAbandoned:
		return ""
	}
	return ""
}

func (s *TimerState) Progress() float64 {
	switch s.Phase {
	case PhaseFocus:
		if s.Target == 0 {
			return 1.0
		}
		elapsed := s.Target - s.Remaining
		return float64(elapsed) / float64(s.Target)
	case PhaseOverflow:
		if s.TargetWas == 0 {
			return 1.0
		}
		return 1.0 + float64(s.Elapsed)/float64(s.TargetWas)
	case PhaseBreak:
		if s.Target == 0 {
			return 1.0
		}
		elapsed := s.Target - s.Remaining
		return float64(elapsed) / float64(s.Target)
	case PhasePaused:
		if s.PausedPhase == PhaseFocus {
			if s.Target == 0 {
				return 1.0
			}
			elapsed := s.Target - s.Remaining
			return float64(elapsed) / float64(s.Target)
		}
		return 0
	}
	return 0
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	mins := total / 60
	secs := total % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
