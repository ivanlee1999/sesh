use chrono::{DateTime, Utc};
use std::time::Duration;

/// Core timer state machine matching DESIGN.md Section 5
#[derive(Debug, Clone)]
pub enum TimerState {
    Idle,
    Focus {
        remaining: Duration,
        target: Duration,
        started_at: DateTime<Utc>,
        total_paused: Duration,
    },
    Overflow {
        elapsed: Duration,
        target_was: Duration,
        started_at: DateTime<Utc>,
        total_paused: Duration,
    },
    Paused {
        inner: Box<TimerState>,
        paused_at: DateTime<Utc>,
    },
    Break {
        remaining: Duration,
        target: Duration,
        break_type: BreakType,
        started_at: DateTime<Utc>,
    },
    BreakOverflow {
        elapsed: Duration,
        break_type: BreakType,
    },
    Abandoned {
        previous: Box<TimerState>,
        undo_deadline: std::time::Instant,
    },
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BreakType {
    Short,
    Long,
}

impl std::fmt::Display for BreakType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BreakType::Short => write!(f, "Short"),
            BreakType::Long => write!(f, "Long"),
        }
    }
}

impl TimerState {
    pub fn is_active(&self) -> bool {
        !matches!(self, TimerState::Idle)
    }

    pub fn is_running(&self) -> bool {
        matches!(
            self,
            TimerState::Focus { .. }
                | TimerState::Overflow { .. }
                | TimerState::Break { .. }
                | TimerState::BreakOverflow { .. }
        )
    }

    pub fn is_paused(&self) -> bool {
        matches!(self, TimerState::Paused { .. })
    }

    pub fn state_label(&self) -> &'static str {
        match self {
            TimerState::Idle => "IDLE",
            TimerState::Focus { .. } => "FOCUS",
            TimerState::Overflow { .. } => "OVERFLOW",
            TimerState::Paused { .. } => "PAUSED",
            TimerState::Break { .. } => "BREAK",
            TimerState::BreakOverflow { .. } => "BREAK OVERFLOW",
            TimerState::Abandoned { .. } => "ABANDONED",
        }
    }

    /// Get the display time string (e.g., "18:42" or "+2:14")
    pub fn display_time(&self) -> String {
        match self {
            TimerState::Idle => String::new(),
            TimerState::Focus { remaining, .. } => format_duration(*remaining),
            TimerState::Overflow { elapsed, .. } => format!("+{}", format_duration(*elapsed)),
            TimerState::Paused { inner, .. } => inner.display_time(),
            TimerState::Break { remaining, .. } => format_duration(*remaining),
            TimerState::BreakOverflow { elapsed, .. } => format!("+{}", format_duration(*elapsed)),
            TimerState::Abandoned { .. } => String::new(),
        }
    }

    /// Get progress ratio 0.0..=1.0 (for focus/break), or >1.0 for overflow
    pub fn progress(&self) -> f64 {
        match self {
            TimerState::Focus {
                remaining, target, ..
            } => {
                let elapsed = target.as_secs_f64() - remaining.as_secs_f64();
                if target.as_secs() == 0 {
                    1.0
                } else {
                    elapsed / target.as_secs_f64()
                }
            }
            TimerState::Overflow { elapsed, target_was, .. } => {
                1.0 + elapsed.as_secs_f64() / target_was.as_secs_f64().max(1.0)
            }
            TimerState::Break {
                remaining, target, ..
            } => {
                let elapsed = target.as_secs_f64() - remaining.as_secs_f64();
                if target.as_secs() == 0 {
                    1.0
                } else {
                    elapsed / target.as_secs_f64()
                }
            }
            TimerState::Paused { inner, .. } => inner.progress(),
            _ => 0.0,
        }
    }
}

pub fn format_duration(d: Duration) -> String {
    let total_secs = d.as_secs();
    let mins = total_secs / 60;
    let secs = total_secs % 60;
    format!("{:02}:{:02}", mins, secs)
}
