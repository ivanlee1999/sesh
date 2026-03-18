package app

import (
	"github.com/ivanlee1999/sesh/internal/config"
)

// SettingKind describes the type of a settings item.
type SettingKind int

const (
	SettingHeader SettingKind = iota // non-selectable section header
	SettingBool
	SettingInt
	SettingString
	SettingAction // e.g., "Authenticate Google Calendar"
)

// SettingItem represents one row in the settings list.
type SettingItem struct {
	Label     string
	Key       string // unique identifier like "timer.focus_duration"
	Kind      SettingKind
	Suffix    string // display suffix like " min"
	Sensitive bool   // mask display value

	GetBool   func(*config.Config) bool
	SetBool   func(*config.Config, bool)
	GetInt    func(*config.Config) int
	SetInt    func(*config.Config, int)
	GetString func(*config.Config) string
	SetString func(*config.Config, string)
}

// BuildSettingsItems returns the flat list of all configurable settings.
func BuildSettingsItems() []SettingItem {
	return []SettingItem{
		// ── General ──
		{Label: "General", Kind: SettingHeader},
		{
			Label: "Theme", Key: "general.theme", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.General.Theme },
			SetString: func(c *config.Config, v string) { c.General.Theme = v },
		},
		{
			Label: "Mouse", Key: "general.mouse", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.General.Mouse },
			SetBool: func(c *config.Config, v bool) { c.General.Mouse = v },
		},
		{
			Label: "Unicode", Key: "general.unicode", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.General.Unicode },
			SetBool: func(c *config.Config, v bool) { c.General.Unicode = v },
		},
		{
			Label: "Tick Rate", Key: "general.tick_rate_ms", Kind: SettingInt, Suffix: " ms",
			GetInt: func(c *config.Config) int { return c.General.TickRateMs },
			SetInt: func(c *config.Config, v int) { c.General.TickRateMs = v },
		},

		// ── Timer ──
		{Label: "Timer", Kind: SettingHeader},
		{
			Label: "Focus Duration", Key: "timer.focus_duration", Kind: SettingInt, Suffix: " min",
			GetInt: func(c *config.Config) int { return c.Timer.FocusDuration },
			SetInt: func(c *config.Config, v int) { c.Timer.FocusDuration = v },
		},
		{
			Label: "Short Break", Key: "timer.short_break_duration", Kind: SettingInt, Suffix: " min",
			GetInt: func(c *config.Config) int { return c.Timer.ShortBreakDuration },
			SetInt: func(c *config.Config, v int) { c.Timer.ShortBreakDuration = v },
		},
		{
			Label: "Long Break", Key: "timer.long_break_duration", Kind: SettingInt, Suffix: " min",
			GetInt: func(c *config.Config) int { return c.Timer.LongBreakDuration },
			SetInt: func(c *config.Config, v int) { c.Timer.LongBreakDuration = v },
		},
		{
			Label: "Long Break After", Key: "timer.long_break_after", Kind: SettingInt, Suffix: " min cumulative",
			GetInt: func(c *config.Config) int { return c.Timer.LongBreakAfter },
			SetInt: func(c *config.Config, v int) { c.Timer.LongBreakAfter = v },
		},
		{
			Label: "Auto Start Break", Key: "timer.auto_start_break", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Timer.AutoStartBreak },
			SetBool: func(c *config.Config, v bool) { c.Timer.AutoStartBreak = v },
		},
		{
			Label: "Auto Start Focus", Key: "timer.auto_start_focus", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Timer.AutoStartFocus },
			SetBool: func(c *config.Config, v bool) { c.Timer.AutoStartFocus = v },
		},
		{
			Label: "Min Save Duration", Key: "timer.min_save_duration", Kind: SettingInt, Suffix: " sec",
			GetInt: func(c *config.Config) int { return c.Timer.MinSaveDuration },
			SetInt: func(c *config.Config, v int) { c.Timer.MinSaveDuration = v },
		},

		// ── Notifications ──
		{Label: "Notifications", Kind: SettingHeader},
		{
			Label: "Enabled", Key: "notifications.enabled", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Notifications.Enabled },
			SetBool: func(c *config.Config, v bool) { c.Notifications.Enabled = v },
		},
		{
			Label: "Sound", Key: "notifications.sound", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Notifications.Sound },
			SetString: func(c *config.Config, v string) { c.Notifications.Sound = v },
		},

		// ── Todoist ──
		{Label: "Todoist", Kind: SettingHeader},
		{
			Label: "API Token", Key: "todoist.api_token", Kind: SettingString, Sensitive: true,
			GetString: func(c *config.Config) string { return c.Todoist.APIToken },
			SetString: func(c *config.Config, v string) { c.Todoist.APIToken = v },
		},
		{
			Label: "Comment on Complete", Key: "todoist.comment_on_complete", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Todoist.CommentOnComplete },
			SetBool: func(c *config.Config, v bool) { c.Todoist.CommentOnComplete = v },
		},

		// ── Calendar ──
		{Label: "Calendar", Kind: SettingHeader},
		{
			Label: "Enabled", Key: "calendar.enabled", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Calendar.Enabled },
			SetBool: func(c *config.Config, v bool) { c.Calendar.Enabled = v },
		},
		{
			Label: "ICS Path", Key: "calendar.ics_path", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.ICSPath },
			SetString: func(c *config.Config, v string) { c.Calendar.ICSPath = v },
		},
		{
			Label: "Auto Export", Key: "calendar.auto_export", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Calendar.AutoExport },
			SetBool: func(c *config.Config, v bool) { c.Calendar.AutoExport = v },
		},

		// ── Google Calendar ──
		{Label: "Google Calendar", Kind: SettingHeader},
		{
			Label: "Enabled", Key: "calendar.google.enabled", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Calendar.Google.Enabled },
			SetBool: func(c *config.Config, v bool) { c.Calendar.Google.Enabled = v },
		},
		{
			Label: "Calendar ID", Key: "calendar.google.calendar_id", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.Google.CalendarID },
			SetString: func(c *config.Config, v string) { c.Calendar.Google.CalendarID = v },
		},
		{
			Label: "Client ID", Key: "calendar.google.client_id", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.Google.ClientID },
			SetString: func(c *config.Config, v string) { c.Calendar.Google.ClientID = v },
		},
		{
			Label: "Client Secret", Key: "calendar.google.client_secret", Kind: SettingString, Sensitive: true,
			GetString: func(c *config.Config) string { return c.Calendar.Google.ClientSecret },
			SetString: func(c *config.Config, v string) { c.Calendar.Google.ClientSecret = v },
		},
		{Label: "Authenticate", Key: "calendar.google.auth", Kind: SettingAction},

		// ── Outlook Calendar ──
		{Label: "Outlook Calendar", Kind: SettingHeader},
		{
			Label: "Enabled", Key: "calendar.outlook.enabled", Kind: SettingBool,
			GetBool: func(c *config.Config) bool { return c.Calendar.Outlook.Enabled },
			SetBool: func(c *config.Config, v bool) { c.Calendar.Outlook.Enabled = v },
		},
		{
			Label: "Calendar ID", Key: "calendar.outlook.calendar_id", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.Outlook.CalendarID },
			SetString: func(c *config.Config, v string) { c.Calendar.Outlook.CalendarID = v },
		},
		{
			Label: "Client ID", Key: "calendar.outlook.client_id", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.Outlook.ClientID },
			SetString: func(c *config.Config, v string) { c.Calendar.Outlook.ClientID = v },
		},
		{
			Label: "Tenant ID", Key: "calendar.outlook.tenant_id", Kind: SettingString,
			GetString: func(c *config.Config) string { return c.Calendar.Outlook.TenantID },
			SetString: func(c *config.Config, v string) { c.Calendar.Outlook.TenantID = v },
		},
		{Label: "Authenticate", Key: "calendar.outlook.auth", Kind: SettingAction},
	}
}
