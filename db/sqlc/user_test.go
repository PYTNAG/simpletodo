package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hash, err := util.HashPassword(util.RandomPassword())
	require.NoError(t, err)

	addUserArg := CreateUserParams{
		Username: util.RandomUsername(),
		Hash:     hash,
	}

	user, err := testQueries.CreateUser(context.Background(), addUserArg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, addUserArg.Username, user.Username)
	require.Equal(t, addUserArg.Hash, user.Hash)

	require.NotZero(t, user.ID)

	createRandomList(t, user)

	return user
}

func deleteTestUser(t *testing.T, u User) {
	r, err := testQueries.DeleteUser(context.Background(), u.ID)

	require.NoError(t, err)
	require.Equal(t, DeleteUserRow{ID: u.ID, Username: u.Username}, r)
}

func TestCreateUser(t *testing.T) {
	newUser := createRandomUser(t)
	deleteTestUser(t, newUser)
}

func TestGetUser(t *testing.T) {
	expectedUser := createRandomUser(t)

	actualUser, err := testQueries.GetUser(context.Background(), expectedUser.Username)

	require.NoError(t, err)
	require.NotEmpty(t, actualUser)

	require.Equal(t, expectedUser.ID, actualUser.ID)
	require.Equal(t, expectedUser.Username, actualUser.Username)

	deleteTestUser(t, expectedUser)
}

func TestDeleteUser(t *testing.T) {
	newUser := createRandomUser(t)

	deleteTestUser(t, newUser)

	noUser, err := testQueries.GetUser(context.Background(), newUser.Username)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	require.Empty(t, noUser)
}

func TestRehashUser(t *testing.T) {
	expectedUser := createRandomUser(t)
	newHash, err := util.HashPassword(util.RandomPassword())

	require.NoError(t, err)

	arg := RehashUserParams{
		ID:      expectedUser.ID,
		OldHash: expectedUser.Hash,
		NewHash: newHash,
	}
	actualUser, err := testQueries.RehashUser(context.Background(), arg)

	require.NoError(t, err)
	require.Equal(t, arg.NewHash, actualUser.Hash)

	deleteTestUser(t, actualUser)
}
