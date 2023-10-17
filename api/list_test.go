package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddListToUserAPI(t *testing.T) {
	user := util.RandomUser()
	newListHeader := util.RandomString(8)

	defaultSettings := struct {
		methodPost string
		url        string
		body       requestBody
		setupAuth  setupAuthFunc
	}{
		methodPost: http.MethodPost,
		url:        fmt.Sprintf("/users/%d/lists", user.ID),
		body: requestBody{
			"header": newListHeader,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				addListParams := db.AddListParams{
					Author: user.ID,
					Header: newListHeader,
				}

				store.EXPECT().
					AddList(gomock.Any(), gomock.Eq(addListParams)).
					Times(1).
					Return(db.List{}, nil).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusCreated),
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      emptyRequestBody(),
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					AddList(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				addListParams := db.AddListParams{
					Author: user.ID,
					Header: newListHeader,
				}

				store.EXPECT().
					AddList(gomock.Any(), gomock.Eq(addListParams)).
					Times(1).
					Return(db.List{}, sql.ErrConnDone).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}

func TestGetUserListsAPI(t *testing.T) {
	user := util.RandomUser()

	defaultSettings := struct {
		methodGet string
		url       string
		body      requestBody
		setupAuth setupAuthFunc
	}{
		methodGet: http.MethodGet,
		url:       fmt.Sprintf("/users/%d/lists", user.ID),
		body:      emptyRequestBody(),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodGet,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, nil).
					After(getUserCall(store, user))
			},
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				lists := util.Unmarshal[getUserListsResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.Empty(t, lists.Lists)
			},
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodGet,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, sql.ErrConnDone).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}

func TestDeleteUserListAPI(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()

	defaultSettings := struct {
		methodDelete string
		url          string
		body         requestBody
		setupAuth    setupAuthFunc
	}{
		methodDelete: http.MethodDelete,
		url:          fmt.Sprintf("/users/%d/lists/%d", user.ID, listId),
		body:         emptyRequestBody(),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						DeleteList(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(nil),
				)
			},
			checkResponseHandler: requierResponseCode(http.StatusNoContent),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						DeleteList(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(sql.ErrConnDone),
				)
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}
