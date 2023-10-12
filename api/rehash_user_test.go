package api

import (
	"bytes"
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

type eqRehashUserParamsMatcher struct {
	params      db.RehashUserParams
	oldPassword string
	newPassword string
}

func (e eqRehashUserParamsMatcher) Matches(x any) bool {
	params, ok := x.(db.RehashUserParams)
	if !ok {
		return false
	}

	errOldHash := util.CheckPassword(e.oldPassword, params.OldHash)
	errNewHash := util.CheckPassword(e.newPassword, params.NewHash)
	if errOldHash != nil || errNewHash != nil {
		return false
	}

	e.params.OldHash = params.OldHash
	e.params.NewHash = params.NewHash
	return reflect.DeepEqual(e.params, params)
}

func (e eqRehashUserParamsMatcher) String() string {
	return fmt.Sprintf("matches params %v and passwords {old: %v ; new: %v}", e.params, e.oldPassword, e.newPassword)
}

func EqRehashUserParams(params db.RehashUserParams, oldPassword, newPassword string) gomock.Matcher {
	return eqRehashUserParamsMatcher{
		params:      params,
		oldPassword: oldPassword,
		newPassword: newPassword,
	}
}

func TestRehashUserAPI(t *testing.T) {
	user, _ := util.RandomUser()
	newPass := util.RandomPassword()
	newHash, err := util.HashPassword(newPass)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		idRequest     int32
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			idRequest: user.ID,
			body: gin.H{
				"old_password": user.Password,
				"new_password": newPass,
			},
			setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getResult := db.User{
					ID:       user.ID,
					Username: user.Username,
					Hash:     user.Hash,
				}

				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(1).
					Return(getResult, nil)

				rehashParams := db.RehashUserParams{
					ID:      user.ID,
					OldHash: user.Hash,
					NewHash: newHash,
				}

				rehashResult := db.User{
					ID:       user.ID,
					Username: user.Username,
					Hash:     newHash,
				}

				store.EXPECT().
					RehashUser(gomock.Any(), EqRehashUserParams(rehashParams, user.Password, newPass)).
					Times(1).
					Return(rehashResult, nil).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/users/%d", tc.idRequest)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))

			recorder := httptest.NewRecorder()

			tc.setupAuth(t, request, server.pasetoMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
