package db

import (
	"github.com/google/uuid"
)

type Category struct {
	ID       string
	Title    string
	HexColor string
	Status   string
	SortOrder int
}

func (d *Database) InsertDefaultCategories() error {
	var count int
	if err := d.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaults := []struct {
		title string
		color string
	}{
		{"Development", "#61AFEF"},
		{"Writing", "#E06C75"},
		{"Design", "#C678DD"},
		{"Research", "#E5C07B"},
		{"Meeting", "#56B6C2"},
		{"Exercise", "#98C379"},
		{"Reading", "#D19A66"},
		{"Admin", "#ABB2BF"},
	}

	for i, d2 := range defaults {
		_, err := d.DB.Exec(
			"INSERT INTO categories (id, title, hex_color, sort_order) VALUES (?, ?, ?, ?)",
			uuid.New().String(), d2.title, d2.color, i,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) GetCategories() ([]Category, error) {
	rows, err := d.DB.Query(
		"SELECT id, title, hex_color, status, sort_order FROM categories WHERE status = 'active' ORDER BY sort_order",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Title, &c.HexColor, &c.Status, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}
