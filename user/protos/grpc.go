package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	userpb.UnimplementedUserServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CreateUser(ctx context.Context, request *userpb.CreateUserRequest) (*emptypb.Empty, error) {
	_, err := G.app.Commands.CreateUser.Handle(ctx, command.CreateUserCommand{
		LoginId:  request.LoginId,
		Password: request.Password,
		Email:    request.Email,
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (G GRPCServer) GetUserById(ctx context.Context, request *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	result, err := G.app.Queries.GetUser.Handle(ctx, query.GetUserQuery{
		Id: request.Id,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserByIdResponse{
		Id:      result.Id,
		LoginId: result.LoginId,
		Email:   result.Email,
		Status:  int32(result.Status),
	}, nil
}

func (G GRPCServer) GetUserByLoginId(ctx context.Context, request *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error) {
	//TODO implement me
	panic("implement me")
}
