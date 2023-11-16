-- name: GetTasks :many
SELECT * FROM tasks
WHERE list_id = $1;

-- name: GetChildTasks :many
SELECT * FROM tasks
WHERE parent_task = $1;

-- name: AddTask :one
INSERT INTO tasks (
	list_id, parent_task, task
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: ToggleTask :exec
UPDATE tasks
	set complete = not complete
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

-- name: GetTaskAuthor :one
SELECT users.id FROM 
users
JOIN lists ON users.id = lists.author
JOIN tasks ON lists.id = tasks.list_id
WHERE tasks.id = $1;