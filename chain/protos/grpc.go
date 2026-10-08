package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/chain/app"
	"github.com/Crows-Storm/Axis/common/genproto/chainpb"
)

const serviceRole = "blockchain-integration"

type GRPCServer struct {
	chainpb.UnimplementedChainServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) Ping(_ context.Context, _ *chainpb.PingRequest) (*chainpb.Status, error) {
	return &chainpb.Status{Service: "chain", Role: serviceRole}, nil
}
