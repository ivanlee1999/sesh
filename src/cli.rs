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
        /// Todoist task ID to link this session to
        #[arg(long)]
        todoist: Option<String>,
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
    /// List today's Todoist tasks and pick one to start
    Todoist,
    /// Export session data
    Export {
        /// Output format: ics, csv, json
        #[arg(short, long, default_value = "ics")]
        format: String,
        /// Output file path (defaults to stdout for csv/json, or config ics_path for ics)
        #[arg(short, long)]
        output: Option<String>,
        /// Only export sessions since this date (YYYY-MM-DD)
        #[arg(long)]
        since: Option<String>,
    },
}

pub fn run_cli(cmd: Commands) -> anyhow::Result<()> {
    let config = crate::config::Config::load();
    let db = crate::db::Database::open_default()?;
    db.insert_default_categories()?;

    match cmd {
        Commands::Start { intention, category: _, duration, todoist: todoist_task_id } => {
            let mut final_intention = intention.clone();
            let mut matched_category: Option<String> = None;

            // If --todoist provided, fetch task info and populate intention/category
            if let Some(task_id) = &todoist_task_id {
                if config.todoist.api_token.is_empty() {
                    eprintln!("Error: No Todoist API token configured. See `sesh todoist` for setup.");
                    return Ok(());
                }
                let client = crate::todoist::TodoistClient::new(&config.todoist.api_token);
                match client.get_task(task_id) {
                    Ok(task) => {
                        if final_intention.is_none() {
                            final_intention = Some(task.content.clone());
                        }
                        // Try to match project to category
                        let projects = client.get_projects().unwrap_or_default();
                        let categories = db.get_categories().unwrap_or_default();
                        if let Some(idx) = client.match_project_to_category(&task.project_id, &projects, &categories) {
                            matched_category = Some(categories[idx].title.clone());
                        }
                        println!("Linked to Todoist task: {}", task.content);
                    }
                    Err(e) => {
                        eprintln!("Warning: Could not fetch Todoist task {}: {}", task_id, e);
                    }
                }
            }

            println!("Starting focus session...");
            if let Some(i) = &final_intention {
                println!("  Intention: {}", i);
            }
            if let Some(c) = &matched_category {
                println!("  Category: {}", c);
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
        Commands::Todoist => {
            run_todoist_picker(&config, &db)?;
        }
        Commands::Export { format, output, since: _ } => {
            run_export(&config, &db, &format, output.as_deref())?;
        }
    }
    Ok(())
}

fn run_todoist_picker(config: &crate::config::Config, db: &crate::db::Database) -> anyhow::Result<()> {
    if config.todoist.api_token.is_empty() {
        eprintln!("Error: No Todoist API token configured.");
        eprintln!("Add your token to {}", crate::config::Config::config_path().display());
        eprintln!("");
        eprintln!("  [todoist]");
        eprintln!("  api_token = \"your-api-token-here\"");
        eprintln!("");
        eprintln!("Get your token at: https://todoist.com/prefs/integrations");
        return Ok(());
    }

    let client = crate::todoist::TodoistClient::new(&config.todoist.api_token);

    println!("Fetching today's Todoist tasks...");
    let tasks = client.get_today_tasks()?;
    let projects = client.get_projects()?;
    let categories = db.get_categories()?;

    if tasks.is_empty() {
        println!("No tasks due today. You're all caught up!");
        return Ok(());
    }

    println!("");
    println!("{:<4} {:<50} {:<20}", "#", "Task", "Project");
    println!("{}", "─".repeat(74));

    for (i, task) in tasks.iter().enumerate() {
        let project_name = projects.iter()
            .find(|p| p.id == task.project_id)
            .map(|p| p.name.as_str())
            .unwrap_or("—");
        println!("{:<4} {:<50} {:<20}", i + 1, truncate(&task.content, 48), project_name);
    }

    println!("");
    println!("Start a session with: sesh start --todoist <task_id>");
    println!("  e.g., sesh start --todoist {}", tasks[0].id);
    Ok(())
}

fn run_export(
    config: &crate::config::Config,
    db: &crate::db::Database,
    format: &str,
    output: Option<&str>,
) -> anyhow::Result<()> {
    let sessions = db.get_sessions(10000)?;

    match format {
        "ics" => {
            let path = output
                .map(std::path::PathBuf::from)
                .unwrap_or_else(|| std::path::PathBuf::from(&config.calendar.ics_path));
            crate::calendar::export_ics(&sessions, &path)?;
            println!("Exported {} sessions to {}", sessions.len(), path.display());
        }
        "json" => {
            let json_sessions: Vec<serde_json::Value> = sessions.iter().map(|s| {
                serde_json::json!({
                    "id": s.id,
                    "title": s.title,
                    "category": s.category_title,
                    "type": s.session_type,
                    "target_seconds": s.target_seconds,
                    "actual_seconds": s.actual_seconds,
                    "pause_seconds": s.pause_seconds,
                    "overflow_seconds": s.overflow_seconds,
                    "started_at": s.started_at,
                    "ended_at": s.ended_at,
                    "notes": s.notes,
                })
            }).collect();
            let json = serde_json::to_string_pretty(&json_sessions)?;
            if let Some(path) = output {
                std::fs::write(path, &json)?;
                println!("Exported {} sessions to {}", sessions.len(), path);
            } else {
                println!("{}", json);
            }
        }
        "csv" => {
            let mut out = String::new();
            out.push_str("id,title,category,type,target_seconds,actual_seconds,pause_seconds,overflow_seconds,started_at,ended_at,notes\n");
            for s in &sessions {
                out.push_str(&format!(
                    "{},{},{},{},{},{},{},{},{},{},{}\n",
                    csv_escape(&s.id),
                    csv_escape(&s.title),
                    csv_escape(s.category_title.as_deref().unwrap_or("")),
                    csv_escape(&s.session_type),
                    s.target_seconds, s.actual_seconds, s.pause_seconds, s.overflow_seconds,
                    csv_escape(&s.started_at),
                    csv_escape(&s.ended_at),
                    csv_escape(s.notes.as_deref().unwrap_or("")),
                ));
            }
            if let Some(path) = output {
                std::fs::write(path, &out)?;
                println!("Exported {} sessions to {}", sessions.len(), path);
            } else {
                print!("{}", out);
            }
        }
        _ => {
            eprintln!("Unknown format: {}. Use ics, json, or csv.", format);
        }
    }
    Ok(())
}

fn truncate(s: &str, max: usize) -> String {
    if s.len() > max {
        format!("{}...", &s[..max - 3])
    } else {
        s.to_string()
    }
}

fn csv_escape(s: &str) -> String {
    if s.contains(',') || s.contains('"') || s.contains('\n') {
        format!("\"{}\"", s.replace('"', "\"\""))
    } else {
        s.to_string()
    }
}
