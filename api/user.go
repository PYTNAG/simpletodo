package api

import (
	"net/http"

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

	arg := db.AddUserParams{
		Username: req.Username,
		Hash:     []byte(req.Password), // TODO: replace with [b/s]crypt alg
	}

	user, err := s.store.AddUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, addUserResponse{ID: user.ID}) // TODO: return auth token
}
