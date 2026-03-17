use crate::config::Config;
use crate::config::theme::Theme;
use crate::db::categories::Category;
use crate::db::Database;
use crate::state::{BreakType, TimerState};
use chrono::Utc;
use std::time::{Duration, Instant};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Screen {
    Timer,
    Analytics,
    History,
    Settings,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum InputMode {
    Normal,
    Intention,
    CategoryPicker,
}

pub struct App {
    pub timer_state: TimerState,
    pub current_screen: Screen,
    pub input_mode: InputMode,
    pub should_quit: bool,

    // Timer config
    pub focus_duration_mins: u64,
    pub target_duration: Duration,

    // Intention & category
    pub intention: String,
    pub intention_cursor: usize,
    pub categories: Vec<Category>,
    pub selected_category_idx: usize,

    // History
    pub history_offset: usize,
    pub history_selected: usize,

    // Analytics
    pub today_focus_mins: f64,
    pub today_sessions: i64,
    pub streak: i64,
    pub category_breakdown: Vec<(String, String, f64)>,

    // DB & Config
    pub db: Database,
    pub config: Config,
    pub theme: Theme,

    // Internal
    pub started_at_chrono: Option<chrono::DateTime<Utc>>,
    pub cumulative_focus: Duration,
    pub pause_accumulated: Duration,
}

impl App {
    pub fn new(db: Database, config: Config) -> Self {
        let theme = Theme::from_name(&config.general.theme);
        let categories = db.get_categories().unwrap_or_default();
        let focus_duration_mins = config.timer.focus_duration;

        let (today_focus_mins, today_sessions) = db.get_today_stats().unwrap_or((0.0, 0));
        let streak = db.get_streak().unwrap_or(0);
        let category_breakdown = db.get_category_breakdown_today().unwrap_or_default();

        Self {
            timer_state: TimerState::Idle,
            current_screen: Screen::Timer,
            input_mode: InputMode::Normal,
            should_quit: false,
            focus_duration_mins,
            target_duration: Duration::from_secs(focus_duration_mins * 60),
            intention: String::new(),
            intention_cursor: 0,
            categories,
            selected_category_idx: 0,
            history_offset: 0,
            history_selected: 0,
            today_focus_mins,
            today_sessions,
            streak,
            category_breakdown,
            db,
            config,
            theme,
            started_at_chrono: None,
            cumulative_focus: Duration::ZERO,
            pause_accumulated: Duration::ZERO,
        }
    }

    pub fn tick(&mut self) {
        let tick = Duration::from_millis(self.config.general.tick_rate_ms);
        match &self.timer_state {
            TimerState::Focus { remaining, target, started_at, total_paused } => {
                let remaining = *remaining;
                let target = *target;
                let started_at = *started_at;
                let total_paused = *total_paused;
                if remaining <= tick {
                    // Transition to overflow
                    self.timer_state = TimerState::Overflow {
                        elapsed: Duration::ZERO,
                        target_was: target,
                        started_at,
                        total_paused,
                    };
                } else {
                    self.timer_state = TimerState::Focus {
                        remaining: remaining - tick,
                        target,
                        started_at,
                        total_paused,
                    };
                }
            }
            TimerState::Overflow { elapsed, target_was, started_at, total_paused } => {
                let elapsed = *elapsed;
                let target_was = *target_was;
                let started_at = *started_at;
                let total_paused = *total_paused;
                self.timer_state = TimerState::Overflow {
                    elapsed: elapsed + tick,
                    target_was,
                    started_at,
                    total_paused,
                };
            }
            TimerState::Break { remaining, target, break_type, started_at } => {
                let remaining = *remaining;
                let target = *target;
                let break_type = *break_type;
                let started_at = *started_at;
                if remaining <= tick {
                    self.timer_state = TimerState::BreakOverflow {
                        elapsed: Duration::ZERO,
                        break_type,
                    };
                } else {
                    self.timer_state = TimerState::Break {
                        remaining: remaining - tick,
                        target,
                        break_type,
                        started_at,
                    };
                }
            }
            TimerState::BreakOverflow { elapsed, break_type } => {
                let elapsed = *elapsed;
                let break_type = *break_type;
                self.timer_state = TimerState::BreakOverflow {
                    elapsed: elapsed + tick,
                    break_type,
                };
            }
            TimerState::Abandoned { undo_deadline, .. } => {
                if Instant::now() >= *undo_deadline {
                    self.timer_state = TimerState::Idle;
                }
            }
            _ => {}
        }
    }

    pub fn start_focus(&mut self) {
        let now = Utc::now();
        self.started_at_chrono = Some(now);
        self.pause_accumulated = Duration::ZERO;
        self.timer_state = TimerState::Focus {
            remaining: self.target_duration,
            target: self.target_duration,
            started_at: now,
            total_paused: Duration::ZERO,
        };
    }

    pub fn start_break(&mut self, break_type: BreakType) {
        let dur = match break_type {
            BreakType::Short => Duration::from_secs(self.config.timer.short_break_duration * 60),
            BreakType::Long => Duration::from_secs(self.config.timer.long_break_duration * 60),
        };
        self.timer_state = TimerState::Break {
            remaining: dur,
            target: dur,
            break_type,
            started_at: Utc::now(),
        };
    }

    pub fn toggle_pause(&mut self) {
        match &self.timer_state {
            TimerState::Focus { .. } | TimerState::Overflow { .. } => {
                let inner = Box::new(self.timer_state.clone());
                self.timer_state = TimerState::Paused {
                    inner,
                    paused_at: Utc::now(),
                };
            }
            TimerState::Paused { inner, paused_at } => {
                let pause_dur = (Utc::now() - *paused_at).to_std().unwrap_or(Duration::ZERO);
                self.pause_accumulated += pause_dur;
                let mut restored = *inner.clone();
                // Update total_paused in the restored state
                match &mut restored {
                    TimerState::Focus { total_paused, .. } => {
                        *total_paused = self.pause_accumulated;
                    }
                    TimerState::Overflow { total_paused, .. } => {
                        *total_paused = self.pause_accumulated;
                    }
                    _ => {}
                }
                self.timer_state = restored;
            }
            _ => {}
        }
    }

    pub fn finish_session(&mut self) {
        if let Some(started) = self.started_at_chrono {
            let now = Utc::now();
            let total_elapsed = (now - started).to_std().unwrap_or(Duration::ZERO);
            let pause_secs = self.pause_accumulated.as_secs() as i64;
            let actual_secs = total_elapsed.as_secs() as i64 - pause_secs;
            let target_secs = self.target_duration.as_secs() as i64;
            let overflow_secs = (actual_secs - target_secs).max(0);

            let session_type = if actual_secs >= target_secs {
                "full_focus"
            } else {
                "partial_focus"
            };

            let cat_id = self.categories.get(self.selected_category_idx).map(|c| c.id.as_str());

            let _ = self.db.save_session(
                &self.intention,
                cat_id,
                session_type,
                target_secs,
                actual_secs,
                pause_secs,
                overflow_secs,
                &started.format("%Y-%m-%dT%H:%M:%S").to_string(),
                &now.format("%Y-%m-%dT%H:%M:%S").to_string(),
                None,
            );

            self.cumulative_focus += Duration::from_secs(actual_secs.max(0) as u64);
            self.refresh_stats();
        }

        self.timer_state = TimerState::Idle;
        self.started_at_chrono = None;
    }

    pub fn abandon_session(&mut self) {
        if self.timer_state.is_active() && !self.timer_state.is_paused() || self.timer_state.is_paused() {
            let prev = Box::new(self.timer_state.clone());
            self.timer_state = TimerState::Abandoned {
                previous: prev,
                undo_deadline: Instant::now() + Duration::from_secs(5),
            };
        }
    }

    pub fn undo_abandon(&mut self) {
        if let TimerState::Abandoned { previous, undo_deadline } = &self.timer_state {
            if Instant::now() < *undo_deadline {
                self.timer_state = *previous.clone();
            }
        }
    }

    pub fn adjust_duration(&mut self, delta_mins: i64) {
        if matches!(self.timer_state, TimerState::Idle) {
            let new = (self.focus_duration_mins as i64 + delta_mins).max(1) as u64;
            self.focus_duration_mins = new;
            self.target_duration = Duration::from_secs(new * 60);
        }
    }

    pub fn refresh_stats(&mut self) {
        if let Ok((mins, count)) = self.db.get_today_stats() {
            self.today_focus_mins = mins;
            self.today_sessions = count;
        }
        if let Ok(s) = self.db.get_streak() {
            self.streak = s;
        }
        if let Ok(b) = self.db.get_category_breakdown_today() {
            self.category_breakdown = b;
        }
    }

    pub fn selected_category(&self) -> Option<&Category> {
        self.categories.get(self.selected_category_idx)
    }
}

