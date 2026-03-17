# sesh

A terminal-native Pomodoro focus timer built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

> For developers who live in the terminal.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

## Features

- **Focus timer** — configurable duration (default 25 min), with a Unicode progress ring
- **Overflow mode** — when the timer hits zero, it keeps counting up in yellow so you can finish your thought
- **Break timers** — short (5 min) and long (20 min) breaks
- **Pause / resume** — freeze mid-session without losing time
- **Intention input** — write what you're working on before starting
- **Categories** — colored labels (Development, Writing, Design, etc.) for organizing sessions
- **Session history** — scrollable table of all past sessions
- **Analytics** — today's focus time, session count, streak, 7-day chart, category breakdown
- **Help overlay** — press `?` from anywhere to see all keybindings
- **CLI mode** — `sesh status`, `sesh history`, `sesh analytics` for scripting and status bars
- **Todoist integration** — link sessions to tasks; auto-populate intention and category
- **Calendar sync** — auto-export sessions as ICS after each completion
- **TOML config** — customize durations, theme, tick rate
- **SQLite storage** — all data stored locally in `~/.local/share/sesh/`

## Installation

### go install

```bash
go install github.com/ivanlee1999/sesh@latest
```

### From source

```bash
git clone https://github.com/ivanlee1999/sesh
cd sesh
go build -o sesh .
# move to somewhere on your PATH, e.g.:
mv sesh ~/.local/bin/
```

## Usage

### Interactive TUI

```bash
sesh
```

Launches the full terminal UI with Timer, Analytics, History, and Settings tabs.

### CLI Commands

```bash
sesh status              # Current status (JSON)
sesh status -f human     # Human-readable status

sesh history             # Recent sessions
sesh history -l 20       # Last 20 sessions

sesh analytics           # Today's stats with category breakdown

sesh export --format ics              # Export sessions to ICS calendar file
sesh export --format ics --output ~/cal/sesh.ics
sesh export --format json
sesh export --format csv --output ~/sessions.csv

sesh todoist             # List today's Todoist tasks
```

## Keybindings

Press `?` in the TUI to open the help overlay at any time.

### Global
| Key | Action |
|-----|--------|
| `1` – `4` | Switch tabs (Timer, Analytics, History, Settings) |
| `Tab` | Next tab |
| `?` | Toggle keybinding help overlay |
| `q` | Quit (only when idle) |
| `Ctrl+C` | Force quit |

### Timer — Idle
| Key | Action |
|-----|--------|
| `Enter` | Start focus session |
| `b` | Start short break (5 min) |
| `B` | Start long break (20 min) |
| `i` | Set intention |
| `c` | Pick category |
| `+` / `-` | Adjust duration ±5 min |
| `>` / `<` | Adjust duration ±1 min |

### Timer — Focus
| Key | Action |
|-----|--------|
| `Space` | Pause / resume |
| `f` | Finish session (no notes) |
| `b` | Finish + start short break |
| `x` | Abandon session (5 s undo with `u`) |

### Timer — Overflow
When the focus timer reaches zero it enters **overflow mode**: the timer counts upward in yellow so you can finish your thought.

| Key | Action |
|-----|--------|
| `f` | Finish and enter notes |
| `b` | Finish + start short break |
| `Space` | Pause / resume |
| `x` | Abandon |

### Break
| Key | Action |
|-----|--------|
| `Enter` / `f` | End break and return to idle |

### History
| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |

## Configuration

Config file: `~/.config/sesh/config.toml`

```toml
[general]
theme = "dark"
tick_rate_ms = 250

[timer]
focus_duration = 25        # minutes
short_break_duration = 5
long_break_duration = 20
long_break_after = 100     # cumulative focus minutes before suggesting a long break

[todoist]
api_token = ""             # get yours at https://todoist.com/prefs/integrations
comment_on_complete = true # add a comment to the linked task after each session

[calendar]
enabled = false
ics_path = "~/.local/share/sesh/sesh.ics"
auto_export = true         # write ICS after every session completion
```

## Todoist Integration

Link Pomodoro sessions to Todoist tasks:

```bash
# List today's and overdue tasks
sesh todoist

# Start a session pre-filled with task content + matched category
sesh start --todoist <task_id>
```

Setup: add your API token to `~/.config/sesh/config.toml` (see config example above). The intention is auto-populated from the task title and the category is matched from the Todoist project name.

## Calendar Sync

Sessions are exported as calendar events (VEVENT) in an ICS file. Subscribe to the file from Google Calendar, Outlook, or any CalDAV client for automatic sync.

Enable auto-export in config:

```toml
[calendar]
enabled = true
ics_path = "~/.local/share/sesh/sesh.ics"
auto_export = true
```

Or export on demand:

```bash
sesh export --format ics --output ~/calendar/sesh.ics
```

## Data

All session data is stored in SQLite at `~/.local/share/sesh/sessions.db`.

## License

MIT
