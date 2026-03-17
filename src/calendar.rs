use crate::db::sessions::SessionRecord;
use std::fmt::Write;
use std::path::Path;

/// Generate an ICS calendar string from session records.
pub fn sessions_to_ics(sessions: &[SessionRecord]) -> String {
    let mut ics = String::new();
    writeln!(ics, "BEGIN:VCALENDAR").unwrap();
    writeln!(ics, "VERSION:2.0").unwrap();
    writeln!(ics, "PRODID:-//sesh//sesh Pomodoro Timer//EN").unwrap();
    writeln!(ics, "CALSCALE:GREGORIAN").unwrap();
    writeln!(ics, "METHOD:PUBLISH").unwrap();
    writeln!(ics, "X-WR-CALNAME:sesh Focus Sessions").unwrap();

    for session in sessions {
        if session.session_type == "abandoned" {
            continue;
        }

        let dtstart = datetime_to_ics(&session.started_at);
        let dtend = datetime_to_ics(&session.ended_at);
        let uid = format!("{}@sesh", session.id);

        let summary = if session.title.is_empty() {
            "[sesh] Focus Session".to_string()
        } else {
            format!("[sesh] {}", session.title)
        };

        let cat_name = session.category_title.as_deref().unwrap_or("Uncategorized");
        let dur_mins = session.actual_seconds / 60;
        let dur_secs = session.actual_seconds % 60;

        let mut description = format!(
            "Category: {}\\nDuration: {}:{:02}\\nType: {}",
            cat_name, dur_mins, dur_secs, session.session_type,
        );
        if session.overflow_seconds > 0 {
            let ov_mins = session.overflow_seconds / 60;
            let ov_secs = session.overflow_seconds % 60;
            write!(description, "\\nOverflow: +{}:{:02}", ov_mins, ov_secs).unwrap();
        }
        if session.pause_seconds > 0 {
            let p_mins = session.pause_seconds / 60;
            let p_secs = session.pause_seconds % 60;
            write!(description, "\\nPaused: {}:{:02}", p_mins, p_secs).unwrap();
        }
        if let Some(notes) = &session.notes {
            if !notes.is_empty() {
                write!(description, "\\nNotes: {}", notes.replace('\n', "\\n")).unwrap();
            }
        }

        writeln!(ics, "BEGIN:VEVENT").unwrap();
        writeln!(ics, "UID:{}", uid).unwrap();
        writeln!(ics, "DTSTART:{}", dtstart).unwrap();
        writeln!(ics, "DTEND:{}", dtend).unwrap();
        writeln!(ics, "SUMMARY:{}", ics_escape(&summary)).unwrap();
        writeln!(ics, "DESCRIPTION:{}", description).unwrap();
        writeln!(ics, "CATEGORIES:{}", cat_name).unwrap();
        writeln!(ics, "STATUS:CONFIRMED").unwrap();
        writeln!(ics, "TRANSP:OPAQUE").unwrap();
        writeln!(ics, "END:VEVENT").unwrap();
    }

    writeln!(ics, "END:VCALENDAR").unwrap();
    ics
}

/// Write sessions to an ICS file at the given path.
pub fn export_ics(sessions: &[SessionRecord], path: &Path) -> anyhow::Result<()> {
    let ics = sessions_to_ics(sessions);
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    std::fs::write(path, ics)?;
    Ok(())
}

/// Auto-export: append a single session event to the ICS file.
/// If the file doesn't exist, creates a new one. If it does, inserts before END:VCALENDAR.
pub fn auto_export_session(session: &SessionRecord, path: &Path) -> anyhow::Result<()> {
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    // Simpler approach: just regenerate the whole file from DB isn't ideal for auto-export.
    // Instead, we'll read the existing file, insert the event before the closing tag.
    if path.exists() {
        let mut content = std::fs::read_to_string(path)?;
        if let Some(pos) = content.rfind("END:VCALENDAR") {
            let event = session_to_vevent(session);
            content.insert_str(pos, &event);
            std::fs::write(path, content)?;
        } else {
            // Malformed file, recreate
            export_ics(&[session.clone()], path)?;
        }
    } else {
        export_ics(&[session.clone()], path)?;
    }
    Ok(())
}

fn session_to_vevent(session: &SessionRecord) -> String {
    let mut ics = String::new();
    let dtstart = datetime_to_ics(&session.started_at);
    let dtend = datetime_to_ics(&session.ended_at);
    let uid = format!("{}@sesh", session.id);

    let summary = if session.title.is_empty() {
        "[sesh] Focus Session".to_string()
    } else {
        format!("[sesh] {}", session.title)
    };

    let cat_name = session.category_title.as_deref().unwrap_or("Uncategorized");
    let dur_mins = session.actual_seconds / 60;
    let dur_secs = session.actual_seconds % 60;

    let mut description = format!(
        "Category: {}\\nDuration: {}:{:02}\\nType: {}",
        cat_name, dur_mins, dur_secs, session.session_type,
    );
    if session.overflow_seconds > 0 {
        write!(description, "\\nOverflow: +{}:{:02}", session.overflow_seconds / 60, session.overflow_seconds % 60).unwrap();
    }
    if session.pause_seconds > 0 {
        write!(description, "\\nPaused: {}:{:02}", session.pause_seconds / 60, session.pause_seconds % 60).unwrap();
    }
    if let Some(notes) = &session.notes {
        if !notes.is_empty() {
            write!(description, "\\nNotes: {}", notes.replace('\n', "\\n")).unwrap();
        }
    }

    writeln!(ics, "BEGIN:VEVENT").unwrap();
    writeln!(ics, "UID:{}", uid).unwrap();
    writeln!(ics, "DTSTART:{}", dtstart).unwrap();
    writeln!(ics, "DTEND:{}", dtend).unwrap();
    writeln!(ics, "SUMMARY:{}", ics_escape(&summary)).unwrap();
    writeln!(ics, "DESCRIPTION:{}", description).unwrap();
    writeln!(ics, "CATEGORIES:{}", cat_name).unwrap();
    writeln!(ics, "STATUS:CONFIRMED").unwrap();
    writeln!(ics, "TRANSP:OPAQUE").unwrap();
    writeln!(ics, "END:VEVENT").unwrap();
    ics
}

/// Convert "2026-03-17T14:32:00" to ICS format "20260317T143200"
fn datetime_to_ics(dt: &str) -> String {
    dt.replace('-', "").replace(':', "").replace(' ', "T")
}

fn ics_escape(s: &str) -> String {
    s.replace('\\', "\\\\")
        .replace(';', r"\;")
        .replace(',', "\\,")
        .replace('\n', "\\n")
}
