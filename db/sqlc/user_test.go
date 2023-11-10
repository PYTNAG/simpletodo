package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T, withDefaultList bool) (*User, *List) {
	hash, err := util.HashPassword(util.RandomPassword())
	require.NoError(t, err)

	createUserParams := CreateUserParams{
		Username: util.RandomUsername(),
		Hash:     hash,
	}

	user, err := testQueries.CreateUser(context.Background(), createUserParams)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, createUserParams.Username, user.Username)
	require.Equal(t, createUserParams.Hash, user.Hash)

	require.Greater(t, user.ID, int32(0))

	var defaultList *List = nil

	if withDefaultList {
		defaultList = createRandomList(t, &user)
	}

	return &user, defaultList
}

func deleteTestUser(t *testing.T, u *User) {
	response, err := testQueries.DeleteUser(context.Background(), u.ID)

	require.NoError(t, err)

	require.Equal(t, u.ID, response.ID)
	require.Equal(t, u.Username, response.Username)
}

func TestCreateUser(t *testing.T) {
	newUser, _ := createRandomUser(t, false)
	deleteTestUser(t, newUser)
}

func TestGetUser(t *testing.T) {
	expectedUser, _ := createRandomUser(t, false)

	actualUser, err := testQueries.GetUser(context.Background(), expectedUser.Username)

	require.NoError(t, err)
	require.NotEmpty(t, actualUser)

	require.Equal(t, expectedUser.ID, actualUser.ID)
	require.Equal(t, expectedUser.Username, actualUser.Username)

	deleteTestUser(t, expectedUser)
}

func TestDeleteUser(t *testing.T) {
	newUser, _ := createRandomUser(t, false)

	deleteTestUser(t, newUser)

	noUser, err := testQueries.GetUser(context.Background(), newUser.Username)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	require.Empty(t, noUser)
}

func TestRehashUser(t *testing.T) {
	expectedUser, _ := createRandomUser(t, false)
	newHash, err := util.HashPassword(util.RandomPassword())

	require.NoError(t, err)

	params := RehashUserParams{
		ID:      expectedUser.ID,
		NewHash: newHash,
	}
	actualUser, err := testQueries.RehashUser(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, params.NewHash, actualUser.Hash)

	deleteTestUser(t, &actualUser)
}
