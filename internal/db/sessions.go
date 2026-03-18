package db

import (
	"github.com/google/uuid"
)

type SessionRecord struct {
	ID              string
	Title           string
	CategoryID      *string
	CategoryTitle   *string
	CategoryColor   *string
	SessionType     string
	TargetSeconds   int64
	ActualSeconds   int64
	PauseSeconds    int64
	OverflowSeconds int64
	StartedAt       string
	EndedAt         string
	Notes           *string
}

func (d *Database) SaveSession(
	title string,
	categoryID *string,
	sessionType string,
	targetSecs, actualSecs, pauseSecs, overflowSecs int64,
	startedAt, endedAt string,
	notes *string,
) (string, error) {
	id := uuid.New().String()
	_, err := d.DB.Exec(
		`INSERT INTO sessions (id, title, category_id, session_type, target_seconds, actual_seconds, pause_seconds, overflow_seconds, started_at, ended_at, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, title, categoryID, sessionType,
		targetSecs, actualSecs, pauseSecs, overflowSecs,
		startedAt, endedAt, notes,
	)
	return id, err
}

func (d *Database) GetSessions(limit int) ([]SessionRecord, error) {
	rows, err := d.DB.Query(
		`SELECT s.id, s.title, s.category_id, c.title, c.hex_color, s.session_type,
		        s.target_seconds, s.actual_seconds, s.pause_seconds, s.overflow_seconds,
		        s.started_at, s.ended_at, s.notes
		 FROM sessions s
		 LEFT JOIN categories c ON s.category_id = c.id
		 ORDER BY s.started_at DESC
		 LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionRecord
	for rows.Next() {
		var s SessionRecord
		if err := rows.Scan(
			&s.ID, &s.Title, &s.CategoryID, &s.CategoryTitle, &s.CategoryColor,
			&s.SessionType, &s.TargetSeconds, &s.ActualSeconds, &s.PauseSeconds,
			&s.OverflowSeconds, &s.StartedAt, &s.EndedAt, &s.Notes,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (d *Database) GetTodayStats() (focusMins float64, sessionCount int64, err error) {
	err = d.DB.QueryRow(
		`SELECT COALESCE(SUM(actual_seconds - pause_seconds), 0) / 60.0
		 FROM sessions
		 WHERE session_type IN ('full_focus', 'partial_focus')
		   AND date(started_at) = date('now', 'localtime')`,
	).Scan(&focusMins)
	if err != nil {
		return
	}
	err = d.DB.QueryRow(
		`SELECT COUNT(*)
		 FROM sessions
		 WHERE session_type IN ('full_focus', 'partial_focus')
		   AND date(started_at) = date('now', 'localtime')`,
	).Scan(&sessionCount)
	return
}

func (d *Database) GetStreak() int64 {
	var streak int64
	_ = d.DB.QueryRow(
		`WITH RECURSIVE dates AS (
			SELECT date('now', 'localtime') AS d
			UNION ALL
			SELECT date(d, '-1 day') FROM dates
			WHERE EXISTS (
				SELECT 1 FROM sessions
				WHERE session_type IN ('full_focus', 'partial_focus')
				  AND date(started_at, 'localtime') = date(d, '-1 day')
			)
		)
		SELECT COUNT(*) - 1 FROM dates
		WHERE EXISTS (
			SELECT 1 FROM sessions
			WHERE session_type IN ('full_focus', 'partial_focus')
			  AND date(started_at, 'localtime') = d
		)`,
	).Scan(&streak)
	if streak < 0 {
		streak = 0
	}
	return streak
}

type CategoryBreakdown struct {
	Name    string
	Color   string
	Minutes float64
}

func (d *Database) GetCategoryBreakdownToday() ([]CategoryBreakdown, error) {
	rows, err := d.DB.Query(
		`SELECT COALESCE(c.title, 'Uncategorized'), COALESCE(c.hex_color, '#ABB2BF'),
		        SUM(s.actual_seconds - s.pause_seconds) / 60.0
		 FROM sessions s
		 LEFT JOIN categories c ON s.category_id = c.id
		 WHERE s.session_type IN ('full_focus', 'partial_focus')
		   AND date(s.started_at) = date('now', 'localtime')
		 GROUP BY c.id
		 ORDER BY 3 DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CategoryBreakdown
	for rows.Next() {
		var b CategoryBreakdown
		if err := rows.Scan(&b.Name, &b.Color, &b.Minutes); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

func (d *Database) DeleteSession(id string) error {
	_, err := d.DB.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

type DayFocus struct {
	Date  string
	Hours float64
}

func (d *Database) GetLast7DaysFocus() ([]DayFocus, error) {
	rows, err := d.DB.Query(
		`WITH RECURSIVE cnt(n) AS (
			SELECT 0 UNION ALL SELECT n+1 FROM cnt WHERE n < 6
		)
		SELECT date('now', 'localtime', '-' || (6-n) || ' days') AS day,
		       COALESCE(SUM(s.actual_seconds - s.pause_seconds), 0) / 3600.0 AS hours
		FROM cnt
		LEFT JOIN sessions s
			ON date(s.started_at, 'localtime') = date('now', 'localtime', '-' || (6-n) || ' days')
			AND s.session_type IN ('full_focus', 'partial_focus')
		GROUP BY n
		ORDER BY n ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DayFocus
	for rows.Next() {
		var f DayFocus
		if err := rows.Scan(&f.Date, &f.Hours); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func (d *Database) GetTotalFocusAllTime() float64 {
	var mins float64
	_ = d.DB.QueryRow(
		`SELECT COALESCE(SUM(actual_seconds - pause_seconds), 0) / 60.0
		 FROM sessions
		 WHERE session_type IN ('full_focus', 'partial_focus')`,
	).Scan(&mins)
	return mins
}

// GetTodaySessions returns today's focus sessions ordered chronologically (ASC).
func (d *Database) GetTodaySessions() ([]SessionRecord, error) {
	rows, err := d.DB.Query(
		`SELECT s.id, s.title, s.category_id, c.title, c.hex_color, s.session_type,
		        s.target_seconds, s.actual_seconds, s.pause_seconds, s.overflow_seconds,
		        s.started_at, s.ended_at, s.notes
		 FROM sessions s
		 LEFT JOIN categories c ON s.category_id = c.id
		 WHERE s.session_type IN ('full_focus', 'partial_focus')
		   AND date(s.started_at) = date('now', 'localtime')
		 ORDER BY s.started_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionRecord
	for rows.Next() {
		var s SessionRecord
		if err := rows.Scan(
			&s.ID, &s.Title, &s.CategoryID, &s.CategoryTitle, &s.CategoryColor,
			&s.SessionType, &s.TargetSeconds, &s.ActualSeconds, &s.PauseSeconds,
			&s.OverflowSeconds, &s.StartedAt, &s.EndedAt, &s.Notes,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
