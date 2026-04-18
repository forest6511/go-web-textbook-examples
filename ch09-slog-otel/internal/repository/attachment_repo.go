package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/domain"
)

type AttachmentRepo interface {
	Create(ctx context.Context, in NewAttachment) (domain.Attachment, error)
	GetByID(ctx context.Context, id uuid.UUID, ownerID int64) (domain.Attachment, error)
	DeleteByID(ctx context.Context, id uuid.UUID, ownerID int64) error
}

type NewAttachment struct {
	ID          uuid.UUID
	OwnerID     int64
	ObjectKey   string
	Filename    string
	ContentType string
	SizeBytes   int64
}

type PgAttachmentRepo struct {
	pool *pgxpool.Pool
}

func NewPgAttachmentRepo(pool *pgxpool.Pool) *PgAttachmentRepo {
	return &PgAttachmentRepo{pool: pool}
}

const insertAttachmentSQL = `
INSERT INTO attachments (
    id, owner_id, object_key, filename, content_type, size_bytes
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, owner_id, object_key, filename, content_type, size_bytes, uploaded_at
`

func (r *PgAttachmentRepo) Create(
	ctx context.Context, in NewAttachment,
) (domain.Attachment, error) {
	var a domain.Attachment
	err := r.pool.QueryRow(ctx, insertAttachmentSQL,
		in.ID, in.OwnerID, in.ObjectKey, in.Filename,
		in.ContentType, in.SizeBytes,
	).Scan(&a.ID, &a.OwnerID, &a.ObjectKey, &a.Filename,
		&a.ContentType, &a.SizeBytes, &a.UploadedAt)
	if err != nil {
		return domain.Attachment{}, mapPgError(err)
	}
	return a, nil
}

const getAttachmentByIDSQL = `
SELECT id, owner_id, object_key, filename, content_type, size_bytes, uploaded_at
FROM attachments
WHERE id = $1 AND owner_id = $2
`

func (r *PgAttachmentRepo) GetByID(
	ctx context.Context, id uuid.UUID, ownerID int64,
) (domain.Attachment, error) {
	var a domain.Attachment
	err := r.pool.QueryRow(ctx, getAttachmentByIDSQL, id, ownerID).
		Scan(&a.ID, &a.OwnerID, &a.ObjectKey, &a.Filename,
			&a.ContentType, &a.SizeBytes, &a.UploadedAt)
	if err != nil {
		return domain.Attachment{}, mapPgError(err)
	}
	return a, nil
}

const deleteAttachmentByIDSQL = `
DELETE FROM attachments
WHERE id = $1 AND owner_id = $2
`

func (r *PgAttachmentRepo) DeleteByID(
	ctx context.Context, id uuid.UUID, ownerID int64,
) error {
	_, err := r.pool.Exec(ctx, deleteAttachmentByIDSQL, id, ownerID)
	return mapPgError(err)
}
