package api

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type createUserResponse struct {
	ID int32 `json:"id"`
}

func (s *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	arg := db.CreateUserTxParams{
		Username: req.Username,
		Hash:     hashPass(req.Password),
	}

	result, err := s.store.CreateUserTx(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "User already exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusCreated, createUserResponse{ID: result.User.ID}) // TODO: return auth token
}

type deleteUserRequest struct {
	Password string `json:"password" binding:"required"`
}

func (s *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	arg := db.DeleteUserParams{
		ID:   int32(id),
		Hash: hashPass(req.Password),
	}

	_, err = s.store.DeleteUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "User doesn't exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type rehashUserRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (s *Server) rehashUser(ctx *gin.Context) {
	var req rehashUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	arg := db.RehashUserParams{
		ID:      int32(id),
		OldHash: hashPass(req.OldPassword),
		NewHash: hashPass(req.NewPassword),
	}

	_, err = s.store.RehashUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, "Wrong actual password")
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func hashPass(pass string) []byte {
	return []byte(pass) // TODO : replace with [b/s]crypt alg
}
