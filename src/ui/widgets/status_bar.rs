use crate::app::App;
use crate::state::TimerState;
use ratatui::layout::Rect;
use ratatui::style::{Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Paragraph};
use ratatui::Frame;

pub fn render(frame: &mut Frame, app: &App, area: Rect) {
    let block = Block::default()
        .borders(Borders::TOP)
        .border_style(Style::default().fg(app.theme.border))
        .style(Style::default().bg(app.theme.status_bar_bg));

    let inner = block.inner(area);
    frame.render_widget(block, area);

    let (state_span, hint_span) = match &app.timer_state {
        TimerState::Idle => (
            Span::styled(" IDLE ", Style::default()
                .fg(app.theme.fg_secondary)
                .add_modifier(Modifier::BOLD)),
            Span::styled(
                "│ Enter:focus  b:break  i:intention  c:category  q:quit  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::Focus { .. } => (
            Span::styled(
                format!(" ⏱ {} FOCUS ", app.timer_state.display_time()),
                Style::default().fg(app.theme.focus_accent).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ space:pause  f:finish  b:break  x:abandon  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::Overflow { .. } => (
            Span::styled(
                format!(" ◆ {} OVERFLOW ", app.timer_state.display_time()),
                Style::default().fg(app.theme.overflow_accent).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ f:finish  b:break  x:abandon  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::Paused { .. } => (
            Span::styled(
                format!(" ⏸ {} PAUSED ", app.timer_state.display_time()),
                Style::default().fg(app.theme.paused_fg).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ space:resume  f:finish  x:abandon  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::Break { break_type, .. } => (
            Span::styled(
                format!(" ☕ {} {} BREAK ", app.timer_state.display_time(), break_type),
                Style::default().fg(app.theme.break_accent).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ Enter:end break  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::BreakOverflow { break_type, .. } => (
            Span::styled(
                format!(" ☕ {} {} BREAK OVER ", app.timer_state.display_time(), break_type),
                Style::default().fg(app.theme.break_accent).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ Enter:start focus  ?:help",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
        TimerState::Abandoned { .. } => (
            Span::styled(
                " ABANDONED ",
                Style::default().fg(app.theme.error).add_modifier(Modifier::BOLD),
            ),
            Span::styled(
                "│ u:undo (5s)  Press any key to dismiss",
                Style::default().fg(app.theme.fg_secondary),
            ),
        ),
    };

    let line = Line::from(vec![state_span, Span::raw(" "), hint_span]);
    let para = Paragraph::new(vec![line]);
    frame.render_widget(para, inner);
}
