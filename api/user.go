package api

import (
	"database/sql"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type getUserRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

type createUserData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type createUserResponse struct {
	ID int32 `json:"id"`
}

func (s *Server) createUser(ctx *gin.Context) {
	var data createUserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	hash, err := util.HashPassword(data.Password)
	if err != nil {
		if err == bcrypt.ErrPasswordTooLong {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
	}

	arg := db.CreateUserTxParams{
		Username: data.Username,
		Hash:     hash,
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

	if _, err := s.store.DeleteUser(ctx, req.ID); err != nil {
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

	oldHash, err := util.HashPassword(data.OldPassword)
	if err != nil {
		if err == bcrypt.ErrPasswordTooLong {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
	}

	newHash, err := util.HashPassword(data.NewPassword)
	if err != nil {
		if err == bcrypt.ErrPasswordTooLong {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
	}

	arg := db.RehashUserParams{
		ID:      req.ID,
		OldHash: oldHash,
		NewHash: newHash,
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
