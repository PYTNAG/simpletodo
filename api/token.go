package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type refreshAccessTokenData struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) refreshAccessToken(ctx *gin.Context) {
	var data refreshAccessTokenData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, ""))
		return
	}

	refreshPayload, err := s.pasetoMaker.VerifyToken(data.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, ""))
		return
	}

	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err, ""))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err, ""))
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("Blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, ""))
		return
	}

	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("Incorrect session user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, ""))
		return
	}

	if session.RefreshToken != data.RefreshToken {
		err := fmt.Errorf("Mismatched session token")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, ""))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("Expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err, ""))
		return
	}

	accesToken, accessPayload, err := s.pasetoMaker.CreateToken(refreshPayload.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, "Cannot create UUID"))
		return
	}

	response := refreshAccessTokenResponse{
		AccessToken:          accesToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, response)
}
