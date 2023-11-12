package gapi

import (
	"context"
	"database/sql"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hash, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	createUserResult, err := s.store.CreateUserTx(
		ctx,
		db.CreateUserTxParams{
			Username: req.GetUsername(),
			Hash:     hash,
		},
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "user %s already exist", req.GetUsername())
			}
		}

		return nil, status.Errorf(codes.Internal, "cannot create user: %s", err)
	}

	response := &pb.CreateUserResponse{
		UserId: createUserResult.User.ID,
	}

	return response, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateDeleteUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if payload.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if _, err := s.store.DeleteUser(ctx, req.GetUserId()); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user with id %d doesn't exist", req.GetUserId())
		}

		return nil, status.Errorf(codes.Internal, "cannot delete user: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RehashUser(ctx context.Context, req *pb.RehashUserRequest) (*emptypb.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateRehashUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if payload.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	newHash, err := util.HashPassword(req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := db.RehashUserParams{
		ID:      req.GetUserId(),
		NewHash: newHash,
	}

	if user, err := s.store.RehashUser(ctx, params); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "wrong user id or actual password: %+v", user)
		}

		return nil, status.Errorf(codes.Internal, "failed to rehash user: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	user, err := s.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "there is no user with username %s", req.GetUsername())
		}

		return nil, status.Errorf(codes.Internal, "cannot find user: %s", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.Hash)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "wrong password")
	}

	accesToken, accessPayload, err := s.pasetoMaker.CreateToken(user.Username, user.ID, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create UUID: %s", err)
	}

	refreshToken, refreshPayload, err := s.pasetoMaker.CreateToken(user.Username, user.ID, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create UUID: %s", err)
	}

	metadata := s.extractMetadata(ctx)

	params := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIp:     metadata.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := s.store.CreateSession(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session: %s", err)
	}

	response := &pb.LoginUserResponse{
		UserId:                user.ID,
		SessionId:             session.ID.String(),
		AccessToken:           accesToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}

	return response, nil
}
