package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	General       GeneralConfig       `toml:"general"`
	Timer         TimerConfig         `toml:"timer"`
	Todoist       TodoistConfig       `toml:"todoist"`
	Calendar      CalendarConfig      `toml:"calendar"`
	Notifications NotificationsConfig `toml:"notifications"`
}

type GeneralConfig struct {
	Theme      string `toml:"theme"`
	Mouse      bool   `toml:"mouse"`
	Unicode    bool   `toml:"unicode"`
	TickRateMs int    `toml:"tick_rate_ms"`
}

type TimerConfig struct {
	FocusDuration      int  `toml:"focus_duration"`
	ShortBreakDuration int  `toml:"short_break_duration"`
	LongBreakDuration  int  `toml:"long_break_duration"`
	LongBreakAfter     int  `toml:"long_break_after"`
	AutoStartBreak     bool `toml:"auto_start_break"`
	AutoStartFocus     bool `toml:"auto_start_focus"`
	MinSaveDuration    int  `toml:"min_save_duration"` // seconds; 0 = always save
}

type TodoistConfig struct {
	APIToken          string `toml:"api_token"`
	CommentOnComplete bool   `toml:"comment_on_complete"`
}

type NotificationsConfig struct {
	Enabled bool   `toml:"enabled"`
	Sound   string `toml:"sound"`
}

type CalendarConfig struct {
	Enabled    bool   `toml:"enabled"`
	ICSPath    string `toml:"ics_path"`
	AutoExport bool   `toml:"auto_export"`
	Google     GoogleCalendarConfig
	Outlook    OutlookCalendarConfig
}

type GoogleCalendarConfig struct {
	Enabled      bool   `toml:"enabled"`
	CalendarID   string `toml:"calendar_id"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

type OutlookCalendarConfig struct {
	Enabled    bool   `toml:"enabled"`
	CalendarID string `toml:"calendar_id"`
	ClientID   string `toml:"client_id"`
	TenantID   string `toml:"tenant_id"`
}

func Default() Config {
	return Config{
		General: GeneralConfig{
			Theme:      "dark",
			Mouse:      true,
			Unicode:    true,
			TickRateMs: 250,
		},
		Timer: TimerConfig{
			FocusDuration:      25,
			ShortBreakDuration: 5,
			LongBreakDuration:  20,
			LongBreakAfter:     100,
			MinSaveDuration:    1500,
		},
		Todoist: TodoistConfig{
			CommentOnComplete: true,
		},
		Notifications: NotificationsConfig{
			Enabled: true,
			Sound:   "Glass",
		},
		Calendar: CalendarConfig{
			Enabled:    false,
			ICSPath:    filepath.Join(DataDir(), "sesh.ics"),
			AutoExport: true,
			Google: GoogleCalendarConfig{
				CalendarID: "primary",
			},
			Outlook: OutlookCalendarConfig{
				TenantID: "common",
			},
		},
	}
}

func Load() Config {
	cfg := Default()
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	// Simple TOML parsing using our own parser
	parseTOML(data, &cfg)
	return cfg
}

func ConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "sesh", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sesh", "config.toml")
}

func DataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "sesh")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "sesh")
}

// Save writes the config to the TOML file at ConfigPath().
func Save(cfg Config) error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	var b strings.Builder

	b.WriteString("[general]\n")
	writeTOMLString(&b, "theme", cfg.General.Theme)
	writeTOMLBool(&b, "mouse", cfg.General.Mouse)
	writeTOMLBool(&b, "unicode", cfg.General.Unicode)
	writeTOMLInt(&b, "tick_rate_ms", cfg.General.TickRateMs)

	b.WriteString("\n[timer]\n")
	writeTOMLInt(&b, "focus_duration", cfg.Timer.FocusDuration)
	writeTOMLInt(&b, "short_break_duration", cfg.Timer.ShortBreakDuration)
	writeTOMLInt(&b, "long_break_duration", cfg.Timer.LongBreakDuration)
	writeTOMLInt(&b, "long_break_after", cfg.Timer.LongBreakAfter)
	writeTOMLBool(&b, "auto_start_break", cfg.Timer.AutoStartBreak)
	writeTOMLBool(&b, "auto_start_focus", cfg.Timer.AutoStartFocus)
	writeTOMLInt(&b, "min_save_duration", cfg.Timer.MinSaveDuration)

	b.WriteString("\n[notifications]\n")
	writeTOMLBool(&b, "enabled", cfg.Notifications.Enabled)
	writeTOMLString(&b, "sound", cfg.Notifications.Sound)

	b.WriteString("\n[todoist]\n")
	writeTOMLString(&b, "api_token", cfg.Todoist.APIToken)
	writeTOMLBool(&b, "comment_on_complete", cfg.Todoist.CommentOnComplete)

	b.WriteString("\n[calendar]\n")
	writeTOMLBool(&b, "enabled", cfg.Calendar.Enabled)
	writeTOMLString(&b, "ics_path", cfg.Calendar.ICSPath)
	writeTOMLBool(&b, "auto_export", cfg.Calendar.AutoExport)

	b.WriteString("\n[calendar.google]\n")
	writeTOMLBool(&b, "enabled", cfg.Calendar.Google.Enabled)
	writeTOMLString(&b, "calendar_id", cfg.Calendar.Google.CalendarID)
	writeTOMLString(&b, "client_id", cfg.Calendar.Google.ClientID)
	writeTOMLString(&b, "client_secret", cfg.Calendar.Google.ClientSecret)

	b.WriteString("\n[calendar.outlook]\n")
	writeTOMLBool(&b, "enabled", cfg.Calendar.Outlook.Enabled)
	writeTOMLString(&b, "calendar_id", cfg.Calendar.Outlook.CalendarID)
	writeTOMLString(&b, "client_id", cfg.Calendar.Outlook.ClientID)
	writeTOMLString(&b, "tenant_id", cfg.Calendar.Outlook.TenantID)

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeTOMLString(b *strings.Builder, key, val string) {
	fmt.Fprintf(b, "%s = %q\n", key, val)
}

func writeTOMLBool(b *strings.Builder, key string, val bool) {
	fmt.Fprintf(b, "%s = %s\n", key, strconv.FormatBool(val))
}

func writeTOMLInt(b *strings.Builder, key string, val int) {
	fmt.Fprintf(b, "%s = %d\n", key, val)
}
