package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestDeleteUserAPI(t *testing.T) {
	user, _ := util.RandomUser()
	wrongUser, _ := util.RandomUser()

	testCases := []struct {
		name          string
		idRequest     int32
		body          gin.H
		authSetup     func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getResult := db.User{
					ID:       user.ID,
					Username: user.Username,
					Hash:     user.Hash,
				}

				deleteResult := db.DeleteUserRow{
					ID:       user.ID,
					Username: user.Username,
				}

				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(getResult, nil)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(deleteResult, nil).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:      "Wrong ID",
			idRequest: -1,
			body:      gin.H{},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "Wrong body request",
			idRequest: user.ID,
			body:      gin.H{},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{ID: user.ID}, nil)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "Wrong user",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, wrongUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getResult := db.User{
					ID:       wrongUser.ID,
					Username: wrongUser.Username,
					Hash:     wrongUser.Hash,
				}

				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(wrongUser.Username)).
					Times(1).
					Return(getResult, nil)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "Authorized user doesn't exist",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, wrongUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(wrongUser.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:      "Internal Error (GetUser)",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, wrongUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(wrongUser.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "User doesn't exist",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, wrongUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getResult := db.User{
					ID:       user.ID,
					Username: user.Username,
					Hash:     user.Hash,
				}

				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(wrongUser.Username)).
					Times(1).
					Return(getResult, nil)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.DeleteUserRow{}, sql.ErrNoRows).
					After(getUserCall)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:      "Internal Error (DeleteUser)",
			idRequest: user.ID,
			body: gin.H{
				"password": user.Password,
			},
			authSetup: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
				addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, wrongUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getResult := db.User{
					ID:       user.ID,
					Username: user.Username,
					Hash:     user.Hash,
				}

				getUserCall := store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(wrongUser.Username)).
					Times(1).
					Return(getResult, nil)

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.DeleteUserRow{}, sql.ErrConnDone).
					After(getUserCall)
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/users/%d", tc.idRequest)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.authSetup(t, request, server.pasetoMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
