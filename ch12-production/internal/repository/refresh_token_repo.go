package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "github.com/forest6511/go-web-textbook-examples/ch12-production/internal/db/gen"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/domain"
)

type RefreshTokenRepo struct {
	pool *pgxpool.Pool
	q    *dbgen.Queries
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool, q: dbgen.New(pool)}
}

// RefreshTokenRow は新規発行時のパラメータ
type RefreshTokenRow struct {
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
}

// RotateResult は Rotate で返す、新 Refresh Token の ID とユーザー情報
type RotateResult struct {
	NewID  pgtype.UUID
	UserID int64
}

// InsertRoot は初回発行（signup / login）時に呼ぶ
// family_id を自分自身の id に揃えるため、Go 側で UUID を事前生成する
func (r *RefreshTokenRepo) InsertRoot(
	ctx context.Context, row *RefreshTokenRow,
) (pgtype.UUID, error) {
	rawID, err := uuid.NewRandom()
	if err != nil {
		return pgtype.UUID{}, err
	}
	pgID := pgtype.UUID{Bytes: rawID, Valid: true}
	inserted, err := r.q.InsertRootRefreshToken(
		ctx,
		dbgen.InsertRootRefreshTokenParams{
			ID:        pgID,
			UserID:    row.UserID,
			TokenHash: row.TokenHash,
			ExpiresAt: row.ExpiresAt,
		},
	)
	if err != nil {
		return pgtype.UUID{}, mapPgError(err)
	}
	return inserted.ID, nil
}

// Rotate は 1 つのトランザクションで used フラグ立てと新規発行を行う
func (r *RefreshTokenRepo) Rotate(
	ctx context.Context, oldHash, newHash []byte,
	newExpiresAt time.Time,
) (*RotateResult, error) {
	var result RotateResult
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return r.rotateInTx(ctx, tx, oldHash, newHash,
			newExpiresAt, &result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *RefreshTokenRepo) rotateInTx(
	ctx context.Context, tx pgx.Tx,
	oldHash, newHash []byte, newExpiresAt time.Time,
	result *RotateResult,
) error {
	q := r.q.WithTx(tx)
	row, err := q.FindRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrRefreshRevoked
		}
		return mapPgError(err)
	}
	if row.RevokedAt.Valid {
		return domain.ErrRefreshRevoked
	}
	if row.ExpiresAt.Before(time.Now()) {
		return domain.ErrRefreshExpired
	}
	// ConsumeRefreshToken は used_at IS NULL のときだけ UPDATE して RETURNING する。
	// 並行 Refresh が 2 回叩かれても、どちらか 1 回だけが行を返す（他方は ErrNoRows）。
	// これで Ch 07 の「SELECT→UPDATE の非原子性による Refresh Token race」を塞ぐ。
	if _, err := q.ConsumeRefreshToken(ctx, row.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 別コネクションで family revoke を確定させる（TX rollback で消えないよう外に出す）
			_, _ = r.pool.Exec(ctx,
				"UPDATE refresh_tokens SET revoked_at = NOW() "+
					"WHERE family_id = $1 AND revoked_at IS NULL",
				row.FamilyID)
			return domain.ErrRefreshReused
		}
		return mapPgError(err)
	}
	return r.rotateInsert(ctx, q, row, newHash, newExpiresAt, result)
}

func (r *RefreshTokenRepo) rotateInsert(
	ctx context.Context, q *dbgen.Queries,
	row dbgen.RefreshToken, newHash []byte,
	newExpiresAt time.Time, result *RotateResult,
) error {
	inserted, err := q.InsertRefreshToken(ctx,
		dbgen.InsertRefreshTokenParams{
			UserID:    row.UserID,
			TokenHash: newHash,
			FamilyID:  row.FamilyID,
			ParentID:  pgtype.UUID{Bytes: row.ID.Bytes, Valid: true},
			ExpiresAt: newExpiresAt,
		})
	if err != nil {
		return mapPgError(err)
	}
	result.NewID = inserted.ID
	result.UserID = row.UserID
	return nil
}

// Revoke は family 単位で失効する（ログアウト時に使用）
func (r *RefreshTokenRepo) Revoke(ctx context.Context, hash []byte) error {
	row, err := r.q.FindRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return mapPgError(err)
	}
	return r.q.RevokeFamily(ctx, row.FamilyID)
}
