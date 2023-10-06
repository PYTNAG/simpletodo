package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	addUserArg := AddUserParams{
		Username: util.RandomUsername(),
		Hash:     util.RandomByteArray(4),
	}

	user, err := testQueries.AddUser(context.Background(), addUserArg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, addUserArg.Username, user.Username)
	require.Equal(t, addUserArg.Hash, user.Hash)

	require.NotZero(t, user.ID)

	createRandomList(t, user)

	return user
}

func deleteTestUser(t *testing.T, u User) {
	deleteArg := DeleteUserParams{
		ID:   u.ID,
		Hash: u.Hash,
	}
	r, err := testQueries.DeleteUser(context.Background(), deleteArg)

	require.NoError(t, err)
	require.Equal(t, DeleteUserRow{ID: u.ID, Username: u.Username}, r)
}

func TestAddUser(t *testing.T) {
	newUser := createRandomUser(t)
	deleteTestUser(t, newUser)
}

func TestGetUser(t *testing.T) {
	expectedUser := createRandomUser(t)

	arg := GetUserParams{
		Username: expectedUser.Username,
		Hash:     expectedUser.Hash,
	}
	actualUser, err := testQueries.GetUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, actualUser)

	require.Equal(t, expectedUser.ID, actualUser.ID)
	require.Equal(t, expectedUser.Username, actualUser.Username)

	deleteTestUser(t, expectedUser)
}

func TestDeleteUser(t *testing.T) {
	newUser := createRandomUser(t)

	deleteTestUser(t, newUser)

	getArg := GetUserParams{
		Username: newUser.Username,
		Hash:     newUser.Hash,
	}
	noUser, err := testQueries.GetUser(context.Background(), getArg)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	require.Empty(t, noUser)
}

func TestRehashUser(t *testing.T) {
	expectedUser := createRandomUser(t)

	arg := RehashUserParams{
		ID:   expectedUser.ID,
		Hash: util.RandomByteArray(8),
	}
	actualUser, err := testQueries.RehashUser(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, arg.Hash, actualUser.Hash)

	deleteTestUser(t, actualUser)
}
