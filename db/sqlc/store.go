package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Provides all functions to execute db queries and transactions
type Store interface {
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	Querier
}

// Provides all functions to execute SQL queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		Queries: New(db),
		db:      db,
	}
}

func (store *SQLStore) execTx(ctx context.Context, callback func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)

	err = callback(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}

type CreateUserTxParams struct {
	Username string `json:"username"`
	Hash     []byte `json:"hash"`
}

type CreateUserTxResult struct {
	User User `json:"user"`
	List List `json:"list"`
}

// Create a new user with one default list
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// create user
		result.User, err = q.AddUser(ctx, AddUserParams{
			Username: arg.Username,
			Hash:     arg.Hash,
		})

		if err != nil {
			return err
		}

		// add default list
		result.List, err = q.AddList(ctx, AddListParams{
			Author: result.User.ID,
			Header: "default",
		})

		return err
	})

	return result, err
}
