package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/marketpb"
	"github.com/Crows-Storm/Axis/market/app"
)

const serviceRole = "market-data"

type GRPCServer struct {
	marketpb.UnimplementedMarketServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) Ping(_ context.Context, _ *marketpb.PingRequest) (*marketpb.Status, error) {
	return &marketpb.Status{Service: "market", Role: serviceRole}, nil
}
