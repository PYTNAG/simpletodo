package api

import (
	"database/sql"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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

func TestCreateUserAPI(t *testing.T) {
	user := util.RandomUser()

	defaultSettings := struct {
		methodPost         string
		url                string
		body               requestBody
		setupAuth          setupAuthFunc
		createUserTxParams db.CreateUserTxParams
	}{
		methodPost: http.MethodPost,
		url:        "/users",
		body: requestBody{
			"username": user.Username,
			"password": user.Password,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
		createUserTxParams: db.CreateUserTxParams{
			Username: user.Username,
			Hash:     user.Hash,
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				res := db.CreateUserTxResult{
					User: db.User{
						ID:       user.ID,
						Username: user.Username,
						Hash:     user.Hash,
					},
					List: db.List{},
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(defaultSettings.createUserTxParams, user.Password)).
					Times(1).
					Return(res, nil)
			},
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				gotResult := util.Unmarshal[userResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.Greater(t, gotResult.ID, int32(0))
			},
		},
		{
			name:             "UserAlreadyExist",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(defaultSettings.createUserTxParams, user.Password)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrNoRows)
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(defaultSettings.createUserTxParams, user.Password)).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:             "Too long password",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body.replace("password", util.RandomString(73)),
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "Invalid Request Data",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}

func TestDeleteUserAPI(t *testing.T) {
	user := util.RandomUser()

	defaultSettings := struct {
		methodDelete string
		url          string
		body         requestBody
		setupAuth    setupAuthFunc
	}{
		methodDelete: http.MethodDelete,
		url:          fmt.Sprintf("/users/%d", user.ID),
		body: requestBody{
			"password": user.Password,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				deleteResult := db.DeleteUserRow{
					ID:       user.ID,
					Username: user.Username,
				}

				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(deleteResult, nil).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusNoContent),
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:             "WrongUser",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.DeleteUserRow{}, sql.ErrNoRows).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodDelete,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.DeleteUserRow{}, sql.ErrConnDone).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}

func TestRehashUserAPI(t *testing.T) {
	user := util.RandomUser()
	newPass := util.RandomPassword()
	newHash, _ := util.HashPassword(newPass)

	defaultSettings := struct {
		methodPut string
		url       string
		body      requestBody
		setupAuth setupAuthFunc
	}{
		methodPut: http.MethodPut,
		url:       fmt.Sprintf("/users/%d", user.ID),
		body: requestBody{
			"old_password": user.Password,
			"new_password": newPass,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
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
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusNoContent),
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					RehashUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:             "TooLongOldPassword",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body.replace("old_password", util.RandomString(73)),
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					RehashUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "TooLongNewPassword",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body.replace("new_password", util.RandomString(73)),
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					RehashUser(gomock.Any(), gomock.Any()).
					Times(0).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "WrongUserIdOrPassword",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				rehashParams := db.RehashUserParams{
					ID:      user.ID,
					OldHash: user.Hash,
					NewHash: newHash,
				}

				store.EXPECT().
					RehashUser(gomock.Any(), EqRehashUserParams(rehashParams, user.Password, newPass)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusForbidden),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				rehashParams := db.RehashUserParams{
					ID:      user.ID,
					OldHash: user.Hash,
					NewHash: newHash,
				}

				store.EXPECT().
					RehashUser(gomock.Any(), EqRehashUserParams(rehashParams, user.Password, newPass)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone).
					After(getUserCall(store, user))
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}

func TestLoginUserAPI(t *testing.T) {
	user := util.RandomUser()

	defaultSettings := struct {
		methodPost string
		url        string
		body       requestBody
		setupAuth  setupAuthFunc
	}{
		methodPost: http.MethodPost,
		url:        "/users/login",
		body: requestBody{
			"username": user.Username,
			"password": user.Password,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {},
	}

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{
						ID:       user.ID,
						Username: user.Username,
						Hash:     user.Hash,
					}, nil)
			},
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				gotResult := util.Unmarshal[loginUserResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, gotResult.ID, user.ID)
				require.NotEmpty(t, gotResult.AccessToken)
			},
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:             "WrongUsername",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponseHandler: requierResponseCode(http.StatusNotFound),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:             "WrongPassword",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body.replace("password", util.RandomPassword()),
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{
						ID:       user.ID,
						Username: user.Username,
						Hash:     user.Hash,
					}, nil)
			},
			checkResponseHandler: requierResponseCode(http.StatusUnauthorized),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}
