use crate::app::{App, Screen};
use ratatui::layout::Rect;
use ratatui::style::{Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Tabs};
use ratatui::Frame;

pub fn render(frame: &mut Frame, app: &App, area: Rect) {
    let titles = vec!["Timer", "Analytics", "History", "Settings"];
    let selected = match app.current_screen {
        Screen::Timer => 0,
        Screen::Analytics => 1,
        Screen::History => 2,
        Screen::Settings => 3,
    };

    let tabs = Tabs::new(titles.iter().map(|t| Line::from(*t)).collect::<Vec<_>>())
        .block(
            Block::default()
                .borders(Borders::BOTTOM)
                .border_style(Style::default().fg(app.theme.border))
                .style(Style::default().bg(app.theme.bg)),
        )
        .select(selected)
        .style(Style::default().fg(app.theme.fg_secondary))
        .highlight_style(
            Style::default()
                .fg(app.theme.accent)
                .add_modifier(Modifier::BOLD),
        )
        .divider(Span::styled(" │ ", Style::default().fg(app.theme.border)));

    frame.render_widget(tabs, area);
}
