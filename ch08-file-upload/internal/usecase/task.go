package usecase

import (
	"context"

	dbgen "github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/db/gen"
	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/repository"
)

type TaskRepository interface {
	Create(ctx context.Context, t domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, userID, id int64) (domain.Task, error)
	ListByUser(
		ctx context.Context, userID int64, limit, offset int32,
	) ([]domain.Task, error)
	UpdateStatus(
		ctx context.Context, userID, id int64, s domain.Status,
	) error
	Delete(ctx context.Context, userID, id int64) error
}

type TaskUsecase struct {
	repo TaskRepository
	tx   *repository.TxRunner
}

func New(
	repo TaskRepository, tx *repository.TxRunner,
) *TaskUsecase {
	return &TaskUsecase{repo: repo, tx: tx}
}

func (u *TaskUsecase) Create(
	ctx context.Context, userID int64, title string,
) (domain.Task, error) {
	return u.repo.Create(ctx, domain.Task{
		UserID: userID,
		Title:  title,
		Status: domain.StatusOpen,
	})
}

func (u *TaskUsecase) CreateWithAudit(
	ctx context.Context, userID int64, title string,
) (domain.Task, error) {
	var created domain.Task
	err := u.tx.Run(ctx, func(
		ctx context.Context, q *dbgen.Queries,
	) error {
		row, err := q.CreateTask(ctx, dbgen.CreateTaskParams{
			UserID: userID,
			Title:  title,
			Status: string(domain.StatusOpen),
		})
		if err != nil {
			return err
		}
		created = repository.ToDomain(row)
		return q.InsertAudit(ctx, dbgen.InsertAuditParams{
			Action: "task_create",
			TaskID: row.ID,
		})
	})
	if err != nil {
		return domain.Task{}, err
	}
	return created, nil
}

func (u *TaskUsecase) Get(
	ctx context.Context, userID, id int64,
) (domain.Task, error) {
	return u.repo.GetByID(ctx, userID, id)
}

func (u *TaskUsecase) List(
	ctx context.Context, userID int64, limit, offset int32,
) ([]domain.Task, error) {
	return u.repo.ListByUser(ctx, userID, limit, offset)
}

func (u *TaskUsecase) UpdateStatus(
	ctx context.Context, userID, id int64, status domain.Status,
) error {
	return u.repo.UpdateStatus(ctx, userID, id, status)
}

func (u *TaskUsecase) Delete(
	ctx context.Context, userID, id int64,
) error {
	return u.repo.Delete(ctx, userID, id)
}
