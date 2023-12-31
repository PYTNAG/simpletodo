// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package db

import (
	"time"

	db "github.com/PYTNAG/simpletodo/db/types"
	"github.com/google/uuid"
)

type List struct {
	ID     int32  `json:"id"`
	Author int32  `json:"author"`
	Header string `json:"header"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Task struct {
	ID         int32        `json:"id"`
	ListID     int32        `json:"list_id"`
	ParentTask db.NullInt32 `json:"parent_task"`
	Task       string       `json:"task"`
	Complete   bool         `json:"complete"`
}

type User struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Hash     []byte `json:"hash"`
}
