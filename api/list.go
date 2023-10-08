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
	var (
		req  getUserRequest
		data addListToUserData
	)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
	}

	arg := db.AddListParams{
		Author: req.ID,
		Header: data.Header,
	}

	if _, err := s.store.AddList(ctx, arg); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (s *Server) getUserLists(ctx *gin.Context) {
	var req getUserRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	lists, err := s.store.GetLists(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, lists)
}
