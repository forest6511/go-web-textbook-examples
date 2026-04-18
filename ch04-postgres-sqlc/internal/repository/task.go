package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "github.com/forest6511/go-web-textbook-examples/ch04-postgres-sqlc/internal/db/gen"
	"github.com/forest6511/go-web-textbook-examples/ch04-postgres-sqlc/internal/domain"
)

type PostgresTaskRepo struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

func NewPostgresTaskRepo(pool *pgxpool.Pool) *PostgresTaskRepo {
	return &PostgresTaskRepo{
		pool:    pool,
		queries: dbgen.New(pool),
	}
}

func (r *PostgresTaskRepo) Create(
	ctx context.Context, t domain.Task,
) (domain.Task, error) {
	row, err := r.queries.CreateTask(ctx, dbgen.CreateTaskParams{
		UserID: t.UserID,
		Title:  t.Title,
		Status: string(t.Status),
	})
	if err != nil {
		return domain.Task{}, mapPgError(err)
	}
	return toDomain(row), nil
}

func (r *PostgresTaskRepo) GetByID(
	ctx context.Context, userID, id int64,
) (domain.Task, error) {
	row, err := r.queries.GetTask(ctx, dbgen.GetTaskParams{
		ID: id, UserID: userID,
	})
	if err != nil {
		return domain.Task{}, mapPgError(err)
	}
	return toDomain(row), nil
}

func toDomain(t dbgen.Task) domain.Task {
	return domain.Task{
		ID:        t.ID,
		UserID:    t.UserID,
		Title:     t.Title,
		Status:    domain.Status(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// ToDomain は toDomain のエクスポート版。usecase など他パッケージから呼ぶ
func ToDomain(t dbgen.Task) domain.Task {
	return toDomain(t)
}

func (r *PostgresTaskRepo) ListByUser(
	ctx context.Context, userID int64, limit, offset int32,
) ([]domain.Task, error) {
	rows, err := r.queries.ListTasksByUser(ctx,
		dbgen.ListTasksByUserParams{
			UserID: userID, Limit: limit, Offset: offset,
		})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	out := make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *PostgresTaskRepo) UpdateStatus(
	ctx context.Context, userID, id int64, s domain.Status,
) error {
	rows, err := r.queries.UpdateTaskStatus(ctx,
		dbgen.UpdateTaskStatusParams{
			Status: string(s), ID: id, UserID: userID,
		})
	if err != nil {
		return mapPgError(err)
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *PostgresTaskRepo) Delete(
	ctx context.Context, userID, id int64,
) error {
	if err := r.queries.DeleteTask(ctx,
		dbgen.DeleteTaskParams{ID: id, UserID: userID},
	); err != nil {
		return mapPgError(err)
	}
	return nil
}
