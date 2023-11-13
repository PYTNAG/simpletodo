package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
		ctx.JSON(http.StatusForbidden, errorResponse(err, "maximum length of password is 72 bytes"))
		return
	}

	params := db.CreateUserTxParams{
		Username: data.Username,
		Hash:     hash,
	}

	createUserResult, err := s.store.CreateUserTx(ctx, params)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "user "+data.Username+" already exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"user_id": createUserResult.User.ID})
}

func (s *Server) deleteUser(ctx *gin.Context) {
	userId := ctx.MustGet(userIdKey).(int32)

	if _, err := s.store.DeleteUser(ctx, userId); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "user doesn't exist"))
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
		ctx.JSON(http.StatusForbidden, errorResponse(err, "maximum length of password is 72 bytes"))
		return
	}

	newHash, err := util.HashPassword(data.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err, "maximum length of password is 72 bytes"))
		return
	}

	params := db.RehashUserParams{
		ID:      userId,
		OldHash: oldHash,
		NewHash: newHash,
	}

	if _, err := s.store.RehashUser(ctx, params); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "wrong user id or actual password"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type loginUserResponse struct {
	SessionsID            uuid.UUID `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	ID                    int32     `json:"user_id"`
}

type loginUserData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) loginUser(ctx *gin.Context) {
	var data loginUserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	user, err := s.store.GetUser(ctx, data.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err, fmt.Sprintf("there is no user with username %s", data.Username)))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	err = util.CheckPassword(data.Password, user.Hash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, "wrong password"))
		return
	}

	accesToken, accessPayload, err := s.pasetoMaker.CreateToken(user.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "cannot create UUID"))
		return
	}

	refreshToken, refreshPayload, err := s.pasetoMaker.CreateToken(user.Username, s.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "cannot create UUID"))
	}

	params := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := s.store.CreateSession(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "cannot create UUID"))
	}

	response := loginUserResponse{
		SessionsID:            session.ID,
		AccessToken:           accesToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		ID:                    user.ID,
	}

	ctx.JSON(http.StatusOK, response)
}
