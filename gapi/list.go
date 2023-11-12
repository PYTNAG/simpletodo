package gapi

import (
	"context"
	"database/sql"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateList(ctx context.Context, req *pb.CreateListRequest) (*emptypb.Empty, error) {
	violations := validateCreateList(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	params := db.AddListParams{
		Author: req.GetUserId(),
		Header: req.GetHeader(),
	}

	if _, err := s.store.AddList(ctx, params); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create list: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetLists(req *pb.GetListsRequest, stream pb.SimpleTODO_GetListsServer) error {
	lists, err := s.store.GetLists(stream.Context(), req.GetUserId())
	if err != nil && err != sql.ErrNoRows {
		return status.Errorf(codes.Internal, "failed to get lists: %s", err)
	}

	for _, list := range lists {
		msg := &pb.List{
			Id:     list.ID,
			Header: list.Header,
		}
		if err := stream.Send(msg); err != nil {
			return status.Errorf(codes.Internal, "failed to send list: %s", err)
		}
	}

	return nil
}

func (s *Server) DeleteList(ctx context.Context, req *pb.DeleteListRequest) (*emptypb.Empty, error) {
	err := s.store.DeleteList(ctx, req.GetListId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete list: %s", err)
	}

	return &emptypb.Empty{}, nil
}
