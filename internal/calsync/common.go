package calsync

import (
	"fmt"
	"time"

	"github.com/ivanlee1999/sesh/internal/db"
)

// Provider is implemented by each calendar sync backend.
type Provider interface {
	Name() string
	CreateEvent(session *db.SessionRecord) error
}

// SessionSummary returns the event title for a session.
func SessionSummary(s *db.SessionRecord) string {
	if s.Title != "" {
		return "Focus: " + s.Title
	}
	return "Focus Session"
}

// SessionDescription builds a text description from session details.
func SessionDescription(s *db.SessionRecord) string {
	desc := ""

	if s.CategoryTitle != nil && *s.CategoryTitle != "" {
		desc += fmt.Sprintf("Category: %s\n", *s.CategoryTitle)
	}

	actualDur := time.Duration(s.ActualSeconds) * time.Second
	targetDur := time.Duration(s.TargetSeconds) * time.Second
	desc += fmt.Sprintf("Duration: %s (target: %s)\n", formatDuration(actualDur), formatDuration(targetDur))
	desc += fmt.Sprintf("Type: %s\n", s.SessionType)

	if s.OverflowSeconds > 0 {
		desc += fmt.Sprintf("Overflow: %s\n", formatDuration(time.Duration(s.OverflowSeconds)*time.Second))
	}
	if s.PauseSeconds > 0 {
		desc += fmt.Sprintf("Paused: %s\n", formatDuration(time.Duration(s.PauseSeconds)*time.Second))
	}
	if s.Notes != nil && *s.Notes != "" {
		desc += fmt.Sprintf("\nNotes: %s\n", *s.Notes)
	}

	desc += "\nCreated by sesh"
	return desc
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 && s > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", s)
}
