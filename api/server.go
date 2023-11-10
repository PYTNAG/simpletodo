package api

import (
	"fmt"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
)

const (
	userIdKey = "user_id"
	listIdKey = "list_id"
	taskIdKey = "task_key"
)

// Server servers HTTP req-s for todo app
type Server struct {
	config      util.Config
	store       db.Store
	pasetoMaker *token.PasetoMaker
	router      *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// go-paseto decodes key using hex.DecodeString inside
	pasetoMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:      config,
		store:       store,
		pasetoMaker: pasetoMaker,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	authRoutes := router.Group("/")
	authRoutes.Use(authMiddleware(*server.pasetoMaker))

	userRequestRoutes := server.getNewIdRequestGroup(authRoutes, "/users/:%s", userIdKey)
	userRequestRoutes.Use(compareRequestedIdMiddleware(server.store))

	listRequestRoutes := server.getNewIdRequestGroup(userRequestRoutes, "/lists/:%s", listIdKey)
	listRequestRoutes.Use(checkListAuthorMiddleware(server.store))

	taskRequestRoutes := server.getNewIdRequestGroup(listRequestRoutes, "/tasks/:%s", taskIdKey)
	taskRequestRoutes.Use(checkTaskParentListMiddleware(server.store))

	// lists
	userRequestRoutes.GET("/lists", server.getUserLists)
	userRequestRoutes.POST("/lists", server.addListToUser)
	listRequestRoutes.DELETE("", server.deleteUserList)

	// tasks
	listRequestRoutes.GET("/tasks", server.getTasks)
	listRequestRoutes.POST("/tasks", server.addTask)
	taskRequestRoutes.PUT("", server.updateTask)
	taskRequestRoutes.DELETE("", server.deleteTask)

	// tokens
	router.POST("/tokens/refresh_access", server.refreshAccessToken)

	server.router = router
}

func (s *Server) getNewIdRequestGroup(prev *gin.RouterGroup, pathTemplate, idKey string) *gin.RouterGroup {
	idRequestGroup := prev.Group(fmt.Sprintf(pathTemplate, idKey))
	idRequestGroup.Use(idRequestMiddleware(idKey))

	return idRequestGroup
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error, additionalMessage string) gin.H {
	response := gin.H{}

	if err != nil {
		response["error"] = err.Error()
	}

	if additionalMessage != "" {
		response["message"] = additionalMessage
	}

	return response
}
