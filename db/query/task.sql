-- name: GetTasks :many
SELECT * FROM tasks
WHERE list_id = $1;

-- name: AddTask :one
INSERT INTO tasks (
	list_id, parent_task, task
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: UpdateCheckTask :exec
UPDATE tasks
	set complete = $2
WHERE id = $1
RETURNING *;

-- name: UpdateTaskText :exec
UPDATE tasks
	set task = $2
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;