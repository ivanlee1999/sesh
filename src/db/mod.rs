pub mod categories;
pub mod sessions;

use rusqlite::Connection;
use std::path::Path;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum DbError {
    #[error("SQLite error: {0}")]
    Sqlite(#[from] rusqlite::Error),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
}

pub type DbResult<T> = Result<T, DbError>;

pub struct Database {
    pub conn: Connection,
}

impl Database {
    pub fn open(path: &Path) -> DbResult<Self> {
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent)?;
        }
        let conn = Connection::open(path)?;
        conn.execute_batch("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;")?;
        let mut db = Self { conn };
        db.run_migrations()?;
        Ok(db)
    }

    pub fn open_default() -> DbResult<Self> {
        let data_dir = crate::config::Config::data_dir();
        Self::open(&data_dir.join("sessions.db"))
    }

    fn run_migrations(&mut self) -> DbResult<()> {
        self.conn.execute_batch(
            "CREATE TABLE IF NOT EXISTS schema_version (
                version     INTEGER NOT NULL,
                applied_at  TEXT    NOT NULL DEFAULT (datetime('now'))
            );"
        )?;

        let current: i64 = self.conn
            .query_row(
                "SELECT COALESCE(MAX(version), 0) FROM schema_version",
                [],
                |row| row.get(0),
            )?;

        if current < 1 {
            self.conn.execute_batch(MIGRATION_001)?;
            self.conn.execute(
                "INSERT INTO schema_version (version) VALUES (1)",
                [],
            )?;
        }

        Ok(())
    }
}

const MIGRATION_001: &str = "
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
";
