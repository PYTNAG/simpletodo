package api

import (
	"fmt"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/gin-gonic/gin"
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
	pasetoMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot create token maker: %w", err)
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

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(*server.pasetoMaker))

	authRoutes.PUT("/users/:id", server.rehashUser)
	authRoutes.DELETE("/users/:id", server.deleteUser)

	authRoutes.GET("/users/:id/lists", server.getUserLists)
	authRoutes.POST("/users/:id/lists", server.addListToUser)
	authRoutes.DELETE("/users/:user_id/lists/:list_id", server.deleteUserList)

	server.router = router
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
