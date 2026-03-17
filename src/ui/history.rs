use crate::app::App;
use crate::ui::widgets;
use crate::ui::timer::parse_hex_color;
use ratatui::layout::{Constraint, Direction, Layout, Rect};
use ratatui::style::{Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Paragraph};
use ratatui::Frame;

pub fn render(frame: &mut Frame, app: &App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),
            Constraint::Min(0),
            Constraint::Length(3),
        ])
        .split(frame.area());

    widgets::tab_bar::render(frame, app, chunks[0]);
    render_content(frame, app, chunks[1]);
    widgets::status_bar::render(frame, app, chunks[2]);
}

fn render_content(frame: &mut Frame, app: &App, area: Rect) {
    let block = Block::default()
        .borders(Borders::NONE)
        .style(Style::default().bg(app.theme.bg));
    frame.render_widget(block, area);

    let sessions = app.db.get_sessions(50).unwrap_or_default();

    if sessions.is_empty() {
        let para = Paragraph::new(vec![
            Line::from(""),
            Line::from(Span::styled(
                "  No sessions yet. Start focusing!",
                Style::default().fg(app.theme.fg_secondary),
            )),
        ]);
        frame.render_widget(para, area);
        return;
    }

    let mut lines = vec![
        Line::from(Span::styled(
            "  Session History",
            Style::default().fg(app.theme.fg).add_modifier(Modifier::BOLD),
        )),
        Line::from(""),
    ];

    let mut current_date = String::new();

    for (i, session) in sessions.iter().enumerate() {
        // Date grouping
        let date = session.started_at.split('T').next().unwrap_or(&session.started_at);
        if date != current_date {
            current_date = date.to_string();
            lines.push(Line::from(""));
            lines.push(Line::from(Span::styled(
                format!("  ── {} ──", current_date),
                Style::default().fg(app.theme.fg_secondary).add_modifier(Modifier::BOLD),
            )));
            lines.push(Line::from(""));
        }

        let start_time = session.started_at.split('T').nth(1)
            .unwrap_or("")
            .get(..5)
            .unwrap_or("??:??");
        let end_time = session.ended_at.split('T').nth(1)
            .unwrap_or("")
            .get(..5)
            .unwrap_or("??:??");

        let dur_mins = session.actual_seconds / 60;
        let dur_secs = session.actual_seconds % 60;
        let dur_str = format!("{}:{:02}", dur_mins, dur_secs);

        let cat_name = session.category_title.as_deref().unwrap_or("—");
        let cat_color = session.category_color.as_deref()
            .map(parse_hex_color)
            .unwrap_or(app.theme.fg_secondary);

        let marker = if i == app.history_selected { "> " } else { "  " };
        let type_icon = match session.session_type.as_str() {
            "full_focus" => "●",
            "partial_focus" => "◐",
            "rest" => "◯",
            _ => "✕",
        };

        lines.push(Line::from(vec![
            Span::styled(marker, Style::default().fg(app.theme.accent)),
            Span::styled(format!("{} ", type_icon), Style::default().fg(cat_color)),
            Span::styled(format!("{} - {}  ", start_time, end_time), Style::default().fg(app.theme.fg_secondary)),
            Span::styled(
                format!("{:<28}", if session.title.is_empty() { "(no intention)" } else { &session.title }),
                Style::default().fg(app.theme.fg),
            ),
            Span::styled(format!("{:<14}", cat_name), Style::default().fg(cat_color)),
            Span::styled(dur_str, Style::default().fg(app.theme.fg_secondary)),
        ]));
    }

    let visible_height = area.height as usize;
    let offset = app.history_offset.min(lines.len().saturating_sub(visible_height));
    let visible_lines: Vec<Line> = lines.into_iter().skip(offset).take(visible_height).collect();

    let para = Paragraph::new(visible_lines);
    frame.render_widget(para, area);
}
