use crate::app::App;
use crate::state::TimerState;
use ratatui::layout::{Alignment, Rect};
use ratatui::style::{Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::Paragraph;
use ratatui::Frame;

/// Render a Unicode circle with timer text centered inside.
/// The circle is drawn using box-drawing characters.
pub fn render(frame: &mut Frame, app: &App, area: Rect) {
    let (time_str, color) = match &app.timer_state {
        TimerState::Idle => {
            let mins = app.focus_duration_mins;
            (format!("{:02}:00", mins), app.theme.fg)
        }
        TimerState::Focus { .. } => {
            (app.timer_state.display_time(), app.theme.focus_accent)
        }
        TimerState::Overflow { .. } => {
            (app.timer_state.display_time(), app.theme.overflow_accent)
        }
        TimerState::Paused { .. } => {
            (app.timer_state.display_time(), app.theme.paused_fg)
        }
        TimerState::Break { .. } => {
            (app.timer_state.display_time(), app.theme.break_accent)
        }
        TimerState::BreakOverflow { .. } => {
            (app.timer_state.display_time(), app.theme.break_accent)
        }
        TimerState::Abandoned { .. } => {
            ("ABANDONED".to_string(), app.theme.error)
        }
    };

    let state_label = app.timer_state.state_label();
    let label_color = color;

    // Draw the circle with progress fill
    let progress = app.timer_state.progress().clamp(0.0, 1.0);

    // Unicode circle using box-drawing characters
    // Height = 11 lines for the circle + time + label
    let circle_lines = build_circle(&time_str, state_label, color, label_color, &app.theme, progress);

    let para = Paragraph::new(circle_lines).alignment(Alignment::Center);
    frame.render_widget(para, area);
}

fn build_circle(
    time: &str,
    label: &str,
    time_color: ratatui::style::Color,
    label_color: ratatui::style::Color,
    theme: &crate::config::theme::Theme,
    progress: f64,
) -> Vec<Line<'static>> {
    let border_color = time_color;
    let bs = Style::default().fg(border_color);
    let ts = Style::default().fg(time_color).add_modifier(Modifier::BOLD);
    let ls = Style::default().fg(label_color);

    // Progress indicator: fill segments of the circle
    // We use a simple filled/unfilled approach on the border characters
    let fill_color = if progress > 0.0 { time_color } else { theme.border };
    let empty_color = theme.bg_secondary;

    // Simple filled segments based on progress (12 segments around the clock)
    let segments = 12;
    let filled = (progress * segments as f64) as usize;

    let seg_style = |idx: usize| -> Style {
        if idx < filled {
            Style::default().fg(fill_color)
        } else {
            Style::default().fg(empty_color)
        }
    };

    vec![
        Line::from(""),
        Line::from(vec![
            Span::styled("╭────────────╮", bs),
        ]),
        Line::from(vec![
            Span::styled("╭────╯", seg_style(11)),
            Span::styled("            ", Style::default()),
            Span::styled("╰────╮", seg_style(1)),
        ]),
        Line::from(vec![
            Span::styled("╭─╯", seg_style(10)),
            Span::styled("                  ", Style::default()),
            Span::styled("╰─╮", seg_style(2)),
        ]),
        Line::from(vec![
            Span::styled("│", seg_style(9)),
            Span::styled(format!("      {:^10}      ", time), ts),
            Span::styled("│", seg_style(3)),
        ]),
        Line::from(vec![
            Span::styled("│", seg_style(9)),
            Span::styled(format!("      {:^10}      ", label), ls),
            Span::styled("│", seg_style(3)),
        ]),
        Line::from(vec![
            Span::styled("╰─╮", seg_style(8)),
            Span::styled("                  ", Style::default()),
            Span::styled("╭─╯", seg_style(4)),
        ]),
        Line::from(vec![
            Span::styled("╰────╮", seg_style(7)),
            Span::styled("            ", Style::default()),
            Span::styled("╭────╯", seg_style(5)),
        ]),
        Line::from(vec![
            Span::styled("╰────────────╯", seg_style(6)),
        ]),
        Line::from(""),
    ]
}
