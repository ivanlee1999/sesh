use crate::app::{App, InputMode};
use crate::state::{format_duration, TimerState};
use crate::ui::widgets;
use ratatui::layout::{Alignment, Constraint, Direction, Layout, Rect};
use ratatui::style::{Color, Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Clear, Paragraph, Gauge};
use ratatui::Frame;

pub fn render(frame: &mut Frame, app: &App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),   // tab bar
            Constraint::Min(0),      // main content
            Constraint::Length(3),   // status bar
        ])
        .split(frame.area());

    widgets::tab_bar::render(frame, app, chunks[0]);
    render_timer_content(frame, app, chunks[1]);
    widgets::status_bar::render(frame, app, chunks[2]);

    // Overlays
    match app.input_mode {
        InputMode::Intention => render_intention_overlay(frame, app),
        InputMode::CategoryPicker => render_category_overlay(frame, app),
        InputMode::Normal => {}
    }
}

fn render_timer_content(frame: &mut Frame, app: &App, area: Rect) {
    let block = Block::default()
        .borders(Borders::NONE)
        .style(Style::default().bg(app.theme.bg));
    frame.render_widget(block, area);

    let inner = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(1),   // spacing
            Constraint::Length(13),  // clock circle
            Constraint::Length(1),   // spacing
            Constraint::Length(2),   // intention + category
            Constraint::Length(1),   // spacing
            Constraint::Length(1),   // progress bar
            Constraint::Length(1),   // spacing
            Constraint::Length(2),   // session info
            Constraint::Length(1),   // spacing
            Constraint::Length(2),   // today stats
            Constraint::Min(0),      // buttons/help
        ])
        .split(area);

    // Clock circle
    widgets::clock::render(frame, app, inner[1]);

    // Intention & category
    render_intention_display(frame, app, inner[3]);

    // Progress bar (when active)
    if app.timer_state.is_active() || app.timer_state.is_paused() {
        render_progress(frame, app, inner[5]);
    }

    // Session info
    render_session_info(frame, app, inner[7]);

    // Today stats
    render_today_stats(frame, app, inner[9]);

    // Idle buttons/help
    if matches!(app.timer_state, TimerState::Idle) {
        render_idle_buttons(frame, app, inner[10]);
    }
}

fn render_intention_display(frame: &mut Frame, app: &App, area: Rect) {
    let mut lines = Vec::new();

    if !app.intention.is_empty() || app.timer_state.is_active() || app.timer_state.is_paused() {
        let intent_text = if app.intention.is_empty() {
            "(no intention set)"
        } else {
            &app.intention
        };
        lines.push(Line::from(vec![
            Span::styled("  \u{25B8} ", Style::default().fg(app.theme.accent)),
            Span::styled(intent_text, Style::default().fg(app.theme.fg)),
        ]));
    }

    if let Some(cat) = app.selected_category() {
        let color = parse_hex_color(&cat.hex_color);
        lines.push(Line::from(vec![
            Span::raw("    "),
            Span::styled(&cat.title, Style::default().fg(color)),
        ]));
    }

    let para = Paragraph::new(lines).alignment(Alignment::Center);
    frame.render_widget(para, area);
}

fn render_progress(frame: &mut Frame, app: &App, area: Rect) {
    let progress = app.timer_state.progress().clamp(0.0, 1.0);
    let color = match &app.timer_state {
        TimerState::Focus { .. } => app.theme.focus_accent,
        TimerState::Overflow { .. } => app.theme.overflow_accent,
        TimerState::Break { .. } => app.theme.break_accent,
        TimerState::BreakOverflow { .. } => app.theme.break_accent,
        TimerState::Paused { inner, .. } => match inner.as_ref() {
            TimerState::Focus { .. } => app.theme.paused_fg,
            TimerState::Overflow { .. } => app.theme.overflow_accent,
            _ => app.theme.paused_fg,
        },
        _ => app.theme.fg_secondary,
    };

    // Centered progress bar
    let bar_width = area.width.min(40);
    let bar_area = Rect {
        x: area.x + (area.width.saturating_sub(bar_width)) / 2,
        y: area.y,
        width: bar_width,
        height: 1,
    };

    let pct = (progress * 100.0) as u16;
    let label = if matches!(app.timer_state, TimerState::Overflow { .. }) {
        format!("+{}", app.timer_state.display_time())
    } else {
        format!("{}%", pct)
    };

    let gauge = Gauge::default()
        .gauge_style(Style::default().fg(color).bg(app.theme.progress_empty))
        .ratio(progress)
        .label(label);
    frame.render_widget(gauge, bar_area);
}

fn render_session_info(frame: &mut Frame, app: &App, area: Rect) {
    let lines = match &app.timer_state {
        TimerState::Focus { remaining, target, started_at, total_paused } => {
            let elapsed = *target - *remaining;
            vec![
                Line::from(Span::styled(
                    format!("  Started: {}  │  Elapsed: {}",
                        started_at.format("%H:%M"),
                        format_duration(elapsed),
                    ),
                    Style::default().fg(app.theme.fg_secondary),
                )),
                Line::from(Span::styled(
                    format!("  Pauses: {}    │  Target:  {}",
                        if total_paused.as_secs() > 0 { format_duration(*total_paused) } else { "0".into() },
                        format_duration(*target),
                    ),
                    Style::default().fg(app.theme.fg_secondary),
                )),
            ]
        }
        TimerState::Overflow { elapsed, target_was, started_at: _, .. } => {
            let total = *target_was + *elapsed;
            vec![
                Line::from(Span::styled(
                    format!("  Target: {}  │  Overflow: +{}",
                        format_duration(*target_was),
                        format_duration(*elapsed),
                    ),
                    Style::default().fg(app.theme.fg_secondary),
                )),
                Line::from(Span::styled(
                    format!("  Total:  {}  │  You're in the zone!", format_duration(total)),
                    Style::default().fg(app.theme.overflow_accent),
                )),
            ]
        }
        TimerState::Paused { inner: _, .. } => {
            vec![
                Line::from(Span::styled(
                    "  ⏸  PAUSED — press Space to resume",
                    Style::default().fg(app.theme.paused_fg).add_modifier(Modifier::BOLD),
                )),
            ]
        }
        _ => vec![],
    };

    let para = Paragraph::new(lines).alignment(Alignment::Center);
    frame.render_widget(para, area);
}

