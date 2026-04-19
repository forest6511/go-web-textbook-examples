package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "github.com/forest6511/go-web-textbook-examples/ch11-deploy/internal/db/gen"
	"github.com/forest6511/go-web-textbook-examples/ch11-deploy/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
	q    *dbgen.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool, q: dbgen.New(pool)}
}

func (r *UserRepo) Create(
	ctx context.Context, email string, passwordHash []byte, role string,
) (*domain.User, error) {
	row, err := r.q.CreateUserWithHash(ctx, dbgen.CreateUserWithHashParams{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
	})
	if err != nil {
		return nil, mapPgError(err)
	}
	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: []byte(row.PasswordHash),
		Role:         row.Role,
	}, nil
}

func (r *UserRepo) FindByEmail(
	ctx context.Context, email string,
) (*domain.User, error) {
	row, err := r.q.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, mapPgError(err)
	}
	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: []byte(row.PasswordHash),
		Role:         row.Role,
	}, nil
}

func (r *UserRepo) FindByID(
	ctx context.Context, id int64,
) (*domain.User, error) {
	row, err := r.q.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, mapPgError(err)
	}
	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: []byte(row.PasswordHash),
		Role:         row.Role,
	}, nil
}
