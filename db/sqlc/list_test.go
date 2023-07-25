package db

import (
	"context"
	"testing"

	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomList(t *testing.T, u User) List {
	arg := AddListParams{
		Author: u.ID,
		Header: util.RandomString(10),
	}

	list, err := testQueries.AddList(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, list)

	require.NotZero(t, list.ID)
	require.Equal(t, arg.Author, list.Author)
	require.Equal(t, arg.Header, list.Header)

	return list
}

func TestAddList(t *testing.T) {
	newUser := createRandomUser(t)
	createRandomList(t, newUser)
	deleteTestUser(t, newUser)
}

func TestGetLists(t *testing.T) {
	newUser := createRandomUser(t)

	const listsCount = 5
	for i := 0; i < listsCount; i++ {
		createRandomList(t, newUser)
	}

	lists, err := testQueries.GetLists(context.Background(), newUser.ID)

	require.NoError(t, err)
	require.NotEmpty(t, lists)

	require.Equal(t, listsCount+1, len(lists))

	deleteTestUser(t, newUser)
}

func TestDeleteList(t *testing.T) {
	newUser := createRandomUser(t)

	newList := createRandomList(t, newUser)

	err := testQueries.DeleteList(context.Background(), newList.ID)

	require.NoError(t, err)

	lists, err := testQueries.GetLists(context.Background(), newUser.ID)

	require.NoError(t, err)
	require.Equal(t, 1, len(lists))

	deleteTestUser(t, newUser)
}
