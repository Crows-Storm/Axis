package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/user/app"
)

type GRPCServer struct {
	userpb.UnimplementedUserServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetUserById(ctx context.Context, request *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) GetUserByLoginId(ctx context.Context, request *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) mustEmbedUnimplementedUserServiceServer() {
	//TODO implement me
	panic("implement me")
}
