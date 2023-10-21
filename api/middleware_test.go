package api

import (
	"database/sql"
	"fmt"
	"net/http"
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

func addAuthorization(
	t *testing.T,
	request *http.Request,
	pasetoMaker *token.PasetoMaker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := pasetoMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	defaultSettings := struct {
		method        string
		url           string
		buildStubs    buildStubsFunc
		setupContext  gin.HandlerFunc
		getMiddleware getMiddlewareFunc
	}{
		method:       http.MethodGet,
		url:          "/auth",
		buildStubs:   func(store *mockdb.MockStore) {},
		setupContext: func(ctx *gin.Context) { ctx.Next() },
		getMiddleware: func(server *Server, store db.Store) gin.HandlerFunc {
			return authMiddleware(*server.pasetoMaker)
		},
	}

	testCases := []*middlewareTestCase{
		{
			name:        "OK",
			requestPath: defaultSettings.url,
			requestUrl:  defaultSettings.url,
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusOK),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "NoAuthorization",
			requestPath:   defaultSettings.url,
			requestUrl:    defaultSettings.url,
			setupAuth:     func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusUnauthorized),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "InvalidAuthorizationFormat",
			requestPath: defaultSettings.url,
			requestUrl:  defaultSettings.url,
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, "", "user", time.Minute)
			},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusUnauthorized),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "UnsupportedAuthorizationType",
			requestPath: defaultSettings.url,
			requestUrl:  defaultSettings.url,
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, "unsupported", "user", time.Minute)
			},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusUnauthorized),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "ExpiredToken",
			requestPath: defaultSettings.url,
			requestUrl:  defaultSettings.url,
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusUnauthorized),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}

func TestIdRequestMiddleware(t *testing.T) {
	idKey := "id"

	defaultSettings := struct {
		method        string
		path          string
		buildStubs    buildStubsFunc
		setupAuth     setupAuthFunc
		setupContext  gin.HandlerFunc
		getMiddleware getMiddlewareFunc
	}{
		method:       http.MethodGet,
		path:         fmt.Sprintf("/:%s", idKey),
		buildStubs:   func(store *mockdb.MockStore) {},
		setupAuth:    func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
		setupContext: func(ctx *gin.Context) { ctx.Next() },
		getMiddleware: func(server *Server, store db.Store) gin.HandlerFunc {
			return idRequestMiddleware(idKey)
		},
	}

	testCases := []*middlewareTestCase{
		{
			name:          "OK",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%d", 1),
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusOK),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "NaN",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%s", "nan"),
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "OutOfRange",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%d", int64(^uint32(0)>>1)+1), // maximum int32 + 1
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "Non-Positive",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%d", 0),
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}

func TestCompareRequestedIdMiddleware(t *testing.T) {
	user := util.RandomUser()

	defaultSettings := struct {
		method        string
		path          string
		setupAuth     setupAuthFunc
		setupContext  gin.HandlerFunc
		getMiddleware getMiddlewareFunc
	}{
		method:    http.MethodGet,
		path:      fmt.Sprintf("/:%s", userIdKey),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
		setupContext: func(ctx *gin.Context) {
			token, _ := token.NewPayload(user.Username, time.Minute)

			ctx.Set(authorizationPayloadKey, token)
			ctx.Set(userIdKey, user.ID)

			ctx.Next()
		},
		getMiddleware: func(server *Server, store db.Store) gin.HandlerFunc {
			return compareRequestedIdMiddleware(store)
		},
	}

	testCases := []*middlewareTestCase{
		{
			name:        "OK",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/%d", user.ID),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{ID: user.ID}, nil)
			},
			checkResponse: requierResponseCode(http.StatusOK),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "AuthorizedUserDoesNotExist",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/%d", user.ID),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: requierResponseCode(http.StatusForbidden),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "InternalError",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/%d", user.ID),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "AccessError",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/%d", user.ID),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{ID: util.RandomID()}, nil)
			},
			checkResponse: requierResponseCode(http.StatusUnauthorized),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}

func TestCheckListAuthorMiddleware(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()

	defaultSettings := struct {
		method        string
		path          string
		setupAuth     setupAuthFunc
		setupContext  gin.HandlerFunc
		getMiddleware getMiddlewareFunc
	}{
		method:    http.MethodGet,
		path:      fmt.Sprintf("/:%s/:%s", userIdKey, listIdKey),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
		setupContext: func(ctx *gin.Context) {
			ctx.Set(userIdKey, user.ID)
			ctx.Set(listIdKey, listId)

			ctx.Next()
		},
		getMiddleware: func(server *Server, store db.Store) gin.HandlerFunc {
			return checkListAuthorMiddleware(store)
		},
	}

	testCases := []*middlewareTestCase{
		{
			name:        "OK",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d", user.ID, listId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{
						{ID: listId},
						{ID: util.RandomID()},
						{ID: util.RandomID()},
					}, nil)
			},
			checkResponse: requierResponseCode(http.StatusOK),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "UserDoesNotHaveLists",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d", user.ID, listId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, sql.ErrNoRows)
			},
			checkResponse: requierResponseCode(http.StatusForbidden),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "InternalError",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d", user.ID, listId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{}, sql.ErrConnDone)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "ListDoesNotExist",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d", user.ID, listId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetLists(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.GetListsRow{
						{ID: util.RandomID()},
						{ID: util.RandomID()},
					}, nil)
			},
			checkResponse: requierResponseCode(http.StatusBadRequest),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}

func TestCheckTaskParentListMiddleware(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()
	taskId := util.RandomID()

	defaultSettings := struct {
		method        string
		path          string
		setupAuth     setupAuthFunc
		setupContext  gin.HandlerFunc
		getMiddleware getMiddlewareFunc
	}{
		method:    http.MethodGet,
		path:      fmt.Sprintf("/:%s/:%s/:%s", userIdKey, listIdKey, taskIdKey),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
		setupContext: func(ctx *gin.Context) {
			ctx.Set(userIdKey, user.ID)
			ctx.Set(listIdKey, listId)
			ctx.Set(taskIdKey, taskId)

			ctx.Next()
		},
		getMiddleware: func(server *Server, store db.Store) gin.HandlerFunc {
			return checkTaskParentListMiddleware(store)
		},
	}

	testCases := []*middlewareTestCase{
		{
			name:        "OK",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d/:%d", user.ID, listId, taskId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTasks(gomock.Any(), gomock.Eq(listId)).
					Times(1).
					Return([]db.Task{
						{ID: taskId},
						{ID: util.RandomID()},
						{ID: util.RandomID()},
					}, nil)
			},
			checkResponse: requierResponseCode(http.StatusOK),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "EmptyList",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d/:%d", user.ID, listId, taskId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTasks(gomock.Any(), gomock.Eq(listId)).
					Times(1).
					Return([]db.Task{}, sql.ErrNoRows)
			},
			checkResponse: requierResponseCode(http.StatusForbidden),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "InternalError",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d/:%d", user.ID, listId, taskId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTasks(gomock.Any(), gomock.Eq(listId)).
					Times(1).
					Return([]db.Task{}, sql.ErrConnDone)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:        "TaskDoesNotExist",
			requestPath: defaultSettings.path,
			requestUrl:  fmt.Sprintf("/:%d/:%d/:%d", user.ID, listId, taskId),
			setupAuth:   defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTasks(gomock.Any(), gomock.Eq(listId)).
					Times(1).
					Return([]db.Task{
						{ID: util.RandomID()},
						{ID: util.RandomID()},
					}, nil)
			},
			checkResponse: requierResponseCode(http.StatusBadRequest),
			setupContext:  defaultSettings.setupContext,
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}
