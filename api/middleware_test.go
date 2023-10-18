package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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
		getMiddleware getMiddlewareFunc
	}{
		method:     http.MethodGet,
		url:        "/auth",
		buildStubs: func(store *mockdb.MockStore) {},
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
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "NoAuthorization",
			requestPath:   defaultSettings.url,
			requestUrl:    defaultSettings.url,
			setupAuth:     func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusUnauthorized),
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
		getMiddleware getMiddlewareFunc
	}{
		method:     http.MethodGet,
		path:       fmt.Sprintf("/:%s", idKey),
		buildStubs: func(store *mockdb.MockStore) {},
		setupAuth:  func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
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
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "NaN",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%s", "nan"),
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "OutOfRange",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%d", int64(^uint32(0)>>1)+1), // maximum int32 + 1
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			getMiddleware: defaultSettings.getMiddleware,
		},
		{
			name:          "Non-Positive",
			requestPath:   defaultSettings.path,
			requestUrl:    fmt.Sprintf("/%d", 0),
			setupAuth:     defaultSettings.setupAuth,
			buildStubs:    defaultSettings.buildStubs,
			checkResponse: requierResponseCode(http.StatusBadRequest),
			getMiddleware: defaultSettings.getMiddleware,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingMiddlewareFunc(tc))
	}
}
