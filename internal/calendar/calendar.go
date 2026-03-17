package calendar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivanlee1999/sesh/internal/db"
)

func SessionsToICS(sessions []db.SessionRecord) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\n")
	b.WriteString("VERSION:2.0\n")
	b.WriteString("PRODID:-//sesh//sesh Pomodoro Timer//EN\n")
	b.WriteString("CALSCALE:GREGORIAN\n")
	b.WriteString("METHOD:PUBLISH\n")
	b.WriteString("X-WR-CALNAME:sesh Focus Sessions\n")

	for _, s := range sessions {
		if s.SessionType == "abandoned" {
			continue
		}
		b.WriteString(sessionToVEvent(&s))
	}

	b.WriteString("END:VCALENDAR\n")
	return b.String()
}

func sessionToVEvent(s *db.SessionRecord) string {
	var b strings.Builder
	uid := s.ID + "@sesh"
	summary := "[sesh] Focus Session"
	if s.Title != "" {
		summary = "[sesh] " + s.Title
	}

	catName := "Uncategorized"
	if s.CategoryTitle != nil {
		catName = *s.CategoryTitle
	}
	durMins := s.ActualSeconds / 60
	durSecs := s.ActualSeconds % 60
	desc := fmt.Sprintf("Category: %s\\nDuration: %d:%02d\\nType: %s",
		catName, durMins, durSecs, s.SessionType)
	if s.OverflowSeconds > 0 {
		desc += fmt.Sprintf("\\nOverflow: +%d:%02d", s.OverflowSeconds/60, s.OverflowSeconds%60)
	}
	if s.PauseSeconds > 0 {
		desc += fmt.Sprintf("\\nPaused: %d:%02d", s.PauseSeconds/60, s.PauseSeconds%60)
	}
	if s.Notes != nil && *s.Notes != "" {
		desc += "\\nNotes: " + strings.ReplaceAll(*s.Notes, "\n", "\\n")
	}

	b.WriteString("BEGIN:VEVENT\n")
	b.WriteString("UID:" + uid + "\n")
	b.WriteString("DTSTART:" + datetimeToICS(s.StartedAt) + "\n")
	b.WriteString("DTEND:" + datetimeToICS(s.EndedAt) + "\n")
	b.WriteString("SUMMARY:" + icsEscape(summary) + "\n")
	b.WriteString("DESCRIPTION:" + desc + "\n")
	b.WriteString("CATEGORIES:" + catName + "\n")
	b.WriteString("STATUS:CONFIRMED\n")
	b.WriteString("TRANSP:OPAQUE\n")
	b.WriteString("END:VEVENT\n")
	return b.String()
}

func ExportICS(sessions []db.SessionRecord, path string) error {
	ics := SessionsToICS(sessions)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ics), 0644)
}

func AutoExportSession(s *db.SessionRecord, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	event := sessionToVEvent(s)

	data, err := os.ReadFile(path)
	if err == nil {
		content := string(data)
		if idx := strings.LastIndex(content, "END:VCALENDAR"); idx >= 0 {
			content = content[:idx] + event + "END:VCALENDAR\n"
			return os.WriteFile(path, []byte(content), 0644)
		}
	}

	// File doesn't exist or is malformed — create new
	sessions := []db.SessionRecord{*s}
	return ExportICS(sessions, path)
}

func datetimeToICS(dt string) string {
	r := strings.NewReplacer("-", "", ":", "", " ", "T")
	return r.Replace(dt)
}

func icsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", `\;`)
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
