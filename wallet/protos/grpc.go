package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/walletpb"
	"github.com/Crows-Storm/Axis/wallet/app"
)

type GRPCServer struct {
	walletpb.UnimplementedWalletServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetWalletByUserId(ctx context.Context, request *walletpb.GetWalletByUserIdRequest) (*walletpb.GetWalletByUserIdResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) GetBalance(ctx context.Context, request *walletpb.GetBalanceRequest) (*walletpb.GetBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) mustEmbedUnimplementedWalletServiceServer() {
	//TODO implement me
	panic("implement me")
}
