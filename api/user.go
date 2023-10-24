package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
)

type userResponse struct {
	AccessToken string `json:"access_token"`
	ID          int32  `json:"user_id"`
}

type createUserData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,printascii,min=8"`
}

func (s *Server) createUser(ctx *gin.Context) {
	var data createUserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	hash, err := util.HashPassword(data.Password)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
		return
	}

	createUserResult, err := s.store.CreateUserTx(
		ctx,
		db.CreateUserTxParams{
			Username: data.Username,
			Hash:     hash,
		},
	)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "User "+data.Username+" already exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}
	accesToken, err := s.pasetoMaker.CreateToken(createUserResult.User.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "Cannot create UUID"))
		return
	}

	response := userResponse{
		AccessToken: accesToken,
		ID:          createUserResult.User.ID,
	}

	ctx.JSON(http.StatusCreated, response)
}

type deleteUserData struct {
	Password string `json:"password" binding:"required,printascii,min=8"`
}

func (s *Server) deleteUser(ctx *gin.Context) {
	userId := ctx.MustGet(userIdKey).(int32)

	var data deleteUserData

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	if _, err := s.store.DeleteUser(ctx, userId); err != nil {
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
	userId := ctx.MustGet(userIdKey).(int32)

	var data rehashUserData

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	oldHash, err := util.HashPassword(data.OldPassword)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
		return
	}

	newHash, err := util.HashPassword(data.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err, "Maximum length of password is 72 bytes"))
		return
	}

	arg := db.RehashUserParams{
		ID:      userId,
		OldHash: oldHash,
		NewHash: newHash,
	}

	if _, err := s.store.RehashUser(ctx, arg); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "Wrong user id or actual password"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type loginUserData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var data loginUserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	user, err := server.store.GetUser(ctx, data.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err, fmt.Sprintf("There is no user with username %s", data.Username)))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	err = util.CheckPassword(data.Password, user.Hash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, "Wrong password"))
		return
	}

	accesToken, err := server.pasetoMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "Cannot create UUID"))
		return
	}

	response := userResponse{
		AccessToken: accesToken,
		ID:          user.ID,
	}

	ctx.JSON(http.StatusOK, response)
}
