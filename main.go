package main

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ivanlee1999/sesh/internal/app"
	"github.com/ivanlee1999/sesh/internal/calendar"
	"github.com/ivanlee1999/sesh/internal/config"
	"github.com/ivanlee1999/sesh/internal/db"
	"github.com/ivanlee1999/sesh/internal/todoist"
	"github.com/ivanlee1999/sesh/internal/ui"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "sesh",
		Short: "A terminal-native Pomodoro focus timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI()
		},
		SilenceUsage: true,
	}

	root.AddCommand(statusCmd())
	root.AddCommand(historyCmd())
	root.AddCommand(analyticsCmd())
	root.AddCommand(todoistCmd())
	root.AddCommand(exportCmd())
	root.AddCommand(startCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI() error {
	cfg := config.Load()
	database, err := db.OpenDefault()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.DB.Close()
	database.InsertDefaultCategories()

	m := app.NewModel(database, cfg)
	p := tea.NewProgram(viewModel{m}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

// viewModel wraps app.Model to implement View via the ui package
type viewModel struct {
	app.Model
}

func (v viewModel) View() string {
	return ui.View(v.Model)
}

func (v viewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := v.Model.Update(msg)
	return viewModel{m}, cmd
}

// CLI commands

func statusCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current timer status",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()

			focusMins, sessions, _ := database.GetTodayStats()
			if format == "json" {
				fmt.Printf(`{"state":"idle","today_focus_minutes":%.1f,"today_sessions":%d}`+"\n", focusMins, sessions)
			} else {
				h := int(focusMins) / 60
				min := int(focusMins) % 60
				fmt.Println("Status: IDLE")
				fmt.Printf("Today: %dh %dm focused │ %d sessions\n", h, min, sessions)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, human")
	return cmd
}

func historyCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List past sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()

			sessions, err := database.GetSessions(limit)
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions recorded yet.")
				return nil
			}

			fmt.Printf("%-20s %-30s %-15s %8s\n", "Time", "Intention", "Category", "Duration")
			fmt.Println(repeat("─", 75))
			for _, s := range sessions {
				start := extractTime(s.StartedAt)
				end := extractTime(s.EndedAt)
				cat := "—"
				if s.CategoryTitle != nil {
					cat = *s.CategoryTitle
				}
				title := s.Title
				if title == "" {
					title = "(none)"
				}
				dur := fmt.Sprintf("%d:%02d", s.ActualSeconds/60, s.ActualSeconds%60)
				fmt.Printf("%-20s %-30s %-15s %8s\n",
					start+" - "+end, truncate(title, 28), cat, dur)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of sessions to show")
	return cmd
}

func analyticsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analytics",
		Short: "Show analytics summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()

			focusMins, sessions, _ := database.GetTodayStats()
			streak := database.GetStreak()
			fmt.Printf("Today: %s focused │ %d sessions │ Streak: %d days\n",
				app.FormatFocusTime(focusMins), sessions, streak)

			breakdown, _ := database.GetCategoryBreakdownToday()
			var total float64
			for _, b := range breakdown {
				total += b.Minutes
			}
			for _, b := range breakdown {
				pct := 0.0
				if total > 0 {
					pct = b.Minutes / total * 100
				}
				barLen := int(pct / 5)
				bar := repeat("█", barLen) + repeat("░", 20-barLen)
				fmt.Printf("  %-16s %s %3.0f%% %s\n", b.Name, bar, pct, app.FormatFocusTime(b.Minutes))
			}
			return nil
		},
	}
}

func todoistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "todoist",
		Short: "List today's Todoist tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			if cfg.Todoist.APIToken == "" {
				fmt.Fprintln(os.Stderr, "Error: No Todoist API token configured.")
				fmt.Fprintf(os.Stderr, "Add your token to %s\n\n", config.ConfigPath())
				fmt.Fprintln(os.Stderr, "  [todoist]")
				fmt.Fprintln(os.Stderr, `  api_token = "your-api-token-here"`)
				fmt.Fprintln(os.Stderr, "\nGet your token at: https://todoist.com/prefs/integrations")
				return nil
			}

			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()

			client := todoist.NewClient(cfg.Todoist.APIToken)
			fmt.Println("Fetching today's Todoist tasks...")
			tasks, err := client.GetTodayTasks()
			if err != nil {
				return err
			}
			projects, _ := client.GetProjects()

			if len(tasks) == 0 {
				fmt.Println("No tasks due today. You're all caught up!")
				return nil
			}

			fmt.Printf("\n%-4s %-50s %-20s\n", "#", "Task", "Project")
			fmt.Println(repeat("─", 74))
			for i, task := range tasks {
				projName := "—"
				for _, p := range projects {
					if p.ID == task.ProjectID {
						projName = p.Name
						break
					}
				}
				fmt.Printf("%-4d %-50s %-20s\n", i+1, truncate(task.Content, 48), projName)
			}
			fmt.Printf("\nStart a session with: sesh start --todoist %s\n", tasks[0].ID)
			return nil
		},
	}
}

