package db

import (
	"context"
	"testing"

	db "github.com/PYTNAG/simpletodo/db/types"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/stretchr/testify/require"
)

func createRandomTask(t *testing.T, l List, p *Task) Task {
	var pTask = db.NewNullInt32(0, false)
	if p != nil {
		pTask.Int32 = p.ID
		pTask.Valid = true
	}

	arg := AddTaskParams{
		ListID:     l.ID,
		ParentTask: pTask,
		Task:       util.RandomString(24),
	}

	task, err := testQueries.AddTask(context.Background(), arg)

	require.NoError(t, err)

	require.NotZero(t, task.ID)
	require.Equal(t, l.ID, task.ListID)

	require.Equal(t, p != nil, task.ParentTask.Valid)
	if p != nil {
		require.Equal(t, p.ID, task.ParentTask.Int32)
	}

	require.Equal(t, arg.Task, task.Task)
	require.False(t, task.Complete)

	return task
}

func TestAddTask(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	createRandomTask(t, mainList, nil)

	deleteTestUser(t, newUser)
}

func TestAddChildTask(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	p := createRandomTask(t, mainList, nil)
	createRandomTask(t, mainList, &p)

	deleteTestUser(t, newUser)
}

func TestUpdateCheckTask(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	task := createRandomTask(t, mainList, nil)

	arg := UpdateCheckTaskParams{
		ID:       task.ID,
		Complete: true,
	}

	err := testQueries.UpdateCheckTask(context.Background(), arg)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), mainList.ID)

	require.NoError(t, err)

	require.NotEmpty(t, tasks)
	require.True(t, tasks[0].Complete)

	deleteTestUser(t, newUser)
}

func TestDeleteCheckedRootTasks(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	for i := 0; i < 6; i++ {
		task := createRandomTask(t, mainList, nil)
		arg := UpdateCheckTaskParams{
			ID:       task.ID,
			Complete: i%2 == 0,
		}
		testQueries.UpdateCheckTask(context.Background(), arg)
	}

	err := testQueries.DeleteCheckedRootTasks(context.Background(), mainList.ID)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), mainList.ID)

	require.NoError(t, err)

	require.Equal(t, 3, len(tasks))

	deleteTestUser(t, newUser)
}

func TestUpdateTaskText(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	task := createRandomTask(t, mainList, nil)

	arg := UpdateTaskTextParams{
		ID:   task.ID,
		Task: util.RandomString(10),
	}

	err := testQueries.UpdateTaskText(context.Background(), arg)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), mainList.ID)

	require.NoError(t, err)
	require.Equal(t, arg.Task, tasks[0].Task)

	deleteTestUser(t, newUser)
}

func TestDeleteTask(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	task := createRandomTask(t, mainList, nil)

	err := testQueries.DeleteTask(context.Background(), task.ID)

	require.NoError(t, err)

	tasks, err := testQueries.GetTasks(context.Background(), mainList.ID)

	require.NoError(t, err)
	require.Zero(t, len(tasks))

	deleteTestUser(t, newUser)
}

func TestGetChildTasks(t *testing.T) {
	newUser := createRandomUser(t)
	mainList := createRandomList(t, newUser)

	task := createRandomTask(t, mainList, nil)

	var childs [5]Task

	for i := range childs {
		childs[i] = createRandomTask(t, mainList, &task)
	}

	tasks, err := testQueries.GetChildTasks(context.Background(), db.NewNullInt32(task.ID, true))

	require.NoError(t, err)
	require.Equal(t, 5, len(tasks))

	deleteTestUser(t, newUser)
}
