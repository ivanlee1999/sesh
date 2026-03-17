use super::{Database, DbResult};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct SessionRecord {
    pub id: String,
    pub title: String,
    pub category_id: Option<String>,
    pub category_title: Option<String>,
    pub category_color: Option<String>,
    pub session_type: String,
    pub target_seconds: i64,
    pub actual_seconds: i64,
    pub pause_seconds: i64,
    pub overflow_seconds: i64,
    pub started_at: String,
    pub ended_at: String,
    pub notes: Option<String>,
}

impl Database {
    pub fn save_session(
        &self,
        title: &str,
        category_id: Option<&str>,
        session_type: &str,
        target_seconds: i64,
        actual_seconds: i64,
        pause_seconds: i64,
        overflow_seconds: i64,
        started_at: &str,
        ended_at: &str,
        notes: Option<&str>,
    ) -> DbResult<String> {
        let id = Uuid::new_v4().to_string();
        self.conn.execute(
            "INSERT INTO sessions (id, title, category_id, session_type, target_seconds, actual_seconds, pause_seconds, overflow_seconds, started_at, ended_at, notes)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11)",
            rusqlite::params![
                id, title, category_id, session_type,
                target_seconds, actual_seconds, pause_seconds, overflow_seconds,
                started_at, ended_at, notes,
            ],
        )?;
        Ok(id)
    }

    pub fn get_sessions(&self, limit: usize) -> DbResult<Vec<SessionRecord>> {
        let mut stmt = self.conn.prepare(
            "SELECT s.id, s.title, s.category_id, c.title, c.hex_color, s.session_type,
                    s.target_seconds, s.actual_seconds, s.pause_seconds, s.overflow_seconds,
                    s.started_at, s.ended_at, s.notes
             FROM sessions s
             LEFT JOIN categories c ON s.category_id = c.id
             ORDER BY s.started_at DESC
             LIMIT ?1"
        )?;
        let sessions = stmt.query_map(rusqlite::params![limit as i64], |row| {
            Ok(SessionRecord {
                id: row.get(0)?,
                title: row.get(1)?,
                category_id: row.get(2)?,
                category_title: row.get(3)?,
                category_color: row.get(4)?,
                session_type: row.get(5)?,
                target_seconds: row.get(6)?,
                actual_seconds: row.get(7)?,
                pause_seconds: row.get(8)?,
                overflow_seconds: row.get(9)?,
                started_at: row.get(10)?,
                ended_at: row.get(11)?,
                notes: row.get(12)?,
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;
        Ok(sessions)
    }

    pub fn get_today_stats(&self) -> DbResult<(f64, i64)> {
        let focus_minutes: f64 = self.conn.query_row(
            "SELECT COALESCE(SUM(actual_seconds - pause_seconds), 0) / 60.0
             FROM sessions
             WHERE session_type IN ('full_focus', 'partial_focus')
               AND date(started_at) = date('now', 'localtime')",
            [],
            |row| row.get(0),
        )?;
        let session_count: i64 = self.conn.query_row(
            "SELECT COUNT(*)
             FROM sessions
             WHERE session_type IN ('full_focus', 'partial_focus')
               AND date(started_at) = date('now', 'localtime')",
            [],
            |row| row.get(0),
        )?;
        Ok((focus_minutes, session_count))
    }

    pub fn get_streak(&self) -> DbResult<i64> {
        let streak: i64 = self.conn.query_row(
            "WITH RECURSIVE dates AS (
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
            )",
            [],
            |row| row.get(0),
        ).unwrap_or(0);
        Ok(streak)
    }

    pub fn get_category_breakdown_today(&self) -> DbResult<Vec<(String, String, f64)>> {
        let mut stmt = self.conn.prepare(
            "SELECT COALESCE(c.title, 'Uncategorized'), COALESCE(c.hex_color, '#ABB2BF'),
                    SUM(s.actual_seconds - s.pause_seconds) / 60.0
             FROM sessions s
             LEFT JOIN categories c ON s.category_id = c.id
             WHERE s.session_type IN ('full_focus', 'partial_focus')
               AND date(s.started_at) = date('now', 'localtime')
             GROUP BY c.id
             ORDER BY 3 DESC"
        )?;
        let rows = stmt.query_map([], |row| {
            Ok((row.get::<_, String>(0)?, row.get::<_, String>(1)?, row.get::<_, f64>(2)?))
        })?
        .collect::<Result<Vec<_>, _>>()?;
        Ok(rows)
    }

    pub fn delete_session(&self, id: &str) -> DbResult<()> {
        self.conn.execute("DELETE FROM sessions WHERE id = ?1", [id])?;
        Ok(())
    }

    pub fn get_recent_intentions(&self, limit: usize) -> DbResult<Vec<String>> {
        let mut stmt = self.conn.prepare(
            "SELECT DISTINCT title FROM sessions
             WHERE title != '' ORDER BY started_at DESC LIMIT ?1"
        )?;
        let intentions = stmt.query_map(rusqlite::params![limit as i64], |row| {
            row.get(0)
        })?
        .collect::<Result<Vec<String>, _>>()?;
        Ok(intentions)
    }
}
