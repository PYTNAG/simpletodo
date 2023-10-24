package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/PYTNAG/simpletodo/db/mock"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	dbtypes "github.com/PYTNAG/simpletodo/db/types"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetTasksAPI(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()

	listTasks := []db.Task{
		{
			ID:     util.RandomID(),
			ListID: listId,
			Task:   util.RandomString(8),
		},
		{
			ID:     util.RandomID(),
			ListID: listId,
			Task:   util.RandomString(8),
		},
	}

	defaultSettings := struct {
		methodGet string
		url       string
		body      requestBody
		setupAuth setupAuthFunc
	}{
		methodGet: http.MethodGet,
		url:       fmt.Sprintf("/users/%d/lists/%d/tasks", user.ID, listId),
		body:      requestBody{},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(listTasks, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				tasks := util.Unmarshal[getTasksResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.NotEmpty(t, tasks.Tasks)

				require.Len(t, tasks.Tasks, len(listTasks))

				for _, task := range tasks.Tasks {
					require.Contains(t, listTasks, task)
				}
			},
		},
		{
			name:          "EmptyList",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return([]db.Task{}, sql.ErrNoRows),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				tasks := util.Unmarshal[getTasksResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.Empty(t, tasks.Tasks)
			},
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodGet,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return([]db.Task{}, sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}

func TestUpdateTaskAPI(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()
	taskId := util.RandomID()

	newTaskText := util.RandomString(8)

	defaultSettings := struct {
		methodPut string
		url       string
		setupAuth setupAuthFunc
	}{
		methodPut: http.MethodPut,
		url:       fmt.Sprintf("/users/%d/lists/%d/tasks/%d", user.ID, listId, taskId),
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK(Text)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type": "TEXT",
				"text": newTaskText,
			},
			setupAuth: defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.UpdateTaskTextParams{
					ID:   taskId,
					Task: newTaskText,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateTaskText(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(nil),
				)
			},
			checkResponse: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "InternalError(Text)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type": "TEXT",
				"text": newTaskText,
			},
			setupAuth: defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.UpdateTaskTextParams{
					ID:   taskId,
					Task: newTaskText,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateTaskText(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:          "OK(Check)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type":  "CHECK",
				"check": true,
			},
			setupAuth: defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.UpdateCheckTaskParams{
					ID:       taskId,
					Complete: true,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateCheckTask(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(nil),
				)
			},
			checkResponse: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "InternalError(Check)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type":  "CHECK",
				"check": true,
			},
			setupAuth: defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.UpdateCheckTaskParams{
					ID:       taskId,
					Complete: true,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateCheckTask(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:          "WrongBody",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody:   requestBody{},
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateCheckTask(gomock.Any(), gomock.Any()).
						Times(0),
				)
			},
			checkResponse: requierResponseCode(http.StatusBadRequest),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}

func TestAddTaskAPI(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()
	taskId := util.RandomID()

	newTaskText := util.RandomString(8)

	defaultSettings := struct {
		methodPost string
		url        string
		body       requestBody
		setupAuth  setupAuthFunc
	}{
		methodPost: http.MethodPost,
		url:        fmt.Sprintf("/users/%d/lists/%d/tasks", user.ID, listId),
		body: requestBody{
			"task": newTaskText,
		},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK(RootTask)",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: dbtypes.NullInt32{Int32: 0, Valid: false},
					Task:       newTaskText,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						AddTask(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(db.Task{ID: taskId}, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				task := util.Unmarshal[taskResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.NotEmpty(t, task)
				require.Equal(t, taskId, task.ID)
			},
		},
		{
			name:          "OK(ChildTask)",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"task":        newTaskText,
				"parent_task": taskId,
			},
			setupAuth: defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: dbtypes.NullInt32{Int32: taskId, Valid: true},
					Task:       newTaskText,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						AddTask(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(db.Task{ID: taskId}, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				task := util.Unmarshal[taskResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.NotEmpty(t, task)
				require.Equal(t, taskId, task.ID)
			},
		},
		{
			name:          "WrongBody",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   requestBody{},
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						AddTask(gomock.Any(), gomock.Any()).
						Times(0),
				)
			},
			checkResponse: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodPost,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: dbtypes.NullInt32{Int32: 0, Valid: false},
					Task:       newTaskText,
				}

				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						AddTask(gomock.Any(), gomock.Eq(params)).
						Times(1).
						Return(db.Task{}, sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}

func TestDeleteTaskAPI(t *testing.T) {
	user := util.RandomUser()
	listId := util.RandomID()
	taskId := util.RandomID()

	defaultSettings := struct {
		methodDelete string
		url          string
		body         requestBody
		setupAuth    setupAuthFunc
	}{
		methodDelete: http.MethodDelete,
		url:          fmt.Sprintf("/users/%d/lists/%d/tasks/%d", user.ID, listId, taskId),
		body:         requestBody{},
		setupAuth: func(t *testing.T, request *http.Request, pasetoMaker *token.PasetoMaker) {
			addAuthorization(t, request, pasetoMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
	}

	testCases := []*apiTestCase{
		{
			name:          "OK",
			requestMethod: defaultSettings.methodDelete,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						DeleteTask(gomock.Any(), gomock.Eq(taskId)).
						Times(1).
						Return(nil),
				)
			},
			checkResponse: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "WrongTask",
			requestMethod: defaultSettings.methodDelete,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						DeleteTask(gomock.Any(), gomock.Eq(taskId)).
						Times(1).
						Return(sql.ErrNoRows),
				)
			},
			checkResponse: requierResponseCode(http.StatusForbidden),
		},
		{
			name:          "InternalError",
			requestMethod: defaultSettings.methodDelete,
			requestUrl:    defaultSettings.url,
			requestBody:   defaultSettings.body,
			setupAuth:     defaultSettings.setupAuth,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						DeleteTask(gomock.Any(), gomock.Eq(taskId)).
						Times(1).
						Return(sql.ErrConnDone),
				)
			},
			checkResponse: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, apiTestingFunc(tc))
	}
}
