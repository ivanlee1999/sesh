# sesh — Design Document

> A terminal-native reimplementation of [Session](https://www.stayinsession.com) (Pomodoro Focus Timer) for developers who live in the terminal.

**Version:** 0.1.0 (Draft)
**Date:** 2026-03-17
**Status:** Design Phase

---

## Table of Contents

1. [Product Vision](#1-product-vision)
2. [Feature Scope](#2-feature-scope)
3. [Tech Stack](#3-tech-stack)
4. [UI/UX Design](#4-uiux-design)
5. [State Machine](#5-state-machine)
6. [Data Model](#6-data-model)
7. [Architecture](#7-architecture)
8. [Integrations](#8-integrations)
9. [CLI Interface](#9-cli-interface)
10. [Configuration](#10-configuration)
11. [Keybindings](#11-keybindings)
12. [Color Themes](#12-color-themes)
13. [Sound & Notifications](#13-sound--notifications)
14. [Distribution](#14-distribution)

---

## 1. Product Vision

### 1.1 Why a TUI?

Session is an excellent Pomodoro timer — but it's Apple-only, GUI-only, and subscription-based. Developers who spend their day in terminals, SSH sessions, and tmux panes have no equivalent. sesh fills this gap.

### 1.2 Target Audience

- **Terminal-native developers** — engineers who live in neovim/emacs, tmux/zellij, and prefer keyboard-driven workflows
- **Remote/SSH workers** — developers who SSH into dev machines and want a focus timer alongside their work
- **Linux/BSD users** — entirely excluded from the original Session app
- **Self-hosters** — developers who want local-first data with no subscription or cloud dependency
- **ADHD developers** — who benefit from external structure but don't want to leave their terminal context

### 1.3 Design Principles

| Principle | Description |
|-----------|-------------|
| **Terminal-native** | First-class citizen in tmux/zellij panes, not a GUI pretending to be a TUI |
| **Keyboard-first** | Every action reachable via keyboard; mouse optional but supported |
| **Local-first** | All data in SQLite on disk. No account, no cloud, no subscription |
| **Composable** | Plays well with unix pipes, shell scripts, cron, tmux hooks |
| **Minimal friction** | One keystroke to start a session from any screen |
| **Flexible Pomodoro** | Overflow, custom durations, end-early — same philosophy as original |
| **Sensory feedback** | Terminal bell, desktop notifications, tmux status, sound where available |

### 1.4 Non-Goals

- **Not a task manager** — no built-in todo lists (use taskwarrior, todoist-cli, etc.)
- **Not a website/app blocker** — OS-level blocking from a TUI is fragile and platform-specific; instead, we provide hooks so users can wire up their own (e.g., `/etc/hosts` manipulation, cold turkey CLI)
- **Not a sync service** — no server component. Users who want sync can use Syncthing/Dropbox on the SQLite file
- **Not a replacement for the original** — we target a different audience and make different trade-offs

---

## 2. Feature Scope

### 2.1 MVP (v0.1.0) — Core Timer

| Feature | Description | Maps to Original |
|---------|-------------|-----------------|
| Focus timer | Countdown with configurable duration (default 25m) | Core timer |
| Break timer | Short break (5m) and long break (20m) | Core timer |
| Overflow mode | Timer continues past 0, visual indicator changes | Overflow |
| Intention input | Free-text field for what you're working on | Intention field |
| Categories | Named + colored labels for work types | Categories |
| Session history | Scrollable list of past sessions | Analytics |
| Basic analytics | Today's focus time, session count, streak | Analytics |
| Pause/resume | Freeze timer mid-session | Pause state |
| CLI mode | `session start`, `session status`, `session stop` | URL scheme |
| TOML config | Durations, colors, keybindings | Settings |
| Desktop notifications | Via `notify-send`/`osascript`/`terminal-notifier` | Notifications |

### 2.2 v1.0.0 — Full Experience

| Feature | Description | Maps to Original |
|---------|-------------|-----------------|
| Profiles | Per-category settings (durations, sounds, hooks) | Profiles |
| Breathing exercise | ASCII breathing animation before focus | Breathing phase |
| Post-session notes | Reflection prompt after each session | Reflection |
| Weekly/monthly analytics | Bar charts, category breakdown, heatmaps | Analytics dashboard |
| tmux integration | Status line segment, auto-rename window | Live Activities |
| Sound system | Terminal bell + optional audio via `afplay`/`paplay` | Background sounds |
| Session presence | Periodic gentle reminder (bell/notification) | Session presence |
| Pre-end warning | Alert N minutes before timer expires | Pre-end notification |
| CSV/JSON export | Export session history | Data export |
| Shell hooks | Run arbitrary commands on state transitions | AppleScript/Shortcuts |
| Auto-suggest | Past intentions as completions | Auto-suggest |
| Mouse support | Click timer, scroll history, click categories | — |

### 2.3 Future (v2.0+)

| Feature | Description | Maps to Original |
|---------|-------------|-----------------|
| Slack integration | Update status + mute via API token | Slack blocker |
| Calendar sync | Read/write `.ics` files or Google Calendar API | Calendar integration |
| Blocker hooks | Pre-built hook scripts for `/etc/hosts`, Cold Turkey CLI, etc. | Website/App blocker |
| Remote sync | Optional SQLite replication via LiteStream or Syncthing guide | Cross-device sync |
| Daemon mode | Background process with socket IPC for tmux/polybar/waybar | Menu bar icon |
| Web dashboard | Localhost web UI for rich analytics (Plotly/Chart.js) | Analytics |
| Neovim plugin | Lua plugin that talks to daemon socket | Chrome Extension |
| Waybar/Polybar module | Status bar integration for tiling WM users | Menu bar icon |
| Zellij plugin | Native WASM plugin for Zellij | — |

---

## 3. Tech Stack

### 3.1 Language: Rust

**Rationale:**

| Factor | Rust | Python (textual) | Go (bubbletea) |
|--------|------|-------------------|-----------------|
| **Startup time** | ~5ms | ~200ms | ~10ms |
| **Binary size** | ~5MB static | Requires runtime | ~8MB static |
| **TUI ecosystem** | Ratatui (mature, active) | Textual (good, CSS-based) | Bubbletea (good, Elm-like) |
| **Cross-platform** | Excellent | Excellent | Excellent |
| **Async support** | Tokio (mature) | asyncio (mature) | goroutines (native) |
| **Distribution** | `cargo install`, single binary | pip, needs Python | `go install`, single binary |
| **Type safety** | Strong (enums for states) | Weak | Moderate |
| **Memory safety** | Guaranteed | GC | GC |
| **Community** | Large TUI community | Growing | Active |

Rust wins on startup time (critical for CLI subcommands), single-binary distribution, type safety for state machines, and the maturity of the Ratatui ecosystem.

### 3.2 Core Dependencies

```toml
[dependencies]
# TUI
ratatui = "0.29"              # Terminal UI framework
crossterm = "0.28"             # Terminal backend (cross-platform)

# Async runtime
tokio = { version = "1", features = ["full"] }

# Data
rusqlite = { version = "0.32", features = ["bundled"] } # SQLite
chrono = { version = "0.4", features = ["serde"] }       # Date/time

# Config & serialization
serde = { version = "1", features = ["derive"] }
toml = "0.8"                   # Config format
serde_json = "1"               # JSON export
csv = "1"                      # CSV export

# CLI
clap = { version = "4", features = ["derive"] } # Argument parsing

# Notifications
notify-rust = "4"              # Desktop notifications (Linux/macOS)

# Audio (optional)
rodio = { version = "0.19", optional = true }  # Audio playback

# Misc
directories = "5"              # XDG paths
uuid = { version = "1", features = ["v4", "serde"] }
thiserror = "2"                # Error types
tracing = "0.1"                # Structured logging
tracing-subscriber = "0.3"
```

### 3.3 Directory Structure

```
sesh/
├── Cargo.toml
├── src/
│   ├── main.rs                # Entry point, CLI parsing
│   ├── app.rs                 # App state, event loop
│   ├── state.rs               # State machine (timer states)
│   ├── timer.rs               # Timer logic (countdown, overflow)
│   ├── db/
│   │   ├── mod.rs             # Database connection, migrations
│   │   ├── schema.rs          # Table definitions
│   │   ├── sessions.rs        # Session CRUD
│   │   ├── categories.rs      # Category CRUD
│   │   └── profiles.rs        # Profile CRUD
│   ├── ui/
│   │   ├── mod.rs             # UI router (which screen to render)
│   │   ├── timer.rs           # Timer screen
│   │   ├── intention.rs       # Intention input overlay
│   │   ├── analytics.rs       # Analytics screen
│   │   ├── history.rs         # Session history screen
│   │   ├── settings.rs        # Settings screen
│   │   ├── breathing.rs       # Breathing exercise animation
│   │   ├── category.rs        # Category picker
│   │   └── widgets/
│   │       ├── clock.rs       # ASCII clock widget
│   │       ├── bar_chart.rs   # Analytics bar chart
│   │       ├── heatmap.rs     # Contribution heatmap
│   │       └── status_bar.rs  # Bottom status bar
│   ├── config/
│   │   ├── mod.rs             # Config loading, defaults
│   │   ├── theme.rs           # Color theme definitions
│   │   └── keybindings.rs     # Keybinding definitions
│   ├── hooks.rs               # Shell hook execution
│   ├── notifications.rs       # Desktop notification dispatch
│   ├── sound.rs               # Audio playback (optional)
│   ├── export.rs              # CSV/JSON export
│   ├── integrations/
│   │   ├── tmux.rs            # tmux status line integration
│   │   └── slack.rs           # Slack status (future)
│   └── cli.rs                 # Non-interactive CLI commands
├── assets/
│   └── sounds/                # Bundled notification sounds
├── config/
│   └── default.toml           # Default configuration
└── docs/
    └── sesh-design-document.md
```

---

## 4. UI/UX Design

### 4.1 Screen Architecture

sesh uses a single-window, multi-screen architecture with a persistent status bar.

```
┌─────────────────────────────────────────────────────────────────────┐
│ SESSION TUI                                    [Tab1] [Tab2] [Tab3]│
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                        ACTIVE SCREEN                                │
│                   (Timer / Analytics /                               │
│                    History / Settings)                               │
│                                                                     │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ ⏱ 18:42 FOCUS │ ▶ Coding sesh │ 🏷 Development │ ?:help    │
└─────────────────────────────────────────────────────────────────────┘
```

**Screens (navigable via number keys or Tab):**

| Key | Screen | Description |
|-----|--------|-------------|
| `1` | Timer | Main timer display + controls |
| `2` | Analytics | Charts, stats, category breakdown |
| `3` | History | Scrollable session log |
| `4` | Settings | Configuration editor |

### 4.2 Timer Screen — Idle State

```
┌─────────────────────────────────────────────────────────────────────┐
│ TIMER                                    Analytics  History  Config │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                          ╭────────────╮                             │
│                     ╭────╯            ╰────╮                        │
│                   ╭─╯        25:00         ╰─╮                      │
│                  ╭╯                           ╰╮                    │
│                  │                             │                    │
│                  │            ●                │                    │
│                  │           /                 │                    │
│                  │          /                  │                    │
│                  ╰╮        ○                  ╭╯                    │
│                   ╰─╮                      ╭─╯                     │
│                     ╰────╮            ╭────╯                        │
│                          ╰────────────╯                             │
│                                                                     │
│                  Intention: _                                       │
│                  Category:  [Development     ▾]                     │
│                                                                     │
│            ┌──────────────────┐   ┌──────────────────┐              │
│            │   Start Focus    │   │   Start Break    │              │
│            │     (enter)      │   │       (b)        │              │
│            └──────────────────┘   └──────────────────┘              │
│                                                                     │
│                  [+] / [-]  Adjust duration (5m steps)              │
│                  [<] / [>]  Adjust duration (1m steps)              │
│                                                                     │
│  Today: 2h 35m focused │ 6 sessions │ Streak: 3 days              │
├─────────────────────────────────────────────────────────────────────┤
│ IDLE │ Press Enter to start │ i:intention  c:category  ?:help      │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.3 Timer Screen — Focus State

```
┌─────────────────────────────────────────────────────────────────────┐
│ TIMER                                    Analytics  History  Config │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                          ╭────────────╮                             │
│                     ╭────╯   ▓▓▓▓▓▓   ╰────╮                       │
│                   ╭─╯  ▓▓▓  18:42  ▓▓▓  ╰─╮                       │
│                  ╭╯  ▓▓▓               ▓▓▓  ╰╮                     │
│                  │  ▓▓▓                 ▓▓▓  │                     │
│                  │ ▓▓▓        ●         ▓▓▓  │                     │
│                  │ ▓▓▓      ╱           ▓▓▓  │                     │
│                  ╰╮ ▓▓▓   ╱            ▓▓▓ ╭╯                     │
│                   ╰─╮ ▓▓▓○          ▓▓▓ ╭─╯                       │
│                     ╰────╮  ▓▓▓▓▓▓  ╭────╯                         │
│                          ╰────────────╯                             │
│                                                                     │
│                  ▸ Coding sesh                               │
│                    Development                                      │
│                                                                     │
│                  ██████████████████░░░░░░░░  74%                    │
│                                                                     │
│                  Started: 14:32  │  Elapsed: 6:18                   │
│                  Pauses: 0       │  Target:  25:00                  │
│                                                                     │
│  Today: 2h 35m focused │ 6 sessions │ Streak: 3 days              │
├─────────────────────────────────────────────────────────────────────┤
│ ⏱ 18:42 FOCUS │ space:pause  f:finish  b:break  x:abandon  ?:help │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.4 Timer Screen — Overflow State

The background color shifts to the category color (e.g., blue for Development).

```
┌─────────────────────────────────────────────────────────────────────┐
│ TIMER  ◆ OVERFLOW                        Analytics  History  Config │
├─────────────────────────────────────────────────────────────────────┤
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░╭────────────╮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░╭────╯            ╰────╮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░╭─╯       +2:14          ╰─╮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░╭╯                           ╰╮░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░│                             │░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░│            ●                │░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░│             ╲               │░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░╰╮              ╲            ╭╯░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░╰─╮              ○       ╭─╯░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░╰────╮            ╭────╯░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░╰────────────╯░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░▸ Coding sesh░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░ Development░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░████████████████████████████  +2:14░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░Target: 25:00 │ Overflow: +2:14░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░Total:  27:14 │ You're in the zone!░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
├─────────────────────────────────────────────────────────────────────┤
│ ⏱ +2:14 OVERFLOW │ f:finish  b:break  x:abandon  ?:help           │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.5 Breathing Exercise Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│ BREATHE                                                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
│                            ╭──────╮                                 │
│                         ╭──╯      ╰──╮                              │
│                        ╭╯   ○    ○   ╰╮                             │
│                        │              │                             │
│                        │    breathe   │                             │
│                        │      in      │                             │
│                        ╰╮            ╭╯                             │
│                         ╰──╮      ╭──╯                              │
│                            ╰──────╯                                 │
│                                                                     │
│                         ████░░░░░░░░                                │
│                           3 / 8 seconds                             │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ BREATHING │ Esc:skip                                                │
└─────────────────────────────────────────────────────────────────────┘
```

The circle expands on inhale and contracts on exhale using different box-drawing sizes:

```
INHALE (expanding):       HOLD:                   EXHALE (contracting):

     ╭──╮                  ╭──────────╮                  ╭────╮
   ╭─╯  ╰─╮            ╭──╯          ╰──╮             ╭─╯    ╰─╮
   │  in   │            │     hold       │             │  out    │
   ╰─╮  ╭─╯            ╰──╮          ╭──╯             ╰─╮    ╭─╯
     ╰──╯                  ╰──────────╯                  ╰────╯
```

### 4.6 Intention Input Overlay

A modal overlay that appears when pressing `i` or before starting a session:

```
┌─────────────────────────────────────────────────────────────────────┐
│ TIMER                                    Analytics  History  Config │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│          ┌─────────────────────────────────────────┐                │
│          │  What are you working on?               │                │
│          │                                         │                │
│          │  > Coding sesh timer wid█        │                │
│          │                                         │                │
│          │  Recent:                                 │                │
│          │    ▸ Coding sesh                  │                │
│          │      Reviewing PR #142                   │                │
│          │      Writing design document             │                │
│          │      Debugging sync issue                │                │
│          │                                         │                │
│          │  Category: [Development     ▾]          │                │
│          │                                         │                │
│          │  Enter:confirm  Tab:category  Esc:close │                │
│          └─────────────────────────────────────────┘                │
│                                                                     │
│                                                                     │
│                                                                     │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ IDLE │ Editing intention...                                        │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.7 Category Picker

```
┌─────────────────────────────────────┐
│  Select Category                    │
│                                     │
│  > ● Development          #61AFEF  │
│    ● Writing              #E06C75  │
│    ● Design               #C678DD  │
│    ● Research             #E5C07B  │
│    ● Meeting              #56B6C2  │
│    ● Exercise             #98C379  │
│    ● Reading              #D19A66  │
│    ● Admin                #ABB2BF  │
│                                     │
│  n:new  e:edit  d:delete  Esc:close │
└─────────────────────────────────────┘
```

### 4.8 Analytics Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│ Timer  ANALYTICS                             History        Config  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ◀ This Week (Mar 10 - Mar 16, 2026) ▶          [d]aily [w]eekly  │
│                                                  [m]onthly         │
│  Focus Time by Day                                                  │
│  ┌────────────────────────────────────────────────────────────┐     │
│  │                                     ██                     │     │
│  │                           ██        ██                     │ 6h  │
│  │              ██           ██  ██    ██                     │     │
│  │         ██   ██     ██    ██  ██    ██  ██                 │ 4h  │
│  │    ██   ██   ██     ██    ██  ██    ██  ██                 │     │
│  │    ██   ██   ██     ██    ██  ██    ██  ██                 │ 2h  │
│  │    ██   ██   ██     ██    ██  ██    ██  ██                 │     │
│  └────────────────────────────────────────────────────────────┘     │
│    Mon   Tue   Wed   Thu    Fri  Sat   Sun                          │
│                                                                     │
│  Category Breakdown              │  Summary                         │
│  ┌─────────────────────────┐     │  ──────────────────────          │
│  │ ████████████░░░░ 52%    │     │  Total Focus:   24h 15m         │
│  │ Development             │     │  Sessions:      48               │
│  │ ██████░░░░░░░░░░ 25%    │     │  Avg/Day:       3h 28m          │
│  │ Writing                 │     │  Best Day:      Fri (6h 10m)    │
│  │ ████░░░░░░░░░░░░ 15%    │     │  Current Streak: 12 days       │
│  │ Research                │     │  Longest Streak: 31 days        │
│  │ ██░░░░░░░░░░░░░░  8%    │     │                                 │
│  │ Other                   │     │  Today: 2h 35m (6 sessions)     │
│  └─────────────────────────┘     │                                  │
│                                                                     │
│  Contribution Heatmap (Last 52 Weeks)                               │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ Mon ░░▒▒▓▓██▒▒░░▒▒▓▓██████▓▓▒▒░░▒▒▓▓████▓▓▒▒░░▒▒▓▓████▓▓ │   │
│  │ Wed ▒▒▓▓██▓▓▒▒░░▒▒▓▓██████████▓▓▒▒░░▒▒▓▓██████▓▓▒▒▒▒▓▓██ │   │
│  │ Fri ░░▒▒▒▒▓▓▒▒░░▒▒▓▓████▓▓▓▓▒▒░░░░▒▒▓▓████████▓▓▒▒▒▒▓▓▓▓ │   │
│  └──────────────────────────────────────────────────────────────┘   │
│    Apr May Jun Jul Aug Sep Oct Nov Dec Jan Feb Mar                   │
│    ░<1h  ▒1-2h  ▓2-4h  █4h+                                        │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ IDLE │ d/w/m:period  ◀/▶:navigate  e:export  ?:help                │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.9 History Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│ Timer  Analytics  HISTORY                                   Config  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Filter: [All categories ▾]  [All types ▾]  Search: ___________    │
│                                                                     │
│  ── Today, March 17 ──────────────────────── Total: 2h 35m ──────  │
│                                                                     │
│  > ● 14:32 - 15:02  Coding sesh        Development   30:00  │
│    ● 13:15 - 13:42  Reviewing PR #142          Development   27:00  │
│    ● 11:00 - 11:25  Writing design doc         Writing       25:00  │
│    ● 10:00 - 10:25  Email + admin              Admin         25:00  │
│    ● 09:05 - 09:38  Debugging sync issue       Development   33:00  │
│    ● 08:30 - 08:55  Reading chapter 12         Reading       25:00  │
│                                                                     │
│  ── Yesterday, March 16 ─────────────────── Total: 3h 45m ──────  │
│                                                                     │
│    ● 16:00 - 16:50  Writing blog post          Writing       50:00  │
│    ● 14:00 - 14:25  Code review                Development   25:00  │
│    ● ...                                                            │
│                                                                     │
│  ┌─ Session Detail ────────────────────────────────────────────┐    │
│  │ Coding sesh                                          │    │
│  │ Category: Development  │  Duration: 30:00  │  Pauses: 1    │    │
│  │ Started: 14:32  │  Ended: 15:02                             │    │
│  │ Type: Full Focus                                            │    │
│  │ Notes: Implemented clock widget rendering. Need to add      │    │
│  │        overflow color logic next.                           │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ IDLE │ ↑/↓:select  Enter:details  d:delete  e:export  ?:help      │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.10 Settings Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│ Timer  Analytics  History                                  SETTINGS │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐                                                   │
│  │ > General    │  General Settings                                 │
│  │   Timer      │  ────────────────────────────────────────         │
│  │   Sounds     │                                                   │
│  │   Hooks      │  Theme              [dark            ▾]           │
│  │   Categories │  Show breathing     [yes             ▾]           │
│  │   Profiles   │  Prompt for notes   [yes             ▾]           │
│  │   Tmux       │  Auto-suggest       [yes             ▾]           │
│  │   Export     │  Data directory     ~/.local/share/sesh/   │
│  │   About      │  Log level          [warn            ▾]           │
│  │              │                                                   │
│  │              │  Timer Settings                                   │
│  │              │  (change in [Timer] tab for profile-specific)     │
│  │              │  ────────────────────────────────────────         │
│  │              │  Default focus       25m                          │
│  │              │  Default short break  5m                          │
│  │              │  Default long break  20m                          │
│  │              │  Long break after    100m (cumulative focus)      │
│  │              │  Session presence    10m (0 to disable)           │
│  │              │  Pre-end warning      2m (0 to disable)           │
│  │              │                                                   │
│  └──────────────┘  [Save]  [Reset to Defaults]                     │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ IDLE │ Tab:switch pane  ↑/↓:navigate  Enter:edit  ?:help           │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.11 Post-Session Reflection Overlay

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│          ┌─────────────────────────────────────────┐                │
│          │  Session Complete!                       │                │
│          │                                         │                │
│          │  ▸ Coding sesh                   │                │
│          │    Development │ 27:14 (25:00 + 2:14)  │                │
│          │                                         │                │
│          │  Notes (optional):                      │                │
│          │  ┌───────────────────────────────────┐  │                │
│          │  │ Implemented clock widget.         │  │                │
│          │  │ Need to add overflow colors next. │  │                │
│          │  │ █                                 │  │                │
│          │  └───────────────────────────────────┘  │                │
│          │                                         │                │
│          │  Enter:save  Esc:skip  b:start break   │                │
│          └─────────────────────────────────────────┘                │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ FINISHED │ Enter:save notes  Esc:dismiss  b:start break            │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.12 Minimal / Pane Mode

For use in a small tmux pane (minimum 40x10):

```
┌──────────────────────────────┐
│     ╭────╮                   │
│   ╭─╯    ╰─╮                │
│   │  18:42  │  FOCUS         │
│   ╰─╮    ╭─╯                │
│     ╰────╯                   │
│  ▸ Coding sesh        │
│  ██████████████░░░░░░  74%   │
│  space:pause f:finish b:break│
└──────────────────────────────┘
```

Ultra-minimal (20x3, for status-line embedding):

```
⏱ 18:42 FOCUS ██████░░ 74%
▸ Coding sesh
[Development] space:⏸ f:✓ b:☕
```

---

## 5. State Machine

### 5.1 Full State Diagram

```
                              ┌─────────────┐
                              │             │
                    ┌────────▶│    IDLE     │◀──────────────────┐
                    │         │             │                    │
                    │         └──────┬──────┘                    │
                    │                │                           │
                    │                │ start (Enter)             │
                    │                ▼                           │
                    │         ┌─────────────┐                    │
                    │         │  INTENTION   │──── Esc ──────────┘
                    │         │   INPUT      │
                    │         └──────┬──────┘
                    │                │
                    │                │ confirm (Enter)
                    │                ▼
                    │         ┌─────────────┐
                    │         │  BREATHING   │──── Esc/skip ──┐
                    │         │  EXERCISE    │                 │
                    │         └──────┬──────┘                 │
                    │                │                         │
                    │                │ complete                │
                    │                ▼                         ▼
                    │         ┌─────────────┐          ┌─────────────┐
                    │    ┌───▶│   FOCUS     │────────▶│  OVERFLOW    │
                    │    │    │  (counting  │ timer=0  │  (counting   │
                    │    │    │    down)    │          │     up)      │
                    │    │    └──────┬──────┘          └──────┬──────┘
                    │    │           │                         │
                    │    │     ┌─────┼─────────────────────────┤
                    │    │     │     │                         │
                    │    │     │  ┌──▼──────────┐              │
                    │    │     │  │   PAUSED    │              │
                    │    │     │  │  (frozen)   │              │
                    │    │     │  └──┬──────────┘              │
                    │    │     │     │ resume                  │
                    │    │     │     ▼                         │
                    │    │     │  (returns to FOCUS            │
                    │    │     │   or OVERFLOW)                │
                    │    │     │                               │
                    │    │   finish(f)                    finish(f)
                    │    │     │                               │
                    │    │     ▼                               ▼
                    │    │  ┌─────────────┐                    │
                    │    │  │  REFLECTION │◀───────────────────┘
                    │    │  │  (notes)    │
                    │    │  └──────┬──────┘
                    │    │         │
                    │    │    save/skip
                    │    │         │
                    │    │         ▼
                    │    │  ┌─────────────┐
                    │    │  │   BREAK     │
                    │    │  │  (counting  │
                    │    │  │    down)    │
                    │    │  └──────┬──────┘
                    │    │         │
                    │    │         │ timer=0
                    │    │         ▼
                    │    │  ┌─────────────┐
                    │    │  │   BREAK     │
                    │    │  │  OVERFLOW   │
                    │    │  └──────┬──────┘
                    │    │         │
                    │    │    start focus
                    │    │         │
                    │    └─────────┘
                    │
                    │ abandon (x) — from any active state
                    │
              ┌─────┴──────┐
              │  ABANDONED  │──── undo(u) ──▶ (previous state)
              │  (5s undo)  │──── timeout ──▶ IDLE
              └────────────┘
```

### 5.2 State Definitions

```rust
enum TimerState {
    Idle,
    IntentionInput,
    Breathing { elapsed: Duration, phase: BreathPhase },
    Focus { remaining: Duration, started_at: Instant, pauses: Vec<PauseRecord> },
    Overflow { elapsed: Duration, target_was: Duration, started_at: Instant },
    Paused { inner: Box<TimerState>, paused_at: Instant },
    Break { remaining: Duration, break_type: BreakType, started_at: Instant },
    BreakOverflow { elapsed: Duration, break_type: BreakType },
    Reflection { session: CompletedSession },
    Abandoned { previous: Box<TimerState>, undo_deadline: Instant },
}

enum BreathPhase { Inhale, Hold, Exhale }
enum BreakType { Short, Long }

struct PauseRecord {
    paused_at: DateTime<Utc>,
    resumed_at: Option<DateTime<Utc>>,
}
```

### 5.3 Transition Rules

| From | Event | To | Side Effects |
|------|-------|----|-------------|
| Idle | `Enter` | IntentionInput | — |
| Idle | `b` | Break(Short) | Run `break_start` hook |
| IntentionInput | `Enter` | Breathing (if enabled) or Focus | — |
| IntentionInput | `Esc` | Idle | — |
| Breathing | complete | Focus | Run `session_start` hook, start blockers |
| Breathing | `Esc` | Focus | Skip animation |
| Focus | tick (remaining=0) | Overflow | Change visuals, play overflow sound |
| Focus | `Space` | Paused(Focus) | Run `session_pause` hook |
| Focus | `f` | Reflection | Run `session_end` hook, stop blockers |
| Focus | `b` | Reflection→Break | Finish + immediate break |
| Focus | `x` | Abandoned | — |
| Overflow | `f` | Reflection | Run `session_end` hook |
| Overflow | `b` | Reflection→Break | — |
| Overflow | `x` | Abandoned | — |
| Paused | `Space` | (inner state) | Run `session_resume` hook |
| Paused | `f` | Reflection | — |
| Paused | `x` | Abandoned | — |
| Reflection | `Enter`/`Esc` | Break or Idle | Save session to DB |
| Break | tick (remaining=0) | BreakOverflow | Change visuals, notify |
| Break | `Enter`/`f` | Idle or Focus | — |
| BreakOverflow | `Enter` | Focus (new session) | — |
| Abandoned | `u` (within 5s) | (previous state) | Restore |
| Abandoned | timeout | Idle | Discard session |

### 5.4 Long Break Logic

Long break triggers based on **cumulative focus time** (not session count), matching the original app's v2.9+ behavior:

```
if cumulative_focus_since_last_long_break >= config.long_break_after {
    next_break = BreakType::Long;
    reset cumulative counter;
} else {
    next_break = BreakType::Short;
}
```

---

## 6. Data Model

### 6.1 SQLite Schema

```sql
-- Database: ~/.local/share/sesh/sessions.db

-- Tracks schema version for migrations
CREATE TABLE IF NOT EXISTS schema_version (
    version     INTEGER NOT NULL,
    applied_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Categories for organizing sessions
CREATE TABLE IF NOT EXISTS categories (
    id          TEXT PRIMARY KEY,              -- UUID v4
    title       TEXT NOT NULL,
    hex_color   TEXT NOT NULL DEFAULT '#61AFEF', -- 6-char hex
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Core session records
CREATE TABLE IF NOT EXISTS sessions (
    id              TEXT PRIMARY KEY,          -- UUID v4
    title           TEXT NOT NULL DEFAULT '',  -- intention text
    category_id     TEXT REFERENCES categories(id) ON DELETE SET NULL,
    session_type    TEXT NOT NULL CHECK (session_type IN (
                        'full_focus',          -- completed target duration
                        'partial_focus',       -- ended before target
                        'rest',                -- break session
                        'abandoned'            -- cancelled (shouldn't normally appear)
                    )),
    target_seconds  INTEGER NOT NULL,          -- planned duration
    actual_seconds  INTEGER NOT NULL,          -- actual elapsed (includes overflow)
    pause_seconds   INTEGER NOT NULL DEFAULT 0,-- total time paused
    overflow_seconds INTEGER NOT NULL DEFAULT 0,-- time past target (0 if ended early)
    started_at      TEXT NOT NULL,             -- ISO 8601
    ended_at        TEXT NOT NULL,             -- ISO 8601
    notes           TEXT,                      -- post-session reflection
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Pause records within a session
CREATE TABLE IF NOT EXISTS pauses (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    paused_at   TEXT NOT NULL,
    resumed_at  TEXT,                          -- NULL if session ended while paused
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Per-category/intention setting overrides
CREATE TABLE IF NOT EXISTS profiles (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL,
    category_id         TEXT REFERENCES categories(id) ON DELETE CASCADE,
    intention_keyword   TEXT,                  -- match against intention text
    focus_minutes       INTEGER,               -- NULL = use default
    short_break_minutes INTEGER,
    long_break_minutes  INTEGER,
    long_break_after_minutes INTEGER,
    presence_minutes    INTEGER,
    pre_end_minutes     INTEGER,
    hooks_json          TEXT,                  -- JSON: { "session_start": "...", ... }
    created_at          TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at          TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_sessions_category   ON sessions(category_id);
CREATE INDEX IF NOT EXISTS idx_sessions_type       ON sessions(session_type);
CREATE INDEX IF NOT EXISTS idx_pauses_session      ON pauses(session_id);
CREATE INDEX IF NOT EXISTS idx_profiles_category   ON profiles(category_id);
```

### 6.2 Query Patterns

```sql
-- Today's focus time
SELECT COALESCE(SUM(actual_seconds - pause_seconds), 0) / 60.0 AS focus_minutes
FROM sessions
WHERE session_type IN ('full_focus', 'partial_focus')
  AND date(started_at) = date('now');

-- Weekly breakdown by category
SELECT
    c.title,
    c.hex_color,
    strftime('%w', s.started_at) AS day_of_week,
    SUM(s.actual_seconds - s.pause_seconds) / 3600.0 AS hours
FROM sessions s
LEFT JOIN categories c ON s.category_id = c.id
WHERE s.session_type IN ('full_focus', 'partial_focus')
  AND s.started_at >= date('now', 'weekday 1', '-7 days')
GROUP BY c.id, day_of_week;

-- Current streak (consecutive days with at least one focus session)
WITH RECURSIVE dates AS (
    SELECT date('now') AS d
    UNION ALL
    SELECT date(d, '-1 day') FROM dates
    WHERE EXISTS (
        SELECT 1 FROM sessions
        WHERE session_type IN ('full_focus', 'partial_focus')
          AND date(started_at) = date(d, '-1 day')
    )
)
SELECT COUNT(*) - 1 AS streak FROM dates
WHERE EXISTS (
    SELECT 1 FROM sessions
    WHERE session_type IN ('full_focus', 'partial_focus')
      AND date(started_at) = d
);

-- Contribution heatmap data (last 365 days)
SELECT
    date(started_at) AS day,
    SUM(actual_seconds - pause_seconds) / 3600.0 AS hours
FROM sessions
WHERE session_type IN ('full_focus', 'partial_focus')
  AND started_at >= date('now', '-365 days')
GROUP BY day;

-- Auto-suggest: recent unique intentions
SELECT DISTINCT title
FROM sessions
WHERE title != ''
ORDER BY started_at DESC
LIMIT 20;
```

### 6.3 Migration Strategy

Migrations stored as numbered SQL files and tracked via `schema_version` table:

```
migrations/
  001_initial.sql
  002_add_profiles.sql
  003_add_session_notes.sql
```

Applied on startup:

```rust
fn run_migrations(conn: &Connection) -> Result<()> {
    let current = get_schema_version(conn)?; // 0 if fresh
    for migration in MIGRATIONS.iter().skip(current) {
        conn.execute_batch(migration.sql)?;
        set_schema_version(conn, migration.version)?;
    }
    Ok(())
}
```

---

## 7. Architecture

### 7.1 Component Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                        sesh                            │
│                                                              │
│  ┌──────────┐     ┌──────────┐     ┌──────────────────┐     │
│  │   CLI    │     │   TUI    │     │    Daemon (v2)   │     │
│  │  Parser  │     │  App     │     │   (future)       │     │
│  │ (clap)   │     │  Loop    │     │                  │     │
│  └────┬─────┘     └────┬─────┘     └──────────────────┘     │
│       │                │                                     │
│       ▼                ▼                                     │
│  ┌─────────────────────────────────────────────────────┐     │
│  │                    Core Engine                       │     │
│  │                                                     │     │
│  │  ┌───────────┐  ┌──────────┐  ┌─────────────────┐  │     │
│  │  │   State   │  │  Timer   │  │   Notification  │  │     │
│  │  │  Machine  │  │  Engine  │  │    Dispatcher   │  │     │
│  │  └───────────┘  └──────────┘  └─────────────────┘  │     │
│  │                                                     │     │
│  │  ┌───────────┐  ┌──────────┐  ┌─────────────────┐  │     │
│  │  │   Hook    │  │  Config  │  │    Analytics    │  │     │
│  │  │  Runner   │  │  Manager │  │    Engine       │  │     │
│  │  └───────────┘  └──────────┘  └─────────────────┘  │     │
│  │                                                     │     │
│  └──────────────────────┬──────────────────────────────┘     │
│                         │                                    │
│                         ▼                                    │
│  ┌──────────────────────────────────────────────────────┐    │
│  │                  Storage Layer                        │    │
│  │                                                      │    │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │    │
│  │  │  SQLite  │  │  Config  │  │  Export (CSV/JSON) │  │    │
│  │  │   DB     │  │  (TOML)  │  │                   │  │    │
│  │  └──────────┘  └──────────┘  └───────────────────┘  │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐    │
│  │                  Integrations                         │    │
│  │                                                      │    │
│  │  ┌──────┐  ┌────────┐  ┌───────┐  ┌──────────────┐  │    │
│  │  │ tmux │  │ notify │  │ sound │  │ Slack (v2)   │  │    │
│  │  └──────┘  └────────┘  └───────┘  └──────────────┘  │    │
│  └──────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

### 7.2 Event Loop

The TUI uses a standard Ratatui event loop with a tick-based timer:

```rust
// Simplified event loop
async fn run(terminal: &mut Terminal, app: &mut App) -> Result<()> {
    let tick_rate = Duration::from_millis(250); // 4 Hz for smooth countdown
    let mut last_tick = Instant::now();

    loop {
        // 1. Render current state
        terminal.draw(|frame| ui::render(frame, app))?;

        // 2. Poll for input events (non-blocking with timeout)
        let timeout = tick_rate.saturating_sub(last_tick.elapsed());
        if crossterm::event::poll(timeout)? {
            match crossterm::event::read()? {
                Event::Key(key) => app.handle_key(key),
                Event::Mouse(mouse) => app.handle_mouse(mouse),
                Event::Resize(w, h) => app.handle_resize(w, h),
                _ => {}
            }
        }

        // 3. Timer tick (every tick_rate)
        if last_tick.elapsed() >= tick_rate {
            app.tick(); // Updates timer, checks transitions, fires hooks
            last_tick = Instant::now();
        }

        // 4. Check for exit
        if app.should_quit {
            break;
        }
    }
    Ok(())
}
```

### 7.3 App State Structure

```rust
struct App {
    // Core state
    timer_state: TimerState,
    current_screen: Screen,
    intention: String,
    selected_category: Option<Category>,

    // Data
    db: Database,
    config: Config,
    categories: Vec<Category>,

    // UI state
    show_overlay: Option<Overlay>,
    history_list_state: ListState,
    analytics_period: AnalyticsPeriod,
    settings_focus: SettingsPane,

    // Integrations
    tmux: Option<TmuxIntegration>,
    notifier: Notifier,
    hook_runner: HookRunner,
    sound_player: Option<SoundPlayer>,

    // Flags
    should_quit: bool,
    cumulative_focus: Duration, // since last long break
}

enum Screen { Timer, Analytics, History, Settings }
enum Overlay { IntentionInput, CategoryPicker, Reflection, Help, Confirm(String) }
enum AnalyticsPeriod { Daily, Weekly, Monthly }
```

### 7.4 Hook System

Hooks are shell commands executed on state transitions:

```rust
struct HookRunner {
    hooks: HashMap<HookEvent, Vec<String>>,
}

enum HookEvent {
    SessionStart,
    SessionEnd,
    SessionPause,
    SessionResume,
    BreakStart,
    BreakEnd,
    Overflow,
    Abandon,
}

impl HookRunner {
    async fn fire(&self, event: HookEvent, context: &HookContext) {
        if let Some(commands) = self.hooks.get(&event) {
            for cmd in commands {
                let expanded = self.expand_variables(cmd, context);
                // Fire and forget — don't block the UI
                tokio::spawn(async move {
                    Command::new("sh")
                        .args(["-c", &expanded])
                        .env("SESSION_INTENTION", &context.intention)
                        .env("SESSION_CATEGORY", &context.category)
                        .env("SESSION_DURATION", &context.duration.to_string())
                        .env("SESSION_STATE", &context.state)
                        .spawn();
                });
            }
        }
    }
}
```

**Available hook variables:**

| Variable | Description |
|----------|-------------|
| `$SESSION_INTENTION` | Current intention text |
| `$SESSION_CATEGORY` | Category name |
| `$SESSION_CATEGORY_COLOR` | Category hex color |
| `$SESSION_DURATION` | Target duration in seconds |
| `$SESSION_ELAPSED` | Elapsed time in seconds |
| `$SESSION_STATE` | Current state name |
| `$SESSION_TYPE` | `focus` or `break` |

---

## 8. Integrations

### 8.1 tmux Integration

sesh can update the tmux status line and window name to reflect the current timer state.

**Status line segment:**

```bash
# In ~/.tmux.conf:
set -g status-right '#(sesh tmux-status)'
```

Output examples:
```
⏱ 18:42 FOCUS [Development]     # During focus
⏸ 18:42 PAUSED                  # Paused
☕ 3:21 BREAK                    # Break
◆ +2:14 OVERFLOW                 # Overflow
                                  # Idle (empty string)
```

**Window rename:**
```toml
# In sesh config:
[integrations.tmux]
enabled = true
status_command = true                # Enable tmux-status subcommand
rename_window = true                 # Rename active window during session
window_format_focus = "⏱ {remaining} {intention}"
window_format_break = "☕ {remaining}"
restore_window_name = true           # Restore original name after session
```

**Implementation:** Writes to a temp file (`/tmp/sesh-status`) on every tick. The `sesh tmux-status` command reads this file (fast, no DB access). tmux refreshes via `status-interval`.

### 8.2 Desktop Notifications

```rust
// Priority-ordered notification backends:
enum NotifyBackend {
    NotifyRust,       // Cross-platform via notify-rust crate
    OsaScript,        // macOS: osascript -e 'display notification ...'
    NotifySend,       // Linux: notify-send
    TerminalBell,     // Fallback: \x07
}
```

**Notification triggers:**

| Event | Title | Body | Sound |
|-------|-------|------|-------|
| Timer end | "Focus Complete" | "25:00 — {intention}" | Yes |
| Pre-end warning | "2 minutes left" | "{intention}" | Optional |
| Break end | "Break Over" | "Time to focus!" | Yes |
| Overflow start | "Overflow" | "Target reached, keep going!" | Subtle |
| Session presence | "Still focusing?" | "{elapsed} elapsed" | Bell |

### 8.3 Slack Integration (v2.0)

```toml
[integrations.slack]
enabled = true
token = "xoxp-..."                    # User OAuth token
team_id = "T01234567"

[integrations.slack.status]
focus = { emoji = ":pomodoro:", text = "Focusing: {intention}" }
break = { emoji = ":coffee:", text = "On a break" }
clear_on_end = true

[integrations.slack.dnd]
enabled = true
focus = true     # Enable DnD during focus
break = false    # Don't DnD during break
```

Uses Slack Web API:
- `users.profile.set` — update status emoji + text
- `dnd.setSnooze` / `dnd.endSnooze` — toggle Do Not Disturb

### 8.4 Calendar Integration (v2.0)

Two modes:

**Export mode (simple):** Write `.ics` files that can be imported into any calendar:
```
sesh export --format ics --output ~/sessions.ics
```

**Google Calendar API (advanced):**
```toml
[integrations.calendar]
enabled = true
provider = "google"                   # or "ics-file"
calendar_id = "primary"
credentials_path = "~/.config/sesh/google-credentials.json"
create_events = true                  # Auto-create calendar events for sessions
color_by_category = true              # Map category colors to Google Calendar colors
```

---

## 9. CLI Interface

sesh operates in two modes: **interactive** (TUI) and **non-interactive** (CLI subcommands).

### 9.1 Command Reference

```
sesh [OPTIONS] [COMMAND]

COMMANDS:
  (none)              Launch interactive TUI
  start               Start a focus session
  break               Start a break
  stop                Stop current session
  pause               Pause/resume current session
  abandon             Abandon current session
  status              Show current timer status
  history             List past sessions
  analytics           Show analytics summary
  categories          Manage categories
  export              Export session data
  tmux-status         Output tmux status line segment
  config              Show/edit configuration
  version             Show version info

OPTIONS:
  --config <PATH>     Config file path (default: ~/.config/sesh/config.toml)
  --data-dir <PATH>   Data directory (default: ~/.local/share/sesh/)
  --no-tui            Force non-interactive mode
  --pane-mode         Launch in minimal pane mode
  -v, --verbose       Increase log verbosity
  -q, --quiet         Suppress non-essential output
```

### 9.2 Subcommand Details

```bash
# Start a session with intention and category
$ sesh start --intention "Coding sesh" --category Development --duration 30

# Start with defaults (uses last intention or prompts)
$ sesh start

# Start previous session (same intention + category)
$ sesh start --previous

# Check status (great for tmux/polybar)
$ sesh status
{"state":"focus","remaining":"18:42","intention":"Coding sesh","category":"Development","elapsed":"6:18","target":"25:00"}

# Human-readable status
$ sesh status --format human
⏱ 18:42 remaining (FOCUS)
▸ Coding sesh [Development]
Started: 14:32 │ Elapsed: 6:18 │ Target: 25:00

# Status for scripts (exit code based on state)
$ sesh status --quiet
# Exit 0 = running, 1 = idle, 2 = paused, 3 = break

# Analytics
$ sesh analytics --period today
Today: 2h 35m focused │ 6 sessions
  Development  1h 30m  ████████████░░░░  58%
  Writing         45m  ██████░░░░░░░░░░  29%
  Research        20m  ███░░░░░░░░░░░░░  13%

$ sesh analytics --period week
$ sesh analytics --period month
$ sesh analytics --period year

# History
$ sesh history --limit 10
$ sesh history --since 2026-03-01
$ sesh history --category Development --format json

# Categories
$ sesh categories list
$ sesh categories add "Design" --color "#C678DD"
$ sesh categories edit "Design" --color "#E06C75"
$ sesh categories archive "Design"

# Export
$ sesh export --format csv --output ~/sessions.csv
$ sesh export --format json --since 2026-01-01
$ sesh export --format ics  # iCalendar format

# Config
$ sesh config show
$ sesh config edit           # Opens $EDITOR
$ sesh config path           # Print config file path
$ sesh config reset          # Reset to defaults
```

### 9.3 Shell Completions

```bash
# Generate completions
$ sesh completions bash > /etc/bash_completion.d/sesh
$ sesh completions zsh > /usr/local/share/zsh/site-functions/_sesh
$ sesh completions fish > ~/.config/fish/completions/sesh.fish
```

### 9.4 IPC for Multi-Instance Coordination

When the TUI is running, CLI commands communicate via a Unix domain socket:

```
~/.local/share/sesh/sesh.sock
```

If no TUI is running, CLI `start`/`stop`/`pause` commands operate headlessly (timer runs in the CLI process, no UI). The `status` command always works by reading the shared state file.

---

## 10. Configuration

### 10.1 Config File Location

Following XDG Base Directory spec:

| Path | Purpose |
|------|---------|
| `~/.config/sesh/config.toml` | User configuration |
| `~/.local/share/sesh/sessions.db` | SQLite database |
| `~/.local/share/sesh/sesh.sock` | IPC socket |
| `/tmp/sesh-status` | tmux status cache |

### 10.2 Full Configuration Reference

```toml
# ~/.config/sesh/config.toml
# sesh Configuration
# All durations are in minutes unless otherwise noted.

[general]
# Theme: "dark", "light", "catppuccin-mocha", "catppuccin-latte",
#         "gruvbox-dark", "gruvbox-light", "nord", "dracula",
#         "solarized-dark", "solarized-light", "tokyonight",
#         "onedark", "kanagawa"
theme = "dark"

# Show breathing exercise before focus sessions
breathing_enabled = true

# Breathing cycle: [inhale_sec, hold_sec, exhale_sec, hold_sec]
breathing_cycle = [4, 4, 4, 4]

# Number of breath cycles before starting
breathing_rounds = 3

# Prompt for reflection notes after each session
reflection_enabled = true

# Auto-suggest past intentions
autosuggest = true

# Maximum suggestions shown
autosuggest_limit = 10

# Mouse support
mouse = true

# Unicode support (disable for very basic terminals)
unicode = true

# Tick rate in milliseconds (lower = smoother, higher = less CPU)
tick_rate_ms = 250

[timer]
# Default focus duration (minutes)
focus_duration = 25

# Short break duration (minutes)
short_break_duration = 5

# Long break duration (minutes)
long_break_duration = 20

# Cumulative focus minutes before long break triggers
long_break_after = 100

# Auto-start break after focus ends (otherwise stays in reflection)
auto_start_break = false

# Auto-start focus after break ends
auto_start_focus = false

# Session presence reminder interval (minutes, 0 = disabled)
presence_interval = 10

# Pre-end warning (minutes before end, 0 = disabled)
pre_end_warning = 2

# Duration adjustment step for +/- keys (minutes)
adjust_step_large = 5
adjust_step_small = 1

[notifications]
# Enable desktop notifications
enabled = true

# Notification backend: "auto", "notify-rust", "osascript", "notify-send", "bell"
backend = "auto"

# Play terminal bell on notifications
bell = true

# Events to notify on
on_focus_end = true
on_break_end = true
on_overflow = true
on_pre_end = true
on_presence = false

[sound]
# Enable sound playback (requires 'rodio' feature)
enabled = false

# Sound files (paths or built-in names)
# Built-in: "bell", "chime", "bowl", "none"
focus_end = "chime"
break_end = "bell"
overflow = "bowl"
pre_end = "bell"
presence = "bell"

# Volume (0.0 - 1.0)
volume = 0.7

[hooks]
# Shell commands to run on events.
# Available variables: $SESSION_INTENTION, $SESSION_CATEGORY,
#   $SESSION_CATEGORY_COLOR, $SESSION_DURATION, $SESSION_ELAPSED,
#   $SESSION_STATE, $SESSION_TYPE

# session_start = "notify-send 'Focus started: $SESSION_INTENTION'"
# session_end = ""
# session_pause = ""
# session_resume = ""
# break_start = ""
# break_end = ""
# overflow = ""
# abandon = ""

# Example: Toggle macOS DnD
# session_start = "shortcuts run 'Enable DnD'"
# session_end = "shortcuts run 'Disable DnD'"

# Example: Block distracting sites via /etc/hosts
# session_start = "sudo sesh-blocker enable"
# session_end = "sudo sesh-blocker disable"

[integrations.tmux]
enabled = true

# Update tmux status line (via sesh tmux-status)
status_enabled = true

# Rename tmux window during sessions
rename_window = false

# Format strings (available: {state}, {remaining}, {elapsed},
#   {intention}, {category}, {category_color}, {target})
status_format_focus = "⏱ {remaining} {intention}"
status_format_overflow = "◆ +{elapsed} {intention}"
status_format_break = "☕ {remaining}"
status_format_paused = "⏸ {remaining}"
status_format_idle = ""

[integrations.slack]
enabled = false
# token = "xoxp-..."
# status_focus_emoji = ":pomodoro:"
# status_focus_text = "Focusing: {intention}"
# status_break_emoji = ":coffee:"
# status_break_text = "On a break"
# dnd_on_focus = true
# clear_on_end = true

[export]
# Default export format: "csv", "json", "ics"
default_format = "csv"

# Include these fields in export
include_notes = true
include_pauses = true

# Keybindings — see [keybindings] section
# Format: "key" or "modifier+key"
# Modifiers: ctrl, alt, shift
# Special keys: enter, esc, space, tab, backspace, delete,
#   up, down, left, right, home, end, pageup, pagedown,
#   f1-f12

[keybindings]
# Global
quit = "ctrl+c"
help = "?"
screen_timer = "1"
screen_analytics = "2"
screen_history = "3"
screen_settings = "4"

# Timer
start_focus = "enter"
start_break = "b"
finish = "f"
abandon = "x"
pause_resume = "space"
intention = "i"
category = "c"
duration_up_large = "+"
duration_down_large = "-"
duration_up_small = ">"
duration_down_small = "<"

# Navigation
up = "k"
down = "j"
left = "h"
right = "l"
select = "enter"
back = "esc"
next_tab = "tab"
prev_tab = "shift+tab"

# Analytics
period_daily = "d"
period_weekly = "w"
period_monthly = "m"
navigate_prev = "["
navigate_next = "]"
export = "e"

# History
delete_session = "d"
search = "/"
filter_category = "c"
```

---

## 11. Keybindings

### 11.1 Full Reference Table

#### Global (available on all screens)

| Key | Action | Notes |
|-----|--------|-------|
| `Ctrl+c` / `q` | Quit | Confirms if timer running |
| `?` / `F1` | Toggle help overlay | Context-sensitive |
| `1` | Switch to Timer screen | |
| `2` | Switch to Analytics screen | |
| `3` | Switch to History screen | |
| `4` | Switch to Settings screen | |
| `Tab` | Next screen | Cycles 1→2→3→4→1 |
| `Shift+Tab` | Previous screen | Cycles 4→3→2→1→4 |

#### Timer Screen — Idle

| Key | Action | Notes |
|-----|--------|-------|
| `Enter` | Start focus session | Opens intention input if empty |
| `b` | Start break | Short break; `B` for long break |
| `i` | Edit intention | Opens input overlay |
| `c` | Select category | Opens category picker |
| `+` / `=` | Increase duration | By `adjust_step_large` (5m) |
| `-` | Decrease duration | By `adjust_step_large` (5m) |
| `>` / `.` | Increase duration | By `adjust_step_small` (1m) |
| `<` / `,` | Decrease duration | By `adjust_step_small` (1m) |
| `p` | Select profile | Opens profile picker |

#### Timer Screen — Focus / Overflow

| Key | Action | Notes |
|-----|--------|-------|
| `Space` | Pause / Resume | Toggle |
| `f` | Finish session | Saves + opens reflection |
| `b` | Finish + break | Saves + starts break |
| `x` | Abandon | 5s undo window |
| `i` | Edit intention | While running |

#### Timer Screen — Break

| Key | Action | Notes |
|-----|--------|-------|
| `Space` | Pause / Resume | |
| `Enter` | End break + start focus | New session |
| `f` | End break | Return to idle |
| `x` | Abandon break | |

#### Timer Screen — Abandoned (5s undo window)

| Key | Action | Notes |
|-----|--------|-------|
| `u` / `Ctrl+z` | Undo abandon | Restores previous state |
| (any other) | Confirm abandon | Or wait 5s |

#### Intention Input Overlay

| Key | Action | Notes |
|-----|--------|-------|
| (typing) | Input text | Free-form |
| `Tab` | Jump to category | |
| `Up` / `Down` | Navigate suggestions | |
| `Enter` | Confirm + start | |
| `Esc` | Cancel | Returns to previous state |
| `Ctrl+u` | Clear input | |
| `Ctrl+w` | Delete word | |

#### Category Picker

| Key | Action | Notes |
|-----|--------|-------|
| `j` / `Down` | Move down | |
| `k` / `Up` | Move up | |
| `Enter` | Select | |
| `Esc` | Cancel | |
| `n` | New category | |
| `e` | Edit selected | |
| `d` | Delete selected | Confirms first |

#### Analytics Screen

| Key | Action | Notes |
|-----|--------|-------|
| `d` | Daily view | |
| `w` | Weekly view | |
| `m` | Monthly view | |
| `[` / `h` | Previous period | |
| `]` / `l` | Next period | |
| `t` | Jump to today | |
| `e` | Export visible data | |
| `c` | Filter by category | |

#### History Screen

| Key | Action | Notes |
|-----|--------|-------|
| `j` / `Down` | Next session | |
| `k` / `Up` | Previous session | |
| `Enter` | Toggle detail view | |
| `/` | Search intentions | |
| `c` | Filter by category | |
| `d` | Delete session | Confirms first |
| `e` | Export filtered data | |
| `n` | Edit notes | On selected session |

#### Settings Screen

| Key | Action | Notes |
|-----|--------|-------|
| `Tab` | Switch panes (sidebar/main) | |
| `j` / `Down` | Next setting | |
| `k` / `Up` | Previous setting | |
| `Enter` | Edit value | Opens inline editor |
| `Esc` | Cancel edit | |
| `s` | Save changes | |
| `r` | Reset to defaults | Confirms first |

### 11.2 Vim vs Default Mode

The default keybindings use vim-style navigation (`hjkl`). Arrow keys always work as aliases. A future config option could provide alternative schemes:

```toml
[keybindings]
preset = "vim"     # "vim" (default), "arrows", "emacs"
```

---

## 12. Color Themes

### 12.1 Color Tier Support

sesh auto-detects terminal color capabilities and degrades gracefully:

| Tier | Detection | Rendering |
|------|-----------|-----------|
| **Truecolor (24-bit)** | `$COLORTERM == "truecolor"` | Full RGB category colors, smooth gradients |
| **256-color** | `$TERM` contains `256color` | Nearest 256-color palette match |
| **16-color** | Default | Maps to basic ANSI colors |
| **No color** | `$NO_COLOR` set or `--no-color` flag | Monochrome with bold/dim/underline |

### 12.2 Theme Structure

```rust
struct Theme {
    name: String,

    // Base colors
    bg: Color,              // Main background
    fg: Color,              // Main foreground
    bg_secondary: Color,    // Panel/card backgrounds
    fg_secondary: Color,    // Muted text
    border: Color,          // Box borders
    border_focus: Color,    // Focused element borders

    // Semantic colors
    accent: Color,          // Primary action (start button, selections)
    success: Color,         // Completed sessions, positive stats
    warning: Color,         // Overflow, approaching end
    error: Color,           // Abandon, errors
    info: Color,            // Breaks, informational

    // State-specific
    focus_bg: Color,        // Background tint during focus
    overflow_bg: Color,     // Background tint during overflow (uses category color)
    break_bg: Color,        // Background tint during break
    paused_fg: Color,       // Dimmed text during pause

    // UI elements
    progress_filled: Color,
    progress_empty: Color,
    clock_ring: Color,
    clock_hand: Color,
    status_bar_bg: Color,
    status_bar_fg: Color,

    // Chart colors (for analytics, 8 colors cycle)
    chart_palette: [Color; 8],
}
```

### 12.3 Built-in Themes

| Theme | Background | Accent | Style |
|-------|-----------|--------|-------|
| `dark` (default) | `#1E1E2E` | `#A3E635` | Session-inspired, dark with green accents |
| `light` | `#FAFAFA` | `#16A34A` | Clean light theme |
| `catppuccin-mocha` | `#1E1E2E` | `#CBA6F7` | Popular pastel dark |
| `catppuccin-latte` | `#EFF1F5` | `#8839EF` | Popular pastel light |
| `gruvbox-dark` | `#282828` | `#FABD2F` | Warm retro dark |
| `gruvbox-light` | `#FBF1C7` | `#D65D0E` | Warm retro light |
| `nord` | `#2E3440` | `#88C0D0` | Cool arctic |
| `dracula` | `#282A36` | `#BD93F9` | Dark purple |
| `solarized-dark` | `#002B36` | `#268BD2` | Ethan Schoonover dark |
| `solarized-light` | `#FDF6E3` | `#268BD2` | Ethan Schoonover light |
| `tokyonight` | `#1A1B26` | `#7AA2F7` | Modern dark blue |
| `onedark` | `#282C34` | `#61AFEF` | Atom-inspired |
| `kanagawa` | `#1F1F28` | `#DCA561` | Japanese ink |

### 12.4 Custom Theme Definition

```toml
# In config.toml:
[theme.custom]
bg = "#1a1b26"
fg = "#c0caf5"
accent = "#7aa2f7"
# ... (all fields from Theme struct)

[general]
theme = "custom"
```

### 12.5 Category Color Mapping

Categories use user-defined hex colors. In 16-color terminals, these map to nearest ANSI:

```
#E06C75 → Red        #61AFEF → Blue       #98C379 → Green
#C678DD → Magenta    #E5C07B → Yellow     #56B6C2 → Cyan
#D19A66 → Yellow     #ABB2BF → White      #5C6370 → DarkGray
```

---

## 13. Sound & Notifications

### 13.1 Sound Architecture

```
┌──────────────────────────────────────────────────┐
│                Sound Dispatcher                   │
│                                                   │
│  Event ──▶ ┌─────────┐  ┌──────────────────────┐ │
│            │ Enabled? │──│ Backend Selection     │ │
│            └─────────┘  │                      │ │
│                          │  ┌─────────────────┐ │ │
│                          │  │ rodio (bundled)  │ │ │
│                          │  │ afplay (macOS)   │ │ │
│                          │  │ paplay (Linux)   │ │ │
│                          │  │ aplay  (Linux)   │ │ │
│                          │  │ bell (\x07)      │ │ │
│                          │  └─────────────────┘ │ │
│                          └──────────────────────┘ │
└──────────────────────────────────────────────────┘
```

### 13.2 Sound Backends (Priority Order)

| Backend | Platform | Method | Notes |
|---------|----------|--------|-------|
| `rodio` | All | Rust crate, bundled | Best quality, optional compile feature |
| `afplay` | macOS | Shell: `afplay <file>` | Built into macOS, async |
| `paplay` | Linux | Shell: `paplay <file>` | PulseAudio, most common |
| `aplay` | Linux | Shell: `aplay <file>` | ALSA fallback |
| `mpv` | All | Shell: `mpv --no-video <file>` | Fallback if installed |
| Terminal bell | All | Write `\x07` to stdout | Always works, least pleasant |

### 13.3 Built-in Sounds

Bundled as small WAV/OGG files in the binary (via `include_bytes!`):

| Sound | Duration | Use |
|-------|----------|-----|
| `bell` | ~0.5s | Pre-end warning, presence ping |
| `chime` | ~1.5s | Focus complete |
| `bowl` | ~2.0s | Singing bowl — overflow start, break end |
| `click` | ~0.1s | Timer start |

Total bundled audio: ~200KB (compressed OGG).

### 13.4 Notification Priority Matrix

| Event | Desktop Notif | Sound | Bell | tmux |
|-------|:---:|:---:|:---:|:---:|
| Focus end | Yes | `chime` | Yes | Update |
| Break end | Yes | `bowl` | Yes | Update |
| Overflow start | Yes | `bowl` | No | Update |
| Pre-end warning | Optional | `bell` | No | — |
| Session presence | Optional | `bell` | Optional | — |
| Abandon | No | No | No | Clear |
| Pause/Resume | No | No | No | Update |

---

## 14. Distribution

### 14.1 Installation Methods

| Method | Command | Platform |
|--------|---------|----------|
| **Cargo** | `cargo install sesh` | All (requires Rust toolchain) |
| **Homebrew** | `brew install sesh` | macOS, Linux |
| **AUR** | `paru -S sesh` | Arch Linux |
| **Nix** | `nix profile install nixpkgs#sesh` | NixOS, any with Nix |
| **GitHub Releases** | Download binary from releases page | All |
| **Conda-forge** | `conda install sesh` | All (future) |
| **Docker** | `docker run -it sesh` | All (for trying it out) |

### 14.2 Build Matrix

| Target | OS | Arch | Notes |
|--------|----|------|-------|
| `x86_64-unknown-linux-gnu` | Linux | x86_64 | Primary |
| `x86_64-unknown-linux-musl` | Linux | x86_64 | Static binary, Alpine |
| `aarch64-unknown-linux-gnu` | Linux | ARM64 | Raspberry Pi, ARM servers |
| `x86_64-apple-darwin` | macOS | x86_64 | Intel Mac |
| `aarch64-apple-darwin` | macOS | ARM64 | Apple Silicon |
| `x86_64-pc-windows-msvc` | Windows | x86_64 | Windows Terminal |

### 14.3 Feature Flags

```toml
[features]
default = ["sound", "notifications"]
sound = ["rodio"]              # Audio playback via rodio
notifications = ["notify-rust"] # Desktop notifications
slack = ["reqwest"]            # Slack integration
minimal = []                   # No optional deps, smallest binary
```

Build variants:
```bash
# Full-featured (default)
cargo install sesh

# Minimal (no sound, no notifications — good for SSH/headless)
cargo install sesh --no-default-features

# With Slack integration
cargo install sesh --features slack
```

### 14.4 Binary Size Estimates

| Variant | Estimated Size | Notes |
|---------|---------------|-------|
| Full (with sounds) | ~6 MB | Includes bundled audio |
| Full (no sounds) | ~4 MB | External sound files |
| Minimal | ~2.5 MB | Timer + DB + TUI only |

### 14.5 Homebrew Formula (Draft)

```ruby
class SessionTui < Formula
  desc "Terminal-based Pomodoro focus timer with analytics"
  homepage "https://github.com/sesh/sesh"
  url "https://github.com/sesh/sesh/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "rust" => :build

  def install
    system "cargo", "install", *std_cargo_args
    generate_completions_from_executable(bin/"sesh", "completions")
  end

  test do
    assert_match "sesh #{version}", shell_output("#{bin}/sesh version")
  end
end
```

### 14.6 Release Checklist

1. Update `Cargo.toml` version
2. Update CHANGELOG.md
3. `cargo test --all-features`
4. `cargo clippy --all-features -- -D warnings`
5. Tag release: `git tag v0.1.0`
6. GitHub Actions builds all targets
7. Publish to crates.io: `cargo publish`
8. Update Homebrew formula
9. Update AUR PKGBUILD
10. GitHub Release with binaries + changelog

---

## Appendix A: Comparison with Original Session App

| Feature | Session (Original) | sesh | Notes |
|---------|-------------------|-------------|-------|
| Timer | Analog clock (rotary) | ASCII clock + progress bar | Adapted for terminal |
| Overflow | Background color change | Full-screen color tint + symbol | Uses category color |
| Breathing | Animated circle | ASCII expanding/contracting | Simplified |
| Intentions | Text field | Text input with fuzzy suggest | Same concept |
| Categories | Color-coded, unlimited | Color-coded, unlimited | Identical |
| Profiles | Per-category settings | Per-category settings | Identical |
| Analytics | Bar charts, calendar | ASCII charts, heatmap | Adapted for terminal |
| Website blocker | Safari/Chrome/Brave/Edge | Shell hooks (user-provided) | Different approach |
| App blocker | Screen Time API | Shell hooks (user-provided) | Different approach |
| Slack integration | OAuth + API | OAuth token in config | Simplified |
| Calendar | Apple Calendar bidirectional | ICS export + Google API (future) | Different approach |
| Sync | Custom server, real-time | None (local SQLite) | By design |
| Widgets | iOS/macOS widgets | tmux status line | Adapted |
| Live Activities | Dynamic Island | tmux/polybar/waybar | Adapted |
| Siri/Shortcuts | Apple ecosystem | CLI + shell hooks | Unix equivalent |
| Watch app | watchOS native | — | Not applicable |
| Chrome Extension | Todoist/Trello/GitHub | — | Not applicable |
| Dark mode | Yes | Yes (+ 13 themes) | Expanded |
| Pricing | $4.99/mo subscription | Free, open source | Different model |
| Platforms | Apple only | Linux, macOS, Windows | Expanded |

---

## Appendix B: Minimum Terminal Requirements

| Feature | Minimum | Recommended |
|---------|---------|-------------|
| Size (full) | 60x24 | 80x30+ |
| Size (pane) | 40x10 | 50x15+ |
| Size (minimal) | 20x3 | 30x5+ |
| Colors | 16 | Truecolor |
| Unicode | Not required (ASCII fallback) | Recommended |
| Mouse | Not required | Supported |

---

## Appendix C: Prior Art & Inspiration

- **[Porsmo](https://github.com/ColorCookie-dev/porsmo)** — Rust TUI Pomodoro (simple, no analytics)
- **[Pomogoro](https://github.com/Kleysley/pomogoro)** — Go TUI Pomodoro
- **[openpomodoro-cli](https://github.com/open-pomodoro/openpomodoro-cli)** — CLI Pomodoro standard
- **[Watson](https://github.com/TailorDev/Watson)** — CLI time tracker (no Pomodoro)
- **[Taskwarrior](https://taskwarrior.org/)** — CLI task management (integration candidate)
- **[bottom](https://github.com/ClementTsang/bottom)** — Rust TUI system monitor (UI reference for charts)
- **[lazygit](https://github.com/jesseduffield/lazygit)** — Go TUI git client (UI reference for overlays)

---

*Document generated: 2026-03-17*
*Status: Draft — Ready for implementation review*
