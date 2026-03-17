pub mod timer;
pub mod analytics;
pub mod history;
pub mod widgets;

use crate::app::{App, Screen};
use ratatui::Frame;

pub fn render(frame: &mut Frame, app: &App) {
    match app.current_screen {
        Screen::Timer => timer::render(frame, app),
        Screen::Analytics => analytics::render(frame, app),
        Screen::History => history::render(frame, app),
        Screen::Settings => render_settings(frame, app),
    }
}

fn render_settings(frame: &mut Frame, app: &App) {
    use ratatui::layout::{Constraint, Direction, Layout};
    use ratatui::style::Style;
    use ratatui::text::{Line, Span};
    use ratatui::widgets::{Block, Borders, Paragraph};

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),
            Constraint::Min(0),
            Constraint::Length(3),
        ])
        .split(frame.area());

    widgets::tab_bar::render(frame, app, chunks[0]);

    let settings_text = vec![
        Line::from(""),
        Line::from(Span::styled("  Settings", Style::default().fg(app.theme.fg))),
        Line::from(""),
        Line::from(Span::styled(
            format!("  Theme:              {}", app.config.general.theme),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(Span::styled(
            format!("  Focus Duration:     {} min", app.config.timer.focus_duration),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(Span::styled(
            format!("  Short Break:        {} min", app.config.timer.short_break_duration),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(Span::styled(
            format!("  Long Break:         {} min", app.config.timer.long_break_duration),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(Span::styled(
            format!("  Long Break After:   {} min cumulative", app.config.timer.long_break_after),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(""),
        Line::from(Span::styled(
            format!("  Config: {}", crate::config::Config::config_path().display()),
            Style::default().fg(app.theme.fg_secondary),
        )),
        Line::from(Span::styled(
            format!("  Data:   {}", crate::config::Config::data_dir().display()),
            Style::default().fg(app.theme.fg_secondary),
        )),
    ];

    let block = Block::default()
        .borders(Borders::NONE)
        .style(Style::default().bg(app.theme.bg));
    let para = Paragraph::new(settings_text).block(block);
    frame.render_widget(para, chunks[1]);

    widgets::status_bar::render(frame, app, chunks[2]);
}
