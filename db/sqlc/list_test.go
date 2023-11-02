package db

import (
	"context"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomList(t *testing.T, u *User) *List {
	params := AddListParams{
		Author: u.ID,
		Header: util.RandomString(10),
	}

	list, err := testQueries.AddList(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, list)

	require.NotZero(t, list.ID)
	require.Equal(t, params.Author, list.Author)
	require.Equal(t, params.Header, list.Header)

	return &list
}

func TestAddList(t *testing.T) {
	newUser, _ := createRandomUser(t, true)
	deleteTestUser(t, newUser)
}

func TestGetLists(t *testing.T) {
	newUser, _ := createRandomUser(t, true)

	const listsCount = 5
	for i := 0; i < listsCount; i++ {
		createRandomList(t, newUser)
	}

	lists, err := testQueries.GetLists(context.Background(), newUser.ID)

	require.NoError(t, err)
	require.NotEmpty(t, lists)

	// including default list
	require.Equal(t, listsCount+1, len(lists))

	deleteTestUser(t, newUser)
}

func TestDeleteList(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	err := testQueries.DeleteList(context.Background(), defaultList.ID)

	require.NoError(t, err)

	lists, err := testQueries.GetLists(context.Background(), newUser.ID)

	require.NoError(t, err)
	require.Zero(t, len(lists))

	deleteTestUser(t, newUser)
}
