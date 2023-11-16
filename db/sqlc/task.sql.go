// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: task.sql

package db

import (
	"context"

	db "github.com/PYTNAG/simpletodo/db/types"
)

const addTask = `-- name: AddTask :one
INSERT INTO tasks (
	list_id, parent_task, task
) VALUES (
	$1, $2, $3
) RETURNING id, list_id, parent_task, task, complete
`

type AddTaskParams struct {
	ListID     int32        `json:"list_id"`
	ParentTask db.NullInt32 `json:"parent_task"`
	Task       string       `json:"task"`
}

func (q *Queries) AddTask(ctx context.Context, arg AddTaskParams) (Task, error) {
	row := q.db.QueryRowContext(ctx, addTask, arg.ListID, arg.ParentTask, arg.Task)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.ListID,
		&i.ParentTask,
		&i.Task,
		&i.Complete,
	)
	return i, err
}

const deleteTask = `-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1
`

func (q *Queries) DeleteTask(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteTask, id)
	return err
}

const getChildTasks = `-- name: GetChildTasks :many
SELECT id, list_id, parent_task, task, complete FROM tasks
WHERE parent_task = $1
`

func (q *Queries) GetChildTasks(ctx context.Context, parentTask db.NullInt32) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getChildTasks, parentTask)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Task{}
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.ListID,
			&i.ParentTask,
			&i.Task,
			&i.Complete,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasks = `-- name: GetTasks :many
SELECT id, list_id, parent_task, task, complete FROM tasks
WHERE list_id = $1
`

func (q *Queries) GetTasks(ctx context.Context, listID int32) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getTasks, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Task{}
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.ListID,
			&i.ParentTask,
			&i.Task,
			&i.Complete,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const toggleTask = `-- name: ToggleTask :exec
UPDATE tasks
	set complete = not complete
WHERE id = $1
RETURNING id, list_id, parent_task, task, complete
`

func (q *Queries) ToggleTask(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, toggleTask, id)
	return err
}

const updateTaskText = `-- name: UpdateTaskText :exec
UPDATE tasks
	set task = $2
WHERE id = $1
RETURNING id, list_id, parent_task, task, complete
`

type UpdateTaskTextParams struct {
	ID   int32  `json:"id"`
	Task string `json:"task"`
}

func (q *Queries) UpdateTaskText(ctx context.Context, arg UpdateTaskTextParams) error {
	_, err := q.db.ExecContext(ctx, updateTaskText, arg.ID, arg.Task)
	return err
}
