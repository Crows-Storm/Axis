package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/ledgerpb"
	"github.com/Crows-Storm/Axis/ledger/app"
)

type GRPCServer struct {
	ledgerpb.UnimplementedLedgerServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CheckBalance(ctx context.Context, request *ledgerpb.CheckBalanceRequest) (*ledgerpb.CheckBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}
