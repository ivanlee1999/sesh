package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	DB *sql.DB
}

func Open(path string) (*Database, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	sqldb, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, err
	}
	d := &Database{DB: sqldb}
	if err := d.migrate(); err != nil {
		return nil, err
	}
	return d, nil
}

func OpenDefault() (*Database, error) {
	dataDir := defaultDataDir()
	return Open(filepath.Join(dataDir, "sessions.db"))
}

func defaultDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "sesh")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "sesh")
}

func (d *Database) migrate() error {
	_, err := d.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return err
	}

	var current int
	row := d.DB.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err := row.Scan(&current); err != nil {
		return err
	}

	if current < 1 {
		if _, err := d.DB.Exec(migration001); err != nil {
			return fmt.Errorf("migration 001: %w", err)
		}
		if _, err := d.DB.Exec("INSERT INTO schema_version (version) VALUES (1)"); err != nil {
			return err
		}
	}
	return nil
}

const migration001 = `
CREATE TABLE IF NOT EXISTS categories (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    hex_color   TEXT NOT NULL DEFAULT '#61AFEF',
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL DEFAULT '',
    category_id     TEXT REFERENCES categories(id) ON DELETE SET NULL,
    session_type    TEXT NOT NULL CHECK (session_type IN (
                        'full_focus', 'partial_focus', 'rest', 'abandoned'
                    )),
    target_seconds  INTEGER NOT NULL,
    actual_seconds  INTEGER NOT NULL,
    pause_seconds   INTEGER NOT NULL DEFAULT 0,
    overflow_seconds INTEGER NOT NULL DEFAULT 0,
    started_at      TEXT NOT NULL,
    ended_at        TEXT NOT NULL,
    notes           TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS pauses (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    paused_at   TEXT NOT NULL,
    resumed_at  TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_sessions_category   ON sessions(category_id);
CREATE INDEX IF NOT EXISTS idx_sessions_type       ON sessions(session_type);
CREATE INDEX IF NOT EXISTS idx_pauses_session      ON pauses(session_id);
`
