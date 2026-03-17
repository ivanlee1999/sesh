package config

import (
	"strconv"
	"strings"
)

// Minimal TOML parser for our config — handles flat sections with string, int, bool values.
func parseTOML(data []byte, cfg *Config) {
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")

		switch section {
		case "general":
			switch key {
			case "theme":
				cfg.General.Theme = val
			case "mouse":
				cfg.General.Mouse = val == "true"
			case "unicode":
				cfg.General.Unicode = val == "true"
			case "tick_rate_ms":
				if n, err := strconv.Atoi(val); err == nil {
					cfg.General.TickRateMs = n
				}
			}
		case "timer":
			switch key {
			case "focus_duration":
				if n, err := strconv.Atoi(val); err == nil {
					cfg.Timer.FocusDuration = n
				}
			case "short_break_duration":
				if n, err := strconv.Atoi(val); err == nil {
					cfg.Timer.ShortBreakDuration = n
				}
			case "long_break_duration":
				if n, err := strconv.Atoi(val); err == nil {
					cfg.Timer.LongBreakDuration = n
				}
			case "long_break_after":
				if n, err := strconv.Atoi(val); err == nil {
					cfg.Timer.LongBreakAfter = n
				}
			case "auto_start_break":
				cfg.Timer.AutoStartBreak = val == "true"
			case "auto_start_focus":
				cfg.Timer.AutoStartFocus = val == "true"
			}
		case "todoist":
			switch key {
			case "api_token":
				cfg.Todoist.APIToken = val
			case "comment_on_complete":
				cfg.Todoist.CommentOnComplete = val == "true"
			}
		case "calendar":
			switch key {
			case "enabled":
				cfg.Calendar.Enabled = val == "true"
			case "ics_path":
				cfg.Calendar.ICSPath = val
			case "auto_export":
				cfg.Calendar.AutoExport = val == "true"
			}
		case "calendar.google":
			switch key {
			case "enabled":
				cfg.Calendar.Google.Enabled = val == "true"
			case "calendar_id":
				cfg.Calendar.Google.CalendarID = val
			case "client_id":
				cfg.Calendar.Google.ClientID = val
			case "client_secret":
				cfg.Calendar.Google.ClientSecret = val
			}
		case "notifications":
			switch key {
			case "enabled":
				cfg.Notifications.Enabled = val == "true"
			case "sound":
				cfg.Notifications.Sound = val
			}
		case "calendar.outlook":
			switch key {
			case "enabled":
				cfg.Calendar.Outlook.Enabled = val == "true"
			case "calendar_id":
				cfg.Calendar.Outlook.CalendarID = val
			case "client_id":
				cfg.Calendar.Outlook.ClientID = val
			case "tenant_id":
				cfg.Calendar.Outlook.TenantID = val
			}
		}
	}
}
