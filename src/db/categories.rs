use super::{Database, DbResult};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct Category {
    pub id: String,
    pub title: String,
    pub hex_color: String,
    pub status: String,
    pub sort_order: i32,
}

impl Database {
    pub fn insert_default_categories(&self) -> DbResult<()> {
        let defaults = vec![
            ("Development", "#61AFEF"),
            ("Writing", "#E06C75"),
            ("Design", "#C678DD"),
            ("Research", "#E5C07B"),
            ("Meeting", "#56B6C2"),
            ("Exercise", "#98C379"),
            ("Reading", "#D19A66"),
            ("Admin", "#ABB2BF"),
        ];

        let count: i64 = self.conn.query_row(
            "SELECT COUNT(*) FROM categories",
            [],
            |row| row.get(0),
        )?;

        if count == 0 {
            for (i, (title, color)) in defaults.iter().enumerate() {
                self.conn.execute(
                    "INSERT INTO categories (id, title, hex_color, sort_order) VALUES (?1, ?2, ?3, ?4)",
                    rusqlite::params![Uuid::new_v4().to_string(), title, color, i as i32],
                )?;
            }
        }

        Ok(())
    }

    pub fn get_categories(&self) -> DbResult<Vec<Category>> {
        let mut stmt = self.conn.prepare(
            "SELECT id, title, hex_color, status, sort_order FROM categories WHERE status = 'active' ORDER BY sort_order"
        )?;
        let cats = stmt.query_map([], |row| {
            Ok(Category {
                id: row.get(0)?,
                title: row.get(1)?,
                hex_color: row.get(2)?,
                status: row.get(3)?,
                sort_order: row.get(4)?,
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;
        Ok(cats)
    }
}
