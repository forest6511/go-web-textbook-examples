-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token_hash, family_id, parent_id, expires_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token_hash, family_id, parent_id,
          used_at, revoked_at, expires_at, created_at;

-- name: InsertRootRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, family_id, expires_at
)
VALUES ($1, $2, $3, $1, $4)
RETURNING id, user_id, token_hash, family_id, parent_id,
          used_at, revoked_at, expires_at, created_at;

-- name: FindRefreshTokenByHash :one
SELECT id, user_id, token_hash, family_id, parent_id,
       used_at, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: MarkRefreshTokenUsed :exec
UPDATE refresh_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: RevokeFamily :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL;

-- name: DeleteRefreshTokenByHash :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1;