fn render_today_stats(frame: &mut Frame, app: &App, area: Rect) {
    let hours = (app.today_focus_mins / 60.0) as u64;
    let mins = (app.today_focus_mins % 60.0) as u64;
    let time_str = if hours > 0 {
        format!("{}h {}m", hours, mins)
    } else {
        format!("{}m", mins)
    };

    let line = Line::from(Span::styled(
        format!("  Today: {} focused │ {} sessions │ Streak: {} days",
            time_str, app.today_sessions, app.streak),
        Style::default().fg(app.theme.fg_secondary),
    ));
    let para = Paragraph::new(vec![line]).alignment(Alignment::Center);
    frame.render_widget(para, area);
}

fn render_idle_buttons(frame: &mut Frame, app: &App, area: Rect) {
    if area.height < 3 { return; }

    let lines = vec![
        Line::from(""),
        Line::from(vec![
            Span::styled("  Duration: ", Style::default().fg(app.theme.fg_secondary)),
            Span::styled(
                format!("{} min", app.focus_duration_mins),
                Style::default().fg(app.theme.accent).add_modifier(Modifier::BOLD),
            ),
            Span::styled("  (+/- to adjust)", Style::default().fg(app.theme.fg_secondary)),
        ]),
        Line::from(Span::styled(
            "  [Enter] Start Focus  [b] Start Break  [i] Intention  [c] Category",
            Style::default().fg(app.theme.fg_secondary),
        )),
    ];

    let para = Paragraph::new(lines).alignment(Alignment::Center);
    frame.render_widget(para, area);
}

fn render_intention_overlay(frame: &mut Frame, app: &App, ) {
    let area = frame.area();
    let overlay_width = 50.min(area.width.saturating_sub(4));
    let overlay_height = 10.min(area.height.saturating_sub(4));
    let x = (area.width.saturating_sub(overlay_width)) / 2;
    let y = (area.height.saturating_sub(overlay_height)) / 2;
    let rect = Rect::new(x, y, overlay_width, overlay_height);

    frame.render_widget(Clear, rect);

    let block = Block::default()
        .title(" What are you working on? ")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(app.theme.border_focus))
        .style(Style::default().bg(app.theme.bg_secondary));

    let inner = block.inner(rect);
    frame.render_widget(block, rect);

    let input_line = Line::from(vec![
        Span::styled("> ", Style::default().fg(app.theme.accent)),
        Span::styled(&app.intention, Style::default().fg(app.theme.fg)),
        Span::styled("█", Style::default().fg(app.theme.accent)),
    ]);

    let help_line = Line::from(Span::styled(
        "Enter:confirm  Esc:cancel",
        Style::default().fg(app.theme.fg_secondary),
    ));

    let lines = vec![
        Line::from(""),
        input_line,
        Line::from(""),
        help_line,
    ];

    let para = Paragraph::new(lines);
    frame.render_widget(para, inner);
}

fn render_category_overlay(frame: &mut Frame, app: &App) {
    let area = frame.area();
    let overlay_height = (app.categories.len() as u16 + 6).min(area.height.saturating_sub(4));
    let overlay_width = 40.min(area.width.saturating_sub(4));
    let x = (area.width.saturating_sub(overlay_width)) / 2;
    let y = (area.height.saturating_sub(overlay_height)) / 2;
    let rect = Rect::new(x, y, overlay_width, overlay_height);

    frame.render_widget(Clear, rect);

    let block = Block::default()
        .title(" Select Category ")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(app.theme.border_focus))
        .style(Style::default().bg(app.theme.bg_secondary));

    let inner = block.inner(rect);
    frame.render_widget(block, rect);

    let mut lines = vec![Line::from("")];

    for (i, cat) in app.categories.iter().enumerate() {
        let color = parse_hex_color(&cat.hex_color);
        let marker = if i == app.selected_category_idx { "> " } else { "  " };
        let bullet = "● ";
        lines.push(Line::from(vec![
            Span::styled(marker, Style::default().fg(app.theme.accent)),
            Span::styled(bullet, Style::default().fg(color)),
            Span::styled(&cat.title, Style::default().fg(app.theme.fg)),
        ]));
    }

    lines.push(Line::from(""));
    lines.push(Line::from(Span::styled(
        "↑/↓:select  Enter:confirm  Esc:cancel",
        Style::default().fg(app.theme.fg_secondary),
    )));

    let para = Paragraph::new(lines);
    frame.render_widget(para, inner);
}

pub fn parse_hex_color(hex: &str) -> Color {
    let hex = hex.trim_start_matches('#');
    if hex.len() == 6 {
        if let (Ok(r), Ok(g), Ok(b)) = (
            u8::from_str_radix(&hex[0..2], 16),
            u8::from_str_radix(&hex[2..4], 16),
            u8::from_str_radix(&hex[4..6], 16),
        ) {
            return Color::Rgb(r, g, b);
        }
    }
    Color::White
}
