package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type apiTestCase struct {
	name          string
	requestMethod string
	requestUrl    string
	requestBody   requestBody
	setupAuth     setupAuthFunc
	buildStubs    buildStubsFunc
	checkResponse checkResponseFunc
}

func apiTestingFunc(tc *apiTestCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		tc.buildStubs(store)

		server := newTestServer(t, store)

		recorder := httptest.NewRecorder()

		var data []byte = nil

		if tc.requestBody != nil {
			var err error
			data, err = json.Marshal(tc.requestBody)
			require.NoError(t, err)
		}

		request, err := http.NewRequest(tc.requestMethod, tc.requestUrl, bytes.NewReader(data))
		require.NoError(t, err)

		tc.setupAuth(t, request, server.pasetoMaker)

		server.router.ServeHTTP(recorder, request)

		tc.checkResponse(t, recorder)
	}
}
