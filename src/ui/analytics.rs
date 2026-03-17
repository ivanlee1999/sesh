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

    let sections = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(1),
            Constraint::Length(3),   // summary header
            Constraint::Length(1),
            Constraint::Min(4),     // category breakdown
            Constraint::Length(1),
        ])
        .split(area);

    // Summary
    let hours = (app.today_focus_mins / 60.0) as u64;
    let mins = (app.today_focus_mins % 60.0) as u64;
    let time_str = if hours > 0 {
        format!("{}h {}m", hours, mins)
    } else {
        format!("{}m", mins)
    };

    let summary = vec![
        Line::from(Span::styled(
            "  Today's Summary",
            Style::default().fg(app.theme.fg).add_modifier(Modifier::BOLD),
        )),
        Line::from(Span::styled(
            format!("  Total Focus: {}  │  Sessions: {}  │  Streak: {} days",
                time_str, app.today_sessions, app.streak),
            Style::default().fg(app.theme.fg_secondary),
        )),
    ];
    let para = Paragraph::new(summary);
    frame.render_widget(para, sections[1]);

    // Category breakdown
    let mut cat_lines = vec![
        Line::from(Span::styled(
            "  Category Breakdown",
            Style::default().fg(app.theme.fg).add_modifier(Modifier::BOLD),
        )),
        Line::from(""),
    ];

    let total: f64 = app.category_breakdown.iter().map(|(_, _, m)| m).sum();

    for (name, hex, minutes) in &app.category_breakdown {
        let pct = if total > 0.0 { minutes / total * 100.0 } else { 0.0 };
        let bar_width = 20;
        let filled = ((pct / 100.0) * bar_width as f64) as usize;
        let empty = bar_width - filled;
        let color = parse_hex_color(hex);

        let hrs = (*minutes / 60.0) as u64;
        let mns = (*minutes % 60.0) as u64;
        let time = if hrs > 0 { format!("{}h {}m", hrs, mns) } else { format!("{}m", mns) };

        cat_lines.push(Line::from(vec![
            Span::raw("  "),
            Span::styled("█".repeat(filled), Style::default().fg(color)),
            Span::styled("░".repeat(empty), Style::default().fg(app.theme.progress_empty)),
            Span::styled(format!(" {:>3.0}% ", pct), Style::default().fg(app.theme.fg_secondary)),
            Span::styled(format!("{:<16}", name), Style::default().fg(app.theme.fg)),
            Span::styled(time, Style::default().fg(app.theme.fg_secondary)),
        ]));
    }

    if app.category_breakdown.is_empty() {
        cat_lines.push(Line::from(Span::styled(
            "  No sessions today. Start focusing!",
            Style::default().fg(app.theme.fg_secondary),
        )));
    }

    let para = Paragraph::new(cat_lines);
    frame.render_widget(para, sections[3]);
}
