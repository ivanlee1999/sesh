# Polish Task for sesh

Read DESIGN.md for full context. The MVP works. Now polish it.

Do ALL of the following, committing and pushing after each step.

## 1. Better Timer Visuals
- Use Braille dot characters or box-drawing to make the timer circle look more refined
- Add color transitions: green for focus, yellow for overflow, blue for break, red for paused
- Show a spinning animation or pulsing effect on the active timer
- git add -A && git commit -m "feat: improved timer visuals with colors and animations" && git push origin main

## 2. Intention Input Flow
- When pressing "i", show a proper text input overlay with cursor
- Support editing, backspace, enter to confirm, escape to cancel
- Show intention prominently on the timer screen during focus
- git add -A && git commit -m "feat: polished intention input overlay" && git push origin main

## 3. Category Picker
- When pressing "c", show a list picker with colored category names
- Arrow keys to navigate, enter to select, escape to cancel
- Each category should show its color inline
- git add -A && git commit -m "feat: category picker with colored list" && git push origin main

## 4. Session Completion Flow
- When timer hits 0, play terminal bell
- Show completion overlay with notes text area
- Auto-save to SQLite with duration, intention, category, notes
- Then prompt: start break or stop?
- git add -A && git commit -m "feat: session completion flow with notes and bell" && git push origin main

## 5. History Table Polish
- Show sessions in a proper table: Date | Duration | Category | Intention | Notes
- Color-code by category, scrollable with j/k or arrow keys
- Show total focus time at the bottom
- git add -A && git commit -m "feat: polished history table with colors" && git push origin main

## 6. Analytics Dashboard
- Bar chart using block characters for daily focus hours (last 7 days)
- Category breakdown with colored bars
- git add -A && git commit -m "feat: analytics with bar charts" && git push origin main

## 7. Overflow Mode Visual
- When timer goes past 0, invert colors or add a pulsing border
- Show +00:XX counting up, different status bar color
- git add -A && git commit -m "feat: overflow mode with inverted visuals" && git push origin main

## 8. Keybinding Help Overlay
- Press ? to show full keybinding reference as an overlay
- Organized by context (global, timer, overlays)
- git add -A && git commit -m "feat: keybinding help overlay" && git push origin main

## 9. First-run Experience
- If no config exists, show welcome screen on first launch
- Let user set default duration and pick a theme, create config automatically
- git add -A && git commit -m "feat: first-run welcome screen" && git push origin main

## 10. README Update
- Feature list, installation (go install), usage examples for TUI and CLI
- ASCII art showing the TUI, configuration docs, Todoist and Calendar setup
- git add -A && git commit -m "docs: comprehensive README with usage examples" && git push origin main

IMPORTANT: Run go build -o sesh . before each commit to make sure it compiles.
