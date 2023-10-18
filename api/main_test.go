package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type setupAuthFunc func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker)
type buildStubsFunc func(store *mockdb.MockStore)
type checkResponseFunc func(t *testing.T, recorder *httptest.ResponseRecorder)

type testCase interface {
	handlers() *testHandlers
	method() string
	body() requestBody
	url() string
}

type testHandlers struct {
	setupAuth     setupAuthFunc
	buildStubs    buildStubsFunc
	checkResponse checkResponseFunc
}

func testingFunc(tc testCase) func(*testing.T) {
	handlers := tc.handlers()
	return func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		handlers.buildStubs(store)

		server := newTestServer(t, store)

		recorder := httptest.NewRecorder()

		body := tc.body()
		var data []byte = nil

		if body != nil {
			var err error
			data, err = json.Marshal(body)
			require.NoError(t, err)
		}

		request, err := http.NewRequest(tc.method(), tc.url(), bytes.NewReader(data))
		require.NoError(t, err)

		handlers.setupAuth(t, request, server.pasetoMaker)

		server.router.ServeHTTP(recorder, request)

		handlers.checkResponse(t, recorder)
	}
}

func requierResponseCode(code int) checkResponseFunc {
	return func(t *testing.T, recorder *httptest.ResponseRecorder) {
		require.Equal(t, code, recorder.Code)
	}
}

type defaultTestCase struct {
	name                 string
	requestMethod        string
	requestUrl           string
	requestBody          requestBody
	setupAuthHandler     setupAuthFunc
	buildStubsHandler    buildStubsFunc
	checkResponseHandler checkResponseFunc
}

func (tc *defaultTestCase) method() string {
	return tc.requestMethod
}

func (tc *defaultTestCase) body() requestBody {
	return tc.requestBody
}

func (tc *defaultTestCase) url() string {
	return tc.requestUrl
}

func (tc *defaultTestCase) handlers() *testHandlers {
	return &testHandlers{
		setupAuth:     tc.setupAuthHandler,
		buildStubs:    tc.buildStubsHandler,
		checkResponse: tc.checkResponseHandler,
	}
}

type requestBody gin.H

func (body requestBody) replace(key string, newValue any) requestBody {
	newBody := make(requestBody, len(body))

	if _, ok := body[key]; !ok {
		panic(fmt.Errorf("Body doesn't have field %s", key))
	}

	for field, value := range body {
		if field == key {
			newBody[field] = newValue
			continue
		}

		newBody[field] = value
	}

	return newBody
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
