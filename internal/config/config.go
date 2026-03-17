package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	General  GeneralConfig  `toml:"general"`
	Timer    TimerConfig    `toml:"timer"`
	Todoist  TodoistConfig  `toml:"todoist"`
	Calendar CalendarConfig `toml:"calendar"`
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
}

type TodoistConfig struct {
	APIToken          string `toml:"api_token"`
	CommentOnComplete bool   `toml:"comment_on_complete"`
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
		},
		Todoist: TodoistConfig{
			CommentOnComplete: true,
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
