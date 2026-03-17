package todoist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ivanlee1999/sesh/internal/db"
)

const apiBase = "https://api.todoist.com/rest/v2"

type Task struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	ProjectID string `json:"project_id"`
}

type Project struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Client struct {
	Token string
}

func NewClient(token string) *Client {
	return &Client{Token: token}
}

func (c *Client) GetTodayTasks() ([]Task, error) {
	req, _ := http.NewRequest("GET", apiBase+"/tasks?filter=today%20%7C%20overdue", nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("todoist API error %d: %s", resp.StatusCode, string(body))
	}
	var tasks []Task
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return tasks, err
}

func (c *Client) GetTask(taskID string) (*Task, error) {
	req, _ := http.NewRequest("GET", apiBase+"/tasks/"+taskID, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("todoist API error %d", resp.StatusCode)
	}
	var task Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	return &task, err
}

func (c *Client) GetProjects() ([]Project, error) {
	req, _ := http.NewRequest("GET", apiBase+"/projects", nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("todoist API error %d", resp.StatusCode)
	}
	var projects []Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	return projects, err
}

func (c *Client) AddComment(taskID, content string) error {
	body := fmt.Sprintf(`{"task_id":"%s","content":"%s"}`, taskID, content)
	req, _ := http.NewRequest("POST", apiBase+"/comments", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func MatchProjectToCategory(projectID string, projects []Project, categories []db.Category) int {
	var projectName string
	for _, p := range projects {
		if p.ID == projectID {
			projectName = strings.ToLower(p.Name)
			break
		}
	}
	if projectName == "" {
		return -1
	}
	for i, cat := range categories {
		catName := strings.ToLower(cat.Title)
		if strings.Contains(projectName, catName) || strings.Contains(catName, projectName) {
			return i
		}
	}
	return -1
}
