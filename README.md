# sesh

A terminal-native Pomodoro focus timer built with Rust and Ratatui.

> For developers who live in the terminal.

![Rust](https://img.shields.io/badge/Rust-1.70+-orange)
![License](https://img.shields.io/badge/license-MIT-blue)

## Features

- **Focus timer** with configurable duration (default 25 min)
- **Break timer** — short (5 min) and long (20 min) breaks
- **Overflow mode** — timer continues past 0 with visual indicator
- **Intention input** — set what you're working on before starting
- **Categories** — colored labels for organizing sessions (Development, Writing, etc.)
- **Session history** — scrollable log of all past sessions
- **Analytics** — today's focus time, session count, streak, category breakdown
- **Pause/resume** — freeze timer mid-session
- **CLI mode** — `sesh status`, `sesh history`, `sesh analytics` for scripting
- **TOML config** — customize durations, theme, keybindings
- **SQLite storage** — all data stored locally in `~/.local/share/sesh/`
- **Beautiful TUI** — Unicode circle timer, progress bars, tabbed interface

## Installation

### From source

```bash
cargo install --path .
```

### Build and run

```bash
cargo build --release
./target/release/sesh
```

## Usage

### Interactive (TUI)

```bash
sesh
```

This launches the full terminal UI with timer, analytics, history, and settings tabs.

### Keybindings

#### Global
| Key | Action |
|-----|--------|
| `1-4` | Switch tabs (Timer, Analytics, History, Settings) |
| `Tab` | Next tab |
| `q` | Quit (when idle) |
| `Ctrl+C` | Force quit |

#### Timer (Idle)
| Key | Action |
|-----|--------|
| `Enter` | Start focus session |
| `b` | Start break |
| `i` | Set intention |
| `c` | Pick category |
| `+`/`-` | Adjust duration ±5 min |
| `>`/`<` | Adjust duration ±1 min |

#### Timer (Running)
| Key | Action |
|-----|--------|
| `Space` | Pause/resume |
| `f` | Finish session |
| `b` | Finish + start break |
| `x` | Abandon session |

### CLI Commands

```bash
sesh status              # Show current status (JSON)
sesh status -f human     # Human-readable status
sesh history             # List recent sessions
sesh history -l 20       # Last 20 sessions
sesh analytics           # Today's analytics
```

### Todoist Integration

Link your Pomodoro sessions to Todoist tasks:

```bash
# List today's Todoist tasks
sesh todoist

# Start a session linked to a Todoist task
sesh start --todoist <task_id>

# The intention is auto-populated from the task content,
# and the category is auto-matched from the Todoist project name.
```

Setup: add your API token to the config file (get it at https://todoist.com/prefs/integrations):

```toml
[todoist]
api_token = "your-api-token-here"
comment_on_complete = true
```

### Calendar Sync (ICS Export)

Export completed sessions as calendar events:

```bash
# Export all sessions to ICS file
sesh export --format ics

# Export to a specific path
sesh export --format ics --output ~/calendar/sesh.ics

# Export as JSON or CSV
sesh export --format json
sesh export --format csv --output ~/sessions.csv
```

For auto-sync, enable in config — sessions are written to the ICS file automatically after each completion. Subscribe to this file from Google Calendar, Outlook, or any CalDAV client.

```toml
[calendar]
enabled = true
ics_path = "~/.local/share/sesh/sesh.ics"
auto_export = true
```

## Configuration

Config file: `~/.config/sesh/config.toml`

```toml
[general]
theme = "dark"
tick_rate_ms = 250

[timer]
focus_duration = 25        # minutes
short_break_duration = 5   # minutes
long_break_duration = 20   # minutes
long_break_after = 100     # cumulative minutes before long break

[todoist]
api_token = ""             # Todoist API token
comment_on_complete = true # Add comment to task after session

[calendar]
enabled = false            # Enable auto ICS export
ics_path = "~/.local/share/sesh/sesh.ics"
auto_export = true         # Write ICS after each session
```

## Data

All session data is stored in SQLite at `~/.local/share/sesh/sessions.db`.

## License

MIT
