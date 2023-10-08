package api

import (
	"database/sql"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type getUserRequest struct {
	ID int32 `uri:"id" binding:"required, min=1"`
}

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

type deleteUserData struct {
	Password string `json:"password" binding:"required"`
}

func (s *Server) deleteUser(ctx *gin.Context) {
	var (
		req  getUserRequest
		data deleteUserData
	)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be positive integer and greater than 0"))
		return
	}

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	arg := db.DeleteUserParams{
		ID:   req.ID,
		Hash: hashPass(data.Password),
	}

	if _, err := s.store.DeleteUser(ctx, arg); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "User doesn't exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type rehashUserData struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (s *Server) rehashUser(ctx *gin.Context) {
	var (
		req  getUserRequest
		data rehashUserData
	)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be 32-bit integer"))
		return
	}

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	arg := db.RehashUserParams{
		ID:      req.ID,
		OldHash: hashPass(data.OldPassword),
		NewHash: hashPass(data.NewPassword),
	}

	if _, err := s.store.RehashUser(ctx, arg); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, "Wrong current password")
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
