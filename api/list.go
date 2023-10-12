package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type getUserListRequest struct {
	UserID int32 `uri:"user_id" binding:"required,number,min=1"`
	ListID int32 `uri:"list_id" binding:"required,number,min=1"`
}

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
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	lists, err := s.store.GetLists(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusOK, lists)
}

func (s *Server) deleteUserList(ctx *gin.Context) {
	var req getUserListRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Both UserId and ListId must be positive integer"))
		return
	}

	code, errResponse := s.compareRequestedIdWithAuthUser(ctx, req.UserID)
	if len(errResponse) != 0 {
		ctx.JSON(code, errResponse)
		return
	}

	lists, err := s.store.GetLists(ctx, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, fmt.Sprintf("User %d doesn't have any lists", req.UserID)))
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	if !isThereListWithId(lists, req.ListID) {
		err := fmt.Errorf("User %d doesn't have list %d", req.UserID, req.ListID)
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	err = s.store.DeleteList(ctx, req.ListID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func isThereListWithId(lists []db.GetListsRow, id int32) bool {
	for _, list := range lists {
		if list.ID == id {
			return true
		}
	}

	return false
}
