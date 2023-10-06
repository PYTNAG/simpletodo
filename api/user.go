package api

import (
	"net/http"
	"strconv"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

type addUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type addUserResponse struct {
	ID int32 `json:"id"`
}

func (s *Server) addUser(ctx *gin.Context) {
	var req addUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		Username: req.Username,
		Hash:     hashPass(req.Password),
	}

	result, err := s.store.CreateUserTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, addUserResponse{ID: result.User.ID}) // TODO: return auth token
}

type deleteUserRequest struct {
	Password string `json:"password" binding:"required"`
}

func (s *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteUserParams{
		ID:   int32(id),
		Hash: hashPass(req.Password),
	}

	r, err := s.store.DeleteUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, r)
}

func hashPass(pass string) []byte {
	return []byte(pass) // TODO : replace with [b/s]crypt alg
}
