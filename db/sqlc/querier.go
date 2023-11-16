// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package db

import (
	"context"

	db "github.com/PYTNAG/simpletodo/db/types"
	"github.com/google/uuid"
)

type Querier interface {
	AddList(ctx context.Context, arg AddListParams) (List, error)
	AddTask(ctx context.Context, arg AddTaskParams) (Task, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteList(ctx context.Context, id int32) error
	DeleteTask(ctx context.Context, id int32) error
	DeleteUser(ctx context.Context, id int32) (DeleteUserRow, error)
	GetChildTasks(ctx context.Context, parentTask db.NullInt32) ([]Task, error)
	GetLists(ctx context.Context, author int32) ([]GetListsRow, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	GetTasks(ctx context.Context, listID int32) ([]Task, error)
	GetUser(ctx context.Context, username string) (User, error)
	RehashUser(ctx context.Context, arg RehashUserParams) (User, error)
	ToggleTask(ctx context.Context, id int32) error
	UpdateTaskText(ctx context.Context, arg UpdateTaskTextParams) error
}

var _ Querier = (*Queries)(nil)
