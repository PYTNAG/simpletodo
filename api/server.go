package api

import (
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server servers HTTP req-s for todo app
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}

	router := gin.Default()

	router.POST("/users", server.createUser)
	router.PUT("/users/:id", server.rehashUser)
	router.DELETE("/users/:id", server.deleteUser)

	router.GET("/users/:id/lists", server.getUserLists)
	router.POST("/users/:id/lists", server.addListToUser)

	server.router = router
	return server
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
