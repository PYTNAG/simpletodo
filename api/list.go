package api

import (
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type addListToUserData struct {
	Header string `json:"header" binding:"required"`
}

func (s *Server) addListToUser(ctx *gin.Context) {
	var data addListToUserData

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	arg := db.AddListParams{
		Author: ctx.MustGet(userIdKey).(int32),
		Header: data.Header,
	}

	if _, err := s.store.AddList(ctx, arg); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (s *Server) getUserLists(ctx *gin.Context) {
	lists, err := s.store.GetLists(ctx, ctx.MustGet(userIdKey).(int32))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, lists)
}

func (s *Server) deleteUserList(ctx *gin.Context) {
	listId := ctx.MustGet(listIdKey).(int32)

	err := s.store.DeleteList(ctx, listId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
