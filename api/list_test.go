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

	testCases := []*apiTestCase{
		{
			name:          "OK",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
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
			checkResponse: requierResponseCode(http.StatusCreated),
		},
		{
			name:          "WrongBody",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   requestBody{},
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AddList(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponse: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
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
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}

func TestGetUserListsAPI(t *testing.T) {
	user := util.RandomUser()

	userLists := []db.GetListsRow{
		{
			ID:     util.RandomID(),
			Header: util.RandomString(8),
		},
		{
			ID:     util.RandomID(),
			Header: util.RandomString(8),
		},
	}

	defaultSettings := struct {
		methodGet string
		url       string
		body      requestBody
		setupAuth setupAuthFunc
	}{
		methodGet: http.MethodGet,
		url:       fmt.Sprintf("/users/%d/lists", user.ID),
		body:      requestBody{},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(userLists, nil).
					After(getUserCall(store, user))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				lists := util.Unmarshal[getUserListsResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)

				require.NotEmpty(t, lists.Lists)
				require.Len(t, lists.Lists, len(userLists))

				for _, list := range lists.Lists {
					require.Contains(t, userLists, list)
				}
			},
		},
		{
			name:          "NoLists",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, sql.ErrNoRows).
					After(getUserCall(store, user))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				lists := util.Unmarshal[getUserListsResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.Empty(t, lists.Lists)
			},
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, sql.ErrConnDone).
					After(getUserCall(store, user))
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
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
		body:         requestBody{},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK",
			requestMethod: defaultSettings.methodDelete,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						DeleteList(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(nil),
				)
			},
			checkResponse: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodDelete,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						DeleteList(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}
