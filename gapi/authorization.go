package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/PYTNAG/simpletodo/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (s *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)

	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authType)
	}

	accessToken := fields[1]
	payload, err := s.pasetoMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}

func (s *Server) isListAccessAllowed(ctx context.Context, accessPayload *token.Payload, listId int32) error {
	lists, err := s.store.GetLists(ctx, accessPayload.UserId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to get \"%s\"'s lists", accessPayload.Username)
	}

	for _, list := range lists {
		if list.ID == listId {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "user %s doesn't have list with id %d", accessPayload.Username, listId)
}

func (s *Server) isTaskAccessAllowed(ctx context.Context, accessPayload *token.Payload, taskId int32) error {
	author, err := s.store.GetTaskAuthor(ctx, taskId)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get author of task: %s", err)
	}

	if author != accessPayload.UserId {
		return status.Errorf(codes.PermissionDenied, "user %s doesn't have task %d", accessPayload.Username, taskId)
	}

	return nil
}
