package db

import (
	"context"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUserTx(t *testing.T) {
	store := NewStore(testDB)

	n := 5

	errs := make(chan error)
	results := make(chan CreateUserTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.CreateUserTx(context.Background(), CreateUserTxParams{
				Username: util.RandomString(24),
				Hash:     util.RandomByteArray(10),
			})

			errs <- err
			results <- result
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check user
		require.NotEmpty(t, result.User)
		require.NotZero(t, result.User.ID)

		_, err = store.GetUser(context.Background(), GetUserParams{
			Username: result.User.Username,
			Hash:     result.User.Hash,
		})
		require.NoError(t, err)

		// check list
		require.NotEmpty(t, result.List)
		require.NotZero(t, result.List.ID)
		require.Equal(t, result.User.ID, result.List.Author)
		require.Equal(t, "default", result.List.Header)

		lists, err := store.GetLists(context.Background(), result.User.ID)

		require.NoError(t, err)
		require.NotEmpty(t, lists)
		require.Equal(t, 1, len(lists))
		require.Equal(t, "default", lists[0].Header)
	}
}
