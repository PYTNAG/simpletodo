package db

import (
	"context"
	"testing"

	db "github.com/PYTNAG/simpletodo/db/types"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomTask(t *testing.T, list *List, parentTask *Task) *Task {
	var pTask = db.NewNullInt32(0, false)
	if parentTask != nil {
		pTask.Int32 = parentTask.ID
		pTask.Valid = true
	}

	params := AddTaskParams{
		ListID:     list.ID,
		ParentTask: pTask,
		Task:       util.RandomString(24),
	}

	task, err := testQueries.AddTask(context.Background(), params)

	require.NoError(t, err)

	require.NotZero(t, task.ID)
	require.Equal(t, list.ID, task.ListID)

	require.Equal(t, parentTask != nil, task.ParentTask.Valid)
	if parentTask != nil {
		require.Equal(t, parentTask.ID, task.ParentTask.Int32)
	}

	require.Equal(t, params.Task, task.Task)
	require.False(t, task.Complete)

	return &task
}

func TestAddTask(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	createRandomTask(t, defaultList, nil)

	deleteTestUser(t, newUser)
}

func TestAddChildTask(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	parentTask := createRandomTask(t, defaultList, nil)
	createRandomTask(t, defaultList, parentTask)

	deleteTestUser(t, newUser)
}

func TestToggleTask(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	task := createRandomTask(t, defaultList, nil)

	err := testQueries.ToggleTask(context.Background(), task.ID)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), defaultList.ID)

	require.NoError(t, err)

	require.NotEmpty(t, tasks)
	require.True(t, tasks[0].Complete)

	deleteTestUser(t, newUser)
}

func TestDeleteCheckedRootTasks(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	tasksCount := 6
	for i := 0; i < tasksCount; i++ {
		task := createRandomTask(t, defaultList, nil)
		if i%2 == 0 {
			testQueries.ToggleTask(context.Background(), task.ID)
		}
	}

	err := testQueries.DeleteCheckedRootTasks(context.Background(), defaultList.ID)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), defaultList.ID)

	require.NoError(t, err)

	require.Len(t, tasks, tasksCount/2)

	deleteTestUser(t, newUser)
}

func TestUpdateTaskText(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	task := createRandomTask(t, defaultList, nil)

	params := UpdateTaskTextParams{
		ID:   task.ID,
		Task: util.RandomString(10),
	}

	err := testQueries.UpdateTaskText(context.Background(), params)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), defaultList.ID)

	require.NoError(t, err)
	require.Equal(t, params.Task, tasks[0].Task)

	deleteTestUser(t, newUser)
}

func TestDeleteTask(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	task := createRandomTask(t, defaultList, nil)

	err := testQueries.DeleteTask(context.Background(), task.ID)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), defaultList.ID)

	require.NoError(t, err)
	require.Len(t, tasks, 0)

	deleteTestUser(t, newUser)
}

func TestGetChildTasks(t *testing.T) {
	newUser, defaultList := createRandomUser(t, true)

	parentTask := createRandomTask(t, defaultList, nil)

	var childTasksCount int = 5
	childs := make([]*Task, childTasksCount)

	for i := range childs {
		childs[i] = createRandomTask(t, defaultList, parentTask)
	}

	tasks, err := testQueries.GetChildTasks(context.Background(), db.NewNullInt32(parentTask.ID, true))

	require.NoError(t, err)
	require.Len(t, tasks, childTasksCount)

	deleteTestUser(t, newUser)
}
