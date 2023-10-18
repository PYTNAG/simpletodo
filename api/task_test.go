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

	testCases := []*defaultTestCase{
		{
			name:             "OK",
			requestMethod:    defaultSettings.methodGet,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return(listTasks, nil),
				)
			},
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
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
			name:             "EmptyList",
			requestMethod:    defaultSettings.methodGet,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return([]db.Task{}, sql.ErrNoRows),
				)
			},
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				tasks := util.Unmarshal[getTasksResponse](t, recorder.Body)

				require.Equal(t, http.StatusOK, recorder.Code)
				require.Empty(t, tasks.Tasks)
			},
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodGet,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						GetTasks(gomock.Any(), gomock.Eq(listId)).
						Times(1).
						Return([]db.Task{}, sql.ErrConnDone),
				)
			},
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
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

	testCases := []*defaultTestCase{
		{
			name:          "OK(Text)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type": "TEXT",
				"text": newTaskText,
			},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
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
			checkResponseHandler: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "InternalError(Text)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type": "TEXT",
				"text": newTaskText,
			},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
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
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:          "OK(Check)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type":  "CHECK",
				"check": true,
			},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
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
			checkResponseHandler: requierResponseCode(http.StatusNoContent),
		},
		{
			name:          "InternalError(Check)",
			requestMethod: defaultSettings.methodPut,
			requestUrl:    defaultSettings.url,
			requestBody: requestBody{
				"type":  "CHECK",
				"check": true,
			},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
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
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodPut,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),
					getTasksCall(store, listId, taskId),

					store.EXPECT().
						UpdateCheckTask(gomock.Any(), gomock.Any()).
						Times(0),
				)
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
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

	testCases := []*defaultTestCase{
		{
			name:             "OK(RootTask)",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: sql.NullInt32{Int32: 0, Valid: false},
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
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
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
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: sql.NullInt32{Int32: taskId, Valid: true},
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
			checkResponseHandler: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				task := util.Unmarshal[taskResponse](t, recorder.Body)

				require.Equal(t, http.StatusCreated, recorder.Code)
				require.NotEmpty(t, task)
				require.Equal(t, taskId, task.ID)
			},
		},
		{
			name:             "WrongBody",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      requestBody{},
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				gomock.InOrder(
					getUserCall(store, user),
					getListsCall(store, user.ID, listId),

					store.EXPECT().
						AddTask(gomock.Any(), gomock.Any()).
						Times(0),
				)
			},
			checkResponseHandler: requierResponseCode(http.StatusBadRequest),
		},
		{
			name:             "InternalError",
			requestMethod:    defaultSettings.methodPost,
			requestUrl:       defaultSettings.url,
			requestBody:      defaultSettings.body,
			setupAuthHandler: defaultSettings.setupAuth,
			buildStubsHandler: func(store *mockdb.MockStore) {
				params := db.AddTaskParams{
					ListID:     listId,
					ParentTask: sql.NullInt32{Int32: 0, Valid: false},
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
			checkResponseHandler: requierResponseCode(http.StatusInternalServerError),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testingFunc(tc))
	}
}
