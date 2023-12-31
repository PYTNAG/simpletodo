package db

import (
	"context"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUserTx(t *testing.T) {
	store := NewStore(testDB)

	iterations := 5

	errs := make(chan error)
	results := make(chan CreateUserTxResult)

	for i := 0; i < iterations; i++ {
		go func() {
			hash, err := util.HashPassword(util.RandomPassword())

			if err != nil {
				errs <- err
				results <- CreateUserTxResult{}
				return
			}

			result, err := store.CreateUserTx(context.Background(), CreateUserTxParams{
				Username: util.RandomString(24),
				Hash:     hash,
			})

			errs <- err
			results <- result
		}()
	}

	for i := 0; i < iterations; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check user
		require.NotEmpty(t, result.User)
		require.NotZero(t, result.User.ID)

		_, err = store.GetUser(context.Background(), result.User.Username)
		require.NoError(t, err)

		// check list
		require.NotEmpty(t, result.List)
		require.NotZero(t, result.List.ID)
		require.Equal(t, result.User.ID, result.List.Author)
		require.Equal(t, DefaultLIstHeader, result.List.Header)

		lists, err := store.GetLists(context.Background(), result.User.ID)

		require.NoError(t, err)
		require.NotEmpty(t, lists)
		require.Equal(t, 1, len(lists))
		require.Equal(t, DefaultLIstHeader, lists[0].Header)

		deleteTestUser(t, &result.User)
	}
}
