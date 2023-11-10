package gapi

import (
	"fmt"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/token"
	"github.com/PYTNAG/simpletodo/util"
)

// Server servers GRPC req-s for todo app
type Server struct {
	pb.UnimplementedSimpleTODOServer

	config      util.Config
	store       db.Store
	pasetoMaker *token.PasetoMaker
}

// NewServer creates a new GRPC server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	pasetoMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:      config,
		store:       store,
		pasetoMaker: pasetoMaker,
	}

	return server, nil
}
