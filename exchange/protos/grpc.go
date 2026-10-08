package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/exchangepb"
	"github.com/Crows-Storm/Axis/exchange/app"
)

const serviceRole = "exchange-integration"

type GRPCServer struct {
	exchangepb.UnimplementedExchangeServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) Ping(_ context.Context, _ *exchangepb.PingRequest) (*exchangepb.Status, error) {
	return &exchangepb.Status{Service: "exchange", Role: serviceRole}, nil
}
