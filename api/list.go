package api

import (
	"net/http"
	"strconv"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type addListToUserRequest struct {
	Header string `json:"header" binding:"required"`
}

func (s *Server) addListToUser(ctx *gin.Context) {
	var req addListToUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	arg := db.AddListParams{
		Author: int32(id),
		Header: req.Header,
	}

	_, err = s.store.AddList(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (s *Server) getUserLists(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	var lists []db.GetListsRow
	lists, err = s.store.GetLists(ctx, int32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, lists)
}
