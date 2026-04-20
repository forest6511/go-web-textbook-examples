package domain

import (
	"time"
	"unicode/utf8"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

const MaxTitleLength = 200

type Task struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTask はドメイン制約を満たす Task を生成する。
// 空文字は ErrTitleRequired、MaxTitleLength 超過は ErrTitleTooLong を返す。
func NewTask(userID int64, title string) (Task, error) {
	if title == "" {
		return Task{}, ErrTitleRequired
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return Task{}, ErrTitleTooLong
	}
	return Task{
		UserID: userID,
		Title:  title,
		Status: StatusOpen,
	}, nil
}
