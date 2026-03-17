package calsync

import (
	"log"

	"github.com/ivanlee1999/sesh/internal/config"
	"github.com/ivanlee1999/sesh/internal/db"
)

// SyncSession sends the completed session to all enabled calendar providers.
// Errors are logged but not propagated (fire-and-forget).
func SyncSession(cfg config.CalendarConfig, session *db.SessionRecord) {
	if cfg.Google.Enabled {
		p := NewGoogle(cfg.Google)
		if err := p.CreateEvent(session); err != nil {
			log.Printf("google calendar sync: %v", err)
		}
	}
	if cfg.Outlook.Enabled {
		p := NewOutlook(cfg.Outlook)
		if err := p.CreateEvent(session); err != nil {
			log.Printf("outlook calendar sync: %v", err)
		}
	}
}
