package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type middlewareTestCase struct {
	name          string
	requestPath   string
	requestUrl    string
	setupAuth     setupAuthFunc
	buildStubs    buildStubsFunc
	checkResponse checkResponseFunc
	setupContext  gin.HandlerFunc
	getMiddleware getMiddlewareFunc
}

func testingMiddlewareFunc(tc *middlewareTestCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		tc.buildStubs(store)

		server := newTestServer(t, store)

		server.router.GET(
			tc.requestPath,
			tc.setupContext,
			tc.getMiddleware(server, store),
			func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			},
		)

		recorder := httptest.NewRecorder()

		request, err := http.NewRequest(http.MethodGet, tc.requestUrl, nil)
		require.NoError(t, err)

		tc.setupAuth(t, request, server.pasetoMaker)

		server.router.ServeHTTP(recorder, request)

		tc.checkResponse(t, recorder)
	}
}
