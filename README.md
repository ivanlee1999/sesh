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
```

## Data

All session data is stored in SQLite at `~/.local/share/sesh/sessions.db`.

## License

MIT
