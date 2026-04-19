-- name: InsertAttachment :one
INSERT INTO attachments (
    id, owner_id, object_key, filename, content_type, size_bytes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, owner_id, object_key, filename, content_type, size_bytes, uploaded_at;

-- name: GetAttachmentByID :one
SELECT id, owner_id, object_key, filename, content_type, size_bytes, uploaded_at
FROM attachments
WHERE id = $1;

-- name: DeleteAttachmentByID :exec
DELETE FROM attachments
WHERE id = $1 AND owner_id = $2;
