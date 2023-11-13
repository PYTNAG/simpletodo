package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	dbtypes "github.com/PYTNAG/simpletodo/db/types"
	"github.com/gin-gonic/gin"
)

type getTasksResponse struct {
	Tasks []db.Task `json:"tasks"`
}

func (s *Server) getTasks(ctx *gin.Context) {
	listId := ctx.MustGet(listIdKey).(int32)

	tasks, err := s.store.GetTasks(ctx, listId)
	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, getTasksResponse{Tasks: tasks})
}

type updateTaskData struct {
	Type string `json:"type" binding:"required,oneof=CHECK TEXT"`
	Text string `json:"text" binding:"required_if=Type TEXT"`
}

func (s *Server) updateTask(ctx *gin.Context) {
	taskId := ctx.MustGet(taskIdKey).(int32)

	var data updateTaskData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	switch strings.ToUpper(data.Type) {
	case "CHECK":
		if err := s.store.ToggleTask(ctx, taskId); err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
			return
		}
	case "TEXT":
		params := db.UpdateTaskTextParams{
			ID:   taskId,
			Task: data.Text,
		}
		if err := s.store.UpdateTaskText(ctx, params); err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
			return
		}
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type addTaskData struct {
	ParentTask int32  `json:"parent_task" binding:"omitempty,number,min=1"`
	Task       string `json:"task" binding:"required"`
}

type taskResponse struct {
	ID int32 `json:"created_task_id"`
}

func (s *Server) addTask(ctx *gin.Context) {
	list_id := ctx.MustGet(listIdKey).(int32)

	var data addTaskData

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	params := db.AddTaskParams{
		ListID:     list_id,
		ParentTask: dbtypes.NewNullInt32(data.ParentTask, data.ParentTask > 0),
		Task:       data.Task,
	}

	task, err := s.store.AddTask(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}
	ctx.JSON(http.StatusCreated, taskResponse{ID: task.ID})
}

func (s *Server) deleteTask(ctx *gin.Context) {
	taskId := ctx.MustGet(taskIdKey).(int32)

	err := s.store.DeleteTask(ctx, taskId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, fmt.Sprintf("There is no task %d", taskId)))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
