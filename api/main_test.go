package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type setupAuthFunc func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker)
type buildStubsFunc func(store *mockdb.MockStore)
type checkResponseFunc func(t *testing.T, recorder *httptest.ResponseRecorder)
type getMiddlewareFunc func(*Server, db.Store) gin.HandlerFunc

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   token.RandomSymmetricKey,
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func requierResponseCode(code int) checkResponseFunc {
	return func(t *testing.T, recorder *httptest.ResponseRecorder) {
		require.Equal(t, code, recorder.Code)
	}
}

func getUserCall(store *mockdb.MockStore, user util.FullUserInfo) *gomock.Call {
	getUserResult := db.User{
		ID:       user.ID,
		Username: user.Username,
		Hash:     user.Hash,
	}

	return store.EXPECT().
		GetUser(gomock.Any(), gomock.Eq(user.Username)).
		Times(1).
		Return(getUserResult, nil)
}

func getListsCall(store *mockdb.MockStore, userId int32, returnedListId int32) *gomock.Call {
	return store.EXPECT().
		GetLists(gomock.Any(), gomock.Eq(userId)).
		Times(1).
		Return(
			[]db.GetListsRow{
				{ID: returnedListId},
			}, nil)
}

func getTasksCall(store *mockdb.MockStore, listId int32, returnedTaskId int32) *gomock.Call {
	return store.EXPECT().
		GetTasks(gomock.Any(), gomock.Eq(listId)).
		Times(1).
		Return(
			[]db.Task{
				{ID: returnedTaskId},
			}, nil)
}

func unmarshal[T any](t *testing.T, body *bytes.Buffer) *T {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResult T
	err = json.Unmarshal(data, &gotResult)
	require.NoError(t, err)

	return &gotResult
}
