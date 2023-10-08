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
	user := randomUser()

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Created",
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				gotResult := util.Unmarshal[createUserResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.GreaterOrEqual(t, gotResult.ID, int32(0))
			},
		},
		{
			name: "Forbiden/UserAlreadyExist",
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					Username: user.Username,
					Hash:     user.Hash,
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			body := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user.Username, user.Password))

			request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))

			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser() fullUserInfo {
	pass := util.RandomPassword()
	return fullUserInfo{
		ID:       util.RandomInt32(),
		Username: util.RandomUsername(),
		Password: pass,
		Hash:     hashPass(pass),
	}
}
