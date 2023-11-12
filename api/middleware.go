package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHaderKey   = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(pasetoMaker token.PasetoMaker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, ""))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, ""))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, ""))
			return
		}

		accessToken := fields[1]
		payload, err := pasetoMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, ""))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func idRequestMiddleware(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		value := ctx.Param(key)

		id, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err, key+" must be 32-bit integer"))
			return
		}

		if id < 1 {
			err := fmt.Errorf("%s must be positive integer", key)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err, ""))
			return
		}

		ctx.Set(key, int32(id))
		ctx.Next()
	}
}

func compareRequestedIdMiddleware(store db.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
		requestedUserId := ctx.MustGet(userIdKey).(int32)

		user, err := store.GetUser(ctx, authPayload.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err, "authorized user doesn't exist"))
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err, ""))
			return
		}

		if user.ID != requestedUserId {
			err := fmt.Errorf("you can't update other user")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, ""))
			return
		}

		ctx.Next()
	}
}

func checkListAuthorMiddleware(store db.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestedUserId := ctx.MustGet(userIdKey).(int32)
		requestedListId := ctx.MustGet(listIdKey).(int32)

		userLists, err := store.GetLists(ctx, requestedUserId)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err, fmt.Sprintf("user %d doesn't have any lists", requestedUserId)))
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err, ""))
			return
		}

		for _, list := range userLists {
			if list.ID == requestedListId {
				ctx.Next()
				return
			}
		}

		err = fmt.Errorf("user %d doesn't have list %d", requestedUserId, requestedListId)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err, ""))
	}
}

func checkTaskParentListMiddleware(store db.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestedUserId := ctx.MustGet(userIdKey).(int32)
		requestedListId := ctx.MustGet(listIdKey).(int32)
		requestedTaskId := ctx.MustGet(taskIdKey).(int32)

		tasks, err := store.GetTasks(ctx, requestedListId)
		if err != nil {
			if err == sql.ErrNoRows {
				additionalMsg := fmt.Sprintf("user %d doesn't have any tasks in list %d", requestedUserId, requestedListId)
				ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err, additionalMsg))
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err, ""))
			return
		}

		for _, task := range tasks {
			if task.ID == requestedTaskId {
				ctx.Next()
				return
			}
		}

		err = fmt.Errorf("user %d doesn't have task %d in list %d", requestedUserId, requestedTaskId, requestedListId)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err, ""))
	}
}