func startCmd() *cobra.Command {
	var intention, todoistID string
	var duration int
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a focus session",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()
			database.InsertDefaultCategories()

			finalIntention := intention
			var matchedCategory string

			if todoistID != "" {
				if cfg.Todoist.APIToken == "" {
					return fmt.Errorf("no Todoist API token configured. See `sesh todoist` for setup")
				}
				client := todoist.NewClient(cfg.Todoist.APIToken)
				task, err := client.GetTask(todoistID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not fetch Todoist task %s: %v\n", todoistID, err)
				} else {
					if finalIntention == "" {
						finalIntention = task.Content
					}
					projects, _ := client.GetProjects()
					cats, _ := database.GetCategories()
					if idx := todoist.MatchProjectToCategory(task.ProjectID, projects, cats); idx >= 0 {
						matchedCategory = cats[idx].Title
					}
					fmt.Printf("Linked to Todoist task: %s\n", task.Content)
				}
			}

			fmt.Println("Starting focus session...")
			if finalIntention != "" {
				fmt.Printf("  Intention: %s\n", finalIntention)
			}
			if matchedCategory != "" {
				fmt.Printf("  Category: %s\n", matchedCategory)
			}
			if duration > 0 {
				fmt.Printf("  Duration: %d minutes\n", duration)
			}
			fmt.Println("(Non-interactive mode not yet fully implemented. Use TUI mode instead.)")
			return nil
		},
	}
	cmd.Flags().StringVarP(&intention, "intention", "i", "", "Intention text")
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in minutes")
	cmd.Flags().StringVar(&todoistID, "todoist", "", "Todoist task ID")
	return cmd
}

func exportCmd() *cobra.Command {
	var format, output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export session data",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			database, err := db.OpenDefault()
			if err != nil {
				return err
			}
			defer database.DB.Close()

			sessions, err := database.GetSessions(10000)
			if err != nil {
				return err
			}

			switch format {
			case "ics":
				path := output
				if path == "" {
					path = cfg.Calendar.ICSPath
				}
				if err := calendar.ExportICS(sessions, path); err != nil {
					return err
				}
				fmt.Printf("Exported %d sessions to %s\n", len(sessions), path)
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				type jsonSession struct {
					ID         string  `json:"id"`
					Title      string  `json:"title"`
					Category   *string `json:"category"`
					Type       string  `json:"type"`
					TargetSecs int64   `json:"target_seconds"`
					ActualSecs int64   `json:"actual_seconds"`
					StartedAt  string  `json:"started_at"`
					EndedAt    string  `json:"ended_at"`
				}
				var out []jsonSession
				for _, s := range sessions {
					out = append(out, jsonSession{
						ID: s.ID, Title: s.Title, Category: s.CategoryTitle,
						Type: s.SessionType, TargetSecs: s.TargetSeconds,
						ActualSecs: s.ActualSeconds, StartedAt: s.StartedAt, EndedAt: s.EndedAt,
					})
				}
				enc.Encode(out)
			case "csv":
				fmt.Println("id,title,category,type,target_seconds,actual_seconds,started_at,ended_at")
				for _, s := range sessions {
					cat := ""
					if s.CategoryTitle != nil {
						cat = *s.CategoryTitle
					}
					fmt.Printf("%s,%s,%s,%s,%d,%d,%s,%s\n",
						csvEscape(s.ID), csvEscape(s.Title), csvEscape(cat),
						s.SessionType, s.TargetSeconds, s.ActualSeconds,
						s.StartedAt, s.EndedAt)
				}
			default:
				return fmt.Errorf("unknown format: %s. Use ics, json, or csv", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "ics", "Export format: ics, json, csv")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

// Helpers

func extractTime(dt string) string {
	for i, c := range dt {
		if c == 'T' && len(dt) >= i+6 {
			return dt[i+1 : i+6]
		}
	}
	return "??:??"
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func repeat(s string, n int) string {
	if n < 0 {
		n = 0
	}
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func csvEscape(s string) string {
	if containsAny(s, ",\"\n") {
		return `"` + escapeQuotes(s) + `"`
	}
	return s
}

func containsAny(s, chars string) bool {
	for _, c := range chars {
		for _, sc := range s {
			if c == sc {
				return true
			}
		}
	}
	return false
}

func escapeQuotes(s string) string {
	out := ""
	for _, c := range s {
		if c == '"' {
			out += `""`
		} else {
			out += string(c)
		}
	}
	return out
}
