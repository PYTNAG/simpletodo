package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcUserAgentHeader = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (s *Server) extractMetadata(ctx context.Context) *Metadata {
	meta := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcUserAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		meta.ClientIP = p.Addr.String()
	}

	return meta
}
