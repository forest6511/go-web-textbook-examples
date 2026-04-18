package domain

import "time"

// Task はタスク管理の中核モデル。HTTP も DB も知らない
type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Priority  int       `json:"priority"` // 0..3
	Status    string    `json:"status"`   // "open" | "done"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
