-- name: CreateUserWithHash :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, role, created_at;

-- name: FindUserByEmail :one
SELECT id, email, password_hash, role, created_at
FROM users
WHERE email = $1;

-- name: FindUserByID :one
SELECT id, email, password_hash, role, created_at
FROM users
WHERE id = $1;
