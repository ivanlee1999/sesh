mod app;
mod calendar;
mod cli;
mod config;
mod db;
mod state;
mod todoist;
mod ui;

use app::{App, InputMode, Screen};
use clap::Parser;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{backend::CrosstermBackend, Terminal};
use state::{BreakType, TimerState};
use std::io;
use std::time::{Duration, Instant};

fn main() -> anyhow::Result<()> {
    let cli = cli::Cli::parse();

    if let Some(cmd) = cli.command {
        return cli::run_cli(cmd);
    }

    // TUI mode
    run_tui()?;
    Ok(())
}

fn run_tui() -> anyhow::Result<()> {
    // Setup terminal
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Create app
    let config = config::Config::load();
    let tick_rate = Duration::from_millis(config.general.tick_rate_ms);
    let db = db::Database::open_default()?;
    db.insert_default_categories()?;
    let mut app = App::new(db, config);

    // Main loop
    let mut last_tick = Instant::now();

    loop {
        terminal.draw(|frame| ui::render(frame, &app))?;

        let timeout = tick_rate.saturating_sub(last_tick.elapsed());
        if event::poll(timeout)? {
            if let Event::Key(key) = event::read()? {
                handle_key(&mut app, key.code, key.modifiers);
            }
        }

        if last_tick.elapsed() >= tick_rate {
            app.tick();
            last_tick = Instant::now();
        }

        if app.should_quit {
            break;
        }
    }

    // Restore terminal
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    Ok(())
}

fn handle_key(app: &mut App, code: KeyCode, modifiers: KeyModifiers) {
    // Handle input modes first
    match app.input_mode {
        InputMode::Intention => {
            match code {
                KeyCode::Esc => {
                    app.input_mode = InputMode::Normal;
                }
                KeyCode::Enter => {
                    app.input_mode = InputMode::Normal;
                }
                KeyCode::Backspace => {
                    app.intention.pop();
                }
                KeyCode::Char(c) => {
                    app.intention.push(c);
                }
                _ => {}
            }
            return;
        }
        InputMode::CategoryPicker => {
            match code {
                KeyCode::Esc => {
                    app.input_mode = InputMode::Normal;
                }
                KeyCode::Enter => {
                    app.input_mode = InputMode::Normal;
                }
                KeyCode::Up | KeyCode::Char('k') => {
                    if app.selected_category_idx > 0 {
                        app.selected_category_idx -= 1;
                    }
                }
                KeyCode::Down | KeyCode::Char('j') => {
                    if app.selected_category_idx + 1 < app.categories.len() {
                        app.selected_category_idx += 1;
                    }
                }
                _ => {}
            }
            return;
        }
        InputMode::Normal => {}
    }

    // Ctrl+C always quits
    if modifiers.contains(KeyModifiers::CONTROL) && code == KeyCode::Char('c') {
        app.should_quit = true;
        return;
    }

    // Global keys
    match code {
        KeyCode::Char('q') => {
            if !app.timer_state.is_active() {
                app.should_quit = true;
            }
        }
        KeyCode::Char('1') => app.current_screen = Screen::Timer,
        KeyCode::Char('2') => {
            app.refresh_stats();
            app.current_screen = Screen::Analytics;
        }
        KeyCode::Char('3') => app.current_screen = Screen::History,
        KeyCode::Char('4') => app.current_screen = Screen::Settings,
        KeyCode::Tab => {
            app.current_screen = match app.current_screen {
                Screen::Timer => Screen::Analytics,
                Screen::Analytics => Screen::History,
                Screen::History => Screen::Settings,
                Screen::Settings => Screen::Timer,
            };
            if matches!(app.current_screen, Screen::Analytics) {
                app.refresh_stats();
            }
        }
        _ => {}
    }

    // Screen-specific keys
    match app.current_screen {
        Screen::Timer => handle_timer_key(app, code),
        Screen::History => handle_history_key(app, code),
        _ => {}
    }
}

fn handle_timer_key(app: &mut App, code: KeyCode) {
    match &app.timer_state {
        TimerState::Idle => {
            match code {
                KeyCode::Enter => {
                    app.start_focus();
                }
                KeyCode::Char('b') => {
                    // Determine break type
                    let break_type = if app.cumulative_focus >= Duration::from_secs(app.config.timer.long_break_after * 60) {
                        app.cumulative_focus = Duration::ZERO;
                        BreakType::Long
                    } else {
                        BreakType::Short
                    };
                    app.start_break(break_type);
                }
                KeyCode::Char('i') => {
                    app.input_mode = InputMode::Intention;
                }
                KeyCode::Char('c') => {
                    app.input_mode = InputMode::CategoryPicker;
                }
                KeyCode::Char('+') | KeyCode::Char('=') => {
                    app.adjust_duration(5);
                }
                KeyCode::Char('-') => {
                    app.adjust_duration(-5);
                }
                KeyCode::Char('>') | KeyCode::Char('.') => {
                    app.adjust_duration(1);
                }
                KeyCode::Char('<') | KeyCode::Char(',') => {
                    app.adjust_duration(-1);
                }
                _ => {}
            }
        }
        TimerState::Focus { .. } | TimerState::Overflow { .. } => {
            match code {
                KeyCode::Char(' ') => app.toggle_pause(),
                KeyCode::Char('f') => app.finish_session(),
                KeyCode::Char('x') => app.abandon_session(),
                KeyCode::Char('b') => {
                    app.finish_session();
                    let break_type = if app.cumulative_focus >= Duration::from_secs(app.config.timer.long_break_after * 60) {
                        app.cumulative_focus = Duration::ZERO;
                        BreakType::Long
                    } else {
                        BreakType::Short
                    };
                    app.start_break(break_type);
                }
                _ => {}
            }
        }
        TimerState::Paused { .. } => {
            match code {
                KeyCode::Char(' ') => app.toggle_pause(),
                KeyCode::Char('f') => app.finish_session(),
                KeyCode::Char('x') => app.abandon_session(),
                _ => {}
            }
        }
        TimerState::Break { .. } | TimerState::BreakOverflow { .. } => {
            match code {
                KeyCode::Enter | KeyCode::Char('f') => {
                    app.timer_state = TimerState::Idle;
                }
                _ => {}
            }
        }
        TimerState::Abandoned { .. } => {
            match code {
                KeyCode::Char('u') => app.undo_abandon(),
                _ => {
                    app.timer_state = TimerState::Idle;
                    app.started_at_chrono = None;
                }
            }
        }
    }
}

fn handle_history_key(app: &mut App, code: KeyCode) {
    match code {
        KeyCode::Up | KeyCode::Char('k') => {
            if app.history_selected > 0 {
                app.history_selected -= 1;
            }
        }
        KeyCode::Down | KeyCode::Char('j') => {
            app.history_selected += 1;
        }
        _ => {}
    }
}
