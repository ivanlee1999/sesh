use ratatui::style::Color;

#[derive(Debug, Clone)]
pub struct Theme {
    pub bg: Color,
    pub fg: Color,
    pub bg_secondary: Color,
    pub fg_secondary: Color,
    pub border: Color,
    pub border_focus: Color,
    pub accent: Color,
    pub success: Color,
    pub warning: Color,
    pub error: Color,
    pub info: Color,
    pub focus_accent: Color,
    pub overflow_accent: Color,
    pub break_accent: Color,
    pub paused_fg: Color,
    pub progress_filled: Color,
    pub progress_empty: Color,
    pub status_bar_bg: Color,
    pub status_bar_fg: Color,
}

impl Theme {
    pub fn dark() -> Self {
        Self {
            bg: Color::Rgb(30, 30, 46),
            fg: Color::Rgb(205, 214, 244),
            bg_secondary: Color::Rgb(49, 50, 68),
            fg_secondary: Color::Rgb(147, 153, 178),
            border: Color::Rgb(88, 91, 112),
            border_focus: Color::Rgb(163, 230, 53),
            accent: Color::Rgb(163, 230, 53),
            success: Color::Rgb(152, 195, 121),
            warning: Color::Rgb(229, 192, 123),
            error: Color::Rgb(224, 108, 117),
            info: Color::Rgb(97, 175, 239),
            focus_accent: Color::Rgb(163, 230, 53),
            overflow_accent: Color::Rgb(229, 192, 123),
            break_accent: Color::Rgb(86, 182, 194),
            paused_fg: Color::Rgb(108, 112, 134),
            progress_filled: Color::Rgb(163, 230, 53),
            progress_empty: Color::Rgb(49, 50, 68),
            status_bar_bg: Color::Rgb(24, 24, 37),
            status_bar_fg: Color::Rgb(147, 153, 178),
        }
    }

    pub fn from_name(name: &str) -> Self {
        match name {
            "dark" | _ => Self::dark(),
        }
    }
}
