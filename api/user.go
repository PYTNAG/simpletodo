package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type getUserRequest struct {
	ID int32 `uri:"id" binding:"required,number,min=1"`
}

type userResponse struct {
	ID int32 `json:"user_id"`
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

	result, err := s.store.CreateUserTx(
		ctx,
		db.CreateUserTxParams{
			Username: data.Username,
			Hash:     hash,
		},
	)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, errorResponse(err, "User already exist"))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	ctx.JSON(http.StatusCreated, userResponse{ID: result.User.ID})
}

type deleteUserData struct {
	Password string `json:"password" binding:"required,printascii,min=8"`
}

func (s *Server) deleteUser(ctx *gin.Context) {
	var (
		req  getUserRequest
		data deleteUserData
	)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be positive integer"))
		return
	}

	code, errResponse := s.compareRequestedIdWithAuthUser(ctx, req.ID)
	if len(errResponse) != 0 {
		ctx.JSON(code, errorResponse)
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err, "Id must be positive integer"))
		return
	}

	code, errResponse := s.compareRequestedIdWithAuthUser(ctx, req.ID)
	if len(errResponse) != 0 {
		ctx.JSON(code, errResponse)
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

type loginUserResponse struct {
	AccessToken string `json:"access_token"`
	ID          int32  `json:"user_id"`
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	response := loginUserResponse{
		AccessToken: accesToken,
		ID:          user.ID,
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *Server) compareRequestedIdWithAuthUser(ctx *gin.Context, requestedId int32) (code int, errResponse gin.H) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := s.store.GetUser(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return http.StatusForbidden, errorResponse(err, "Authorized user doesn't exist")
		}

		return http.StatusInternalServerError, errorResponse(err, "")
	}

	if user.ID != requestedId {
		err := fmt.Errorf("You can't delete other user")
		return http.StatusUnauthorized, errorResponse(err, "")
	}

	return 0, gin.H{}
}
