package usecase

import (
	"context"
	"time"

	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/domain"
)

// TaskRepository は永続化層との対話を抽象化する。実装は repository パッケージ
type TaskRepository interface {
	Create(ctx context.Context, t *domain.Task) error
	Get(ctx context.Context, id int64) (*domain.Task, bool)
	List(ctx context.Context, status string, limit int) []*domain.Task
	Update(ctx context.Context, id int64, fn func(*domain.Task)) (*domain.Task, bool)
	Delete(ctx context.Context, id int64) bool
}

type TaskUsecase struct {
	repo TaskRepository
}

func New(repo TaskRepository) *TaskUsecase {
	return &TaskUsecase{repo: repo}
}

func (u *TaskUsecase) Create(ctx context.Context, title, body string, priority int) *domain.Task {
	now := time.Now()
	t := &domain.Task{
		Title:     title,
		Body:      body,
		Priority:  priority,
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = u.repo.Create(ctx, t)
	return t
}

func (u *TaskUsecase) Get(ctx context.Context, id int64) (*domain.Task, bool) {
	return u.repo.Get(ctx, id)
}

func (u *TaskUsecase) List(ctx context.Context, status string, limit int) []*domain.Task {
	return u.repo.List(ctx, status, limit)
}

func (u *TaskUsecase) Update(
	ctx context.Context,
	id int64,
	title, body *string,
	priority *int,
	status *string,
) (*domain.Task, bool) {
	return u.repo.Update(ctx, id, func(t *domain.Task) {
		if title != nil {
			t.Title = *title
		}
		if body != nil {
			t.Body = *body
		}
		if priority != nil {
			t.Priority = *priority
		}
		if status != nil {
			t.Status = *status
		}
		t.UpdatedAt = time.Now()
	})
}

func (u *TaskUsecase) Delete(ctx context.Context, id int64) bool {
	return u.repo.Delete(ctx, id)
}
