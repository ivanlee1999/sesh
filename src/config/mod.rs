pub mod theme;

use serde::{Deserialize, Serialize};
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct Config {
    pub general: GeneralConfig,
    pub timer: TimerConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct GeneralConfig {
    pub theme: String,
    pub mouse: bool,
    pub unicode: bool,
    pub tick_rate_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct TimerConfig {
    pub focus_duration: u64,
    pub short_break_duration: u64,
    pub long_break_duration: u64,
    pub long_break_after: u64,
    pub auto_start_break: bool,
    pub auto_start_focus: bool,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            general: GeneralConfig::default(),
            timer: TimerConfig::default(),
        }
    }
}

impl Default for GeneralConfig {
    fn default() -> Self {
        Self {
            theme: "dark".into(),
            mouse: true,
            unicode: true,
            tick_rate_ms: 250,
        }
    }
}

impl Default for TimerConfig {
    fn default() -> Self {
        Self {
            focus_duration: 25,
            short_break_duration: 5,
            long_break_duration: 20,
            long_break_after: 100,
            auto_start_break: false,
            auto_start_focus: false,
        }
    }
}

impl Config {
    pub fn load() -> Self {
        let path = Self::config_path();
        if path.exists() {
            match std::fs::read_to_string(&path) {
                Ok(contents) => toml::from_str(&contents).unwrap_or_default(),
                Err(_) => Self::default(),
            }
        } else {
            Self::default()
        }
    }

    pub fn config_path() -> PathBuf {
        if let Some(proj_dirs) = directories::ProjectDirs::from("", "", "sesh") {
            proj_dirs.config_dir().join("config.toml")
        } else {
            PathBuf::from("~/.config/sesh/config.toml")
        }
    }

    pub fn data_dir() -> PathBuf {
        if let Some(proj_dirs) = directories::ProjectDirs::from("", "", "sesh") {
            proj_dirs.data_dir().to_path_buf()
        } else {
            PathBuf::from("~/.local/share/sesh")
        }
    }
}
