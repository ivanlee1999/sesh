use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "sesh", version, about = "A terminal-native Pomodoro focus timer")]
pub struct Cli {
    #[command(subcommand)]
    pub command: Option<Commands>,

    /// Config file path
    #[arg(long, global = true)]
    pub config: Option<String>,

    /// Data directory path
    #[arg(long, global = true)]
    pub data_dir: Option<String>,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Start a focus session
    Start {
        /// Intention text
        #[arg(short, long)]
        intention: Option<String>,
        /// Category name
        #[arg(short, long)]
        category: Option<String>,
        /// Duration in minutes
        #[arg(short, long)]
        duration: Option<u64>,
    },
    /// Show current timer status
    Status {
        /// Output format: json, human
        #[arg(short, long, default_value = "json")]
        format: String,
    },
    /// Stop current session
    Stop,
    /// Show session history
    History {
        /// Number of sessions to show
        #[arg(short, long, default_value = "10")]
        limit: usize,
    },
    /// Show analytics summary
    Analytics {
        /// Period: today, week, month
        #[arg(short, long, default_value = "today")]
        period: String,
    },
}

pub fn run_cli(cmd: Commands) -> anyhow::Result<()> {
    let db = crate::db::Database::open_default()?;

    match cmd {
        Commands::Start { intention, category: _, duration } => {
            println!("Starting focus session...");
            if let Some(i) = &intention {
                println!("  Intention: {}", i);
            }
            if let Some(d) = duration {
                println!("  Duration: {} minutes", d);
            }
            println!("(Non-interactive mode not yet fully implemented. Use TUI mode instead.)");
        }
        Commands::Status { format } => {
            let (focus_mins, sessions) = db.get_today_stats()?;
            if format == "json" {
                println!(r#"{{"state":"idle","today_focus_minutes":{:.1},"today_sessions":{}}}"#, focus_mins, sessions);
            } else {
                let hours = (focus_mins / 60.0) as u64;
                let mins = (focus_mins % 60.0) as u64;
                println!("Status: IDLE");
                println!("Today: {}h {}m focused │ {} sessions", hours, mins, sessions);
            }
        }
        Commands::Stop => {
            println!("No active session to stop.");
        }
        Commands::History { limit } => {
            let sessions = db.get_sessions(limit)?;
            if sessions.is_empty() {
                println!("No sessions recorded yet.");
                return Ok(());
            }
            println!("{:<20} {:<30} {:<15} {:>8}", "Time", "Intention", "Category", "Duration");
            println!("{}", "─".repeat(75));
            for s in &sessions {
                let start = s.started_at.split('T').nth(1).unwrap_or("").get(..5).unwrap_or("??:??");
                let end = s.ended_at.split('T').nth(1).unwrap_or("").get(..5).unwrap_or("??:??");
                let dur = format!("{}:{:02}", s.actual_seconds / 60, s.actual_seconds % 60);
                let cat = s.category_title.as_deref().unwrap_or("—");
                let title = if s.title.is_empty() { "(none)" } else { &s.title };
                println!("{:<20} {:<30} {:<15} {:>8}",
                    format!("{} - {}", start, end), title, cat, dur);
            }
        }
        Commands::Analytics { period: _ } => {
            let (focus_mins, sessions) = db.get_today_stats()?;
            let streak = db.get_streak()?;
            let hours = (focus_mins / 60.0) as u64;
            let mins = (focus_mins % 60.0) as u64;
            println!("Today: {}h {}m focused │ {} sessions │ Streak: {} days",
                hours, mins, sessions, streak);

            let breakdown = db.get_category_breakdown_today()?;
            let total: f64 = breakdown.iter().map(|(_, _, m)| m).sum();
            for (name, _, minutes) in &breakdown {
                let pct = if total > 0.0 { minutes / total * 100.0 } else { 0.0 };
                let bar_len = (pct / 5.0) as usize;
                let bar: String = "█".repeat(bar_len) + &"░".repeat(20 - bar_len);
                let hrs = (*minutes / 60.0) as u64;
                let mns = (*minutes % 60.0) as u64;
                let time = if hrs > 0 { format!("{}h {}m", hrs, mns) } else { format!("{}m", mns) };
                println!("  {:<16} {} {:>3.0}% {}", name, bar, pct, time);
            }
        }
    }
    Ok(())
}
