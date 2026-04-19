-- name: InsertAudit :exec
INSERT INTO audits (action, task_id) VALUES ($1, $2);
