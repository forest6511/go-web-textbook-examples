-- name: CreateTask :one
INSERT INTO tasks (user_id, title, status)
VALUES ($1, $2, $3)
RETURNING id, user_id, title, status, created_at, updated_at;

-- name: GetTask :one
SELECT id, user_id, title, status, created_at, updated_at
FROM tasks
WHERE id = $1 AND user_id = $2;

-- name: ListTasksByUser :many
SELECT id, user_id, title, status, created_at, updated_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTaskStatus :execrows
UPDATE tasks
SET status = $1, updated_at = NOW()
WHERE id = $2 AND user_id = $3;

-- name: DeleteTask :execrows
DELETE FROM tasks WHERE id = $1 AND user_id = $2;
