package gapi

import (
	"context"
	"database/sql"
	"time"

	"github.com/PYTNAG/simpletodo/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) RefreshAccessToken(ctx context.Context, req *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	refreshPayload, err := s.pasetoMaker.VerifyToken(req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "failed to verify refresh token: %s", err)
	}

	if refreshPayload.UserId != payload.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "refresh token doesn't belong user %s", payload.Username)
	}

	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "there is no session for provided refresh token: %s", err)
		}

		return nil, status.Errorf(codes.Internal, "failed to get session: %s", err)
	}

	if session.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "blocked session")
	}

	if session.Username != refreshPayload.Username {
		return nil, status.Error(codes.PermissionDenied, "incorrect session user")
	}

	if session.RefreshToken != req.GetRefreshToken() {
		return nil, status.Error(codes.PermissionDenied, "mismatched session token")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, status.Error(codes.PermissionDenied, "expired session")
	}

	accesToken, accessPayload, err := s.pasetoMaker.CreateToken(refreshPayload.Username, refreshPayload.UserId, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create uuid: %s", err)
	}

	response := &pb.RefreshAccessTokenResponse{
		AccessToken:          accesToken,
		AccessTokenExpiredAt: timestamppb.New(accessPayload.ExpiredAt),
	}

	return response, nil
}
