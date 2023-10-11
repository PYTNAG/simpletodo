package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

type fullUserInfo struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Hash     []byte `json:"hash"`
}

type eqCreateUserTxParamsMatcher struct {
	params   db.CreateUserTxParams
	password string
}

func (e eqCreateUserTxParamsMatcher) Matches(x any) bool {
	params, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, params.Hash)
	if err != nil {
		return false
	}

	e.params.Hash = params.Hash
	return reflect.DeepEqual(e.params, params)
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches params %v and password %v", e.params, e.password)
}

func EqCreateUserTxParams(params db.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{
		params:   params,
		password: password,
	}
}

func TestCreateUserAPI(t *testing.T) {
	user, err := randomUser()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Created",
			body: gin.H{
				"username": user.Username,
				"password": user.Password,
			},
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				res := db.CreateUserTxResult{
					User: db.User{
						ID:       user.ID,
						Username: user.Username,
						Hash:     user.Hash,
					},
					List: db.List{},
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(params, user.Password)).
					Times(1).
					Return(res, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				gotResult := util.Unmarshal[userResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.Greater(t, gotResult.ID, int32(0))
			},
		},
		{
			name: "UserAlreadyExist",
			body: gin.H{
				"username": user.Username,
				"password": user.Password,
			},
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(params, user.Password)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": user.Password,
			},
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(params, user.Password)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Request Data",
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)

			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.pasetoMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser() (fullUserInfo, error) {
	pass := util.RandomPassword()
	hash, err := util.HashPassword(pass)

	if err != nil {
		return fullUserInfo{}, err
	}

	return fullUserInfo{
		ID:       util.RandomID(),
		Username: util.RandomUsername(),
		Password: pass,
		Hash:     hash,
	}, nil
}
