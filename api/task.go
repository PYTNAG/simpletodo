package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

func (s *Server) getTasks(ctx *gin.Context) {
	userId := ctx.MustGet(userIdKey).(int32)
	listId := ctx.MustGet(listIdKey).(int32)

	tasks, err := s.store.GetTasks(ctx, listId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, fmt.Sprintf("User %d doesn't have any tasks in list %d", userId, listId)))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

type updateTaskData struct {
	Type  string `json:"type" binding:"required,oneof=CHECK TEXT"`
	Text  string `json:"text" binding:"required_without=Check"`
	Check bool   `json:"check" binding:"required_without=Text,boolean"`
}

func (s *Server) updateTask(ctx *gin.Context) {
	taskId := ctx.MustGet(taskIdKey).(int32)

	var data updateTaskData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	switch data.Type {
	case "CHECK":
		params := db.UpdateCheckTaskParams{
			ID:       taskId,
			Complete: data.Check,
		}
		if err := s.store.UpdateCheckTask(ctx, params); err != nil {
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
	ParentTask int32  `json:"parent_task" binding:"number,min=1"`
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

	arg := db.AddTaskParams{
		ListID:     list_id,
		ParentTask: sql.NullInt32{Int32: data.ParentTask, Valid: data.ParentTask != 0},
		Task:       data.Task,
	}

	task, err := s.store.AddTask(ctx, arg)
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
