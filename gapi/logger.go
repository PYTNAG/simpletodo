package gapi

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logInfo struct {
	method        string
	statusCode    int
	statusMessage string
	duration      time.Duration
	requestType   string
}

func UnaryGRPCLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if status, ok := status.FromError(err); ok {
		statusCode = status.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logInfo := logInfo{
		method:        info.FullMethod,
		statusCode:    int(statusCode),
		statusMessage: statusCode.String(),
		duration:      duration,
		requestType:   "unary",
	}

	addLogInfo(logger, logInfo)

	return result, err
}

func ServerGRPCLogger(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	startTime := time.Now()
	err := handler(srv, ss)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logInfo := logInfo{
		method:        info.FullMethod,
		statusCode:    int(statusCode),
		statusMessage: statusCode.String(),
		duration:      duration,
		requestType:   "server-stream",
	}

	addLogInfo(logger, logInfo)

	return handler(srv, ss)
}

func addLogInfo(logger *zerolog.Event, info logInfo) {
	logger.
		Str("protocol", "grpc").
		Str("method", info.method).
		Int("status_code", info.statusCode).
		Str("status_message", info.statusMessage).
		Dur("duration", info.duration).
		Msgf("received a gRPC %s request", info.requestType)
}
