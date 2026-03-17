use serde::{Deserialize, Serialize};

const TODOIST_API_BASE: &str = "https://api.todoist.com/rest/v2";

#[derive(Debug, Clone, Deserialize)]
pub struct TodoistTask {
    pub id: String,
    pub content: String,
    pub description: String,
    pub project_id: String,
    pub is_completed: bool,
    pub due: Option<TodoistDue>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TodoistDue {
    pub date: String,
    pub string: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TodoistProject {
    pub id: String,
    pub name: String,
    pub color: String,
}

#[derive(Debug, Serialize)]
struct TodoistComment {
    task_id: String,
    content: String,
}

pub struct TodoistClient {
    token: String,
}

impl TodoistClient {
    pub fn new(token: &str) -> Self {
        Self {
            token: token.to_string(),
        }
    }

    pub fn get_today_tasks(&self) -> anyhow::Result<Vec<TodoistTask>> {
        let resp: Vec<TodoistTask> = ureq::get(&format!("{}/tasks", TODOIST_API_BASE))
            .set("Authorization", &format!("Bearer {}", self.token))
            .query("filter", "today | overdue")
            .call()?
            .into_json()?;
        Ok(resp)
    }

    pub fn get_task(&self, task_id: &str) -> anyhow::Result<TodoistTask> {
        let resp: TodoistTask = ureq::get(&format!("{}/tasks/{}", TODOIST_API_BASE, task_id))
            .set("Authorization", &format!("Bearer {}", self.token))
            .call()?
            .into_json()?;
        Ok(resp)
    }

    pub fn get_projects(&self) -> anyhow::Result<Vec<TodoistProject>> {
        let resp: Vec<TodoistProject> = ureq::get(&format!("{}/projects", TODOIST_API_BASE))
            .set("Authorization", &format!("Bearer {}", self.token))
            .call()?
            .into_json()?;
        Ok(resp)
    }

    pub fn add_comment(&self, task_id: &str, content: &str) -> anyhow::Result<()> {
        let body = TodoistComment {
            task_id: task_id.to_string(),
            content: content.to_string(),
        };
        ureq::post(&format!("{}/comments", TODOIST_API_BASE))
            .set("Authorization", &format!("Bearer {}", self.token))
            .send_json(&body)?;
        Ok(())
    }

    /// Match a Todoist project name to a sesh category (case-insensitive substring match)
    pub fn match_project_to_category(
        &self,
        project_id: &str,
        projects: &[TodoistProject],
        categories: &[crate::db::categories::Category],
    ) -> Option<usize> {
        let project = projects.iter().find(|p| p.id == project_id)?;
        let project_name = project.name.to_lowercase();
        categories.iter().position(|c| {
            let cat_name = c.title.to_lowercase();
            project_name.contains(&cat_name) || cat_name.contains(&project_name)
        })
    }
}
