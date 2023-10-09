package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type fullUserInfo struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Hash     []byte `json:"hash"`
}

func TestCreateUserAPI(t *testing.T) {
	user, err := randomUser()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		body          func(user fullUserInfo) *bytes.Buffer
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Created",
			body: func(user fullUserInfo) *bytes.Buffer {
				return bytes.NewBufferString(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user.Username, user.Password))
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
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
					CreateUserTx(
						gomock.Any(),
						gomock.Cond(func(x any) bool {
							var params = x.(db.CreateUserTxParams)
							if err := util.CheckPassword(user.Password, params.Hash); err != nil {
								return false
							}

							return params.Username == arg.Username
						}),
					).
					Times(1).
					Return(res, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				gotResult := util.Unmarshal[createUserResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.GreaterOrEqual(t, gotResult.ID, int32(1))
			},
		},
		{
			name: "Forbiden/UserAlreadyExist",
			body: func(user fullUserInfo) *bytes.Buffer {
				return bytes.NewBufferString(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user.Username, user.Password))
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(
						gomock.Any(),
						gomock.Cond(func(x any) bool {
							var params = x.(db.CreateUserTxParams)
							if err := util.CheckPassword(user.Password, params.Hash); err != nil {
								return false
							}

							return params.Username == arg.Username
						}),
					).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: func(user fullUserInfo) *bytes.Buffer {
				return bytes.NewBufferString(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user.Username, user.Password))
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(
						gomock.Any(),
						gomock.Cond(func(x any) bool {
							var params = x.(db.CreateUserTxParams)
							if err := util.CheckPassword(user.Password, params.Hash); err != nil {
								return false
							}

							return params.Username == arg.Username
						}),
					).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Request Data",
			body: func(user fullUserInfo) *bytes.Buffer {
				return bytes.NewBufferString("{}")
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

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/users", tc.body(user))

			require.NoError(t, err)

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
