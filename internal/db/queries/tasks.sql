-- name: CreateTask :one
INSERT INTO tasks(user_id, title, description, status)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, title, description, status, created_at, updated_at;

-- name: GetTask :one
SELECT id, user_id, title, description, status, created_at, updated_at
FROM tasks
WHERE id = $1 AND user_id = $2;

-- name: UpdateTask :one
UPDATE tasks
SET title = $1, description = $2, status = $3, updated_at = NOW()
WHERE id = $4 AND user_id = $5
RETURNING id, user_id, title, description, status, created_at, updated_at;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;

-- name: ListTasks :many
SELECT id, user_id, title, description, status, created_at, updated_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC;
