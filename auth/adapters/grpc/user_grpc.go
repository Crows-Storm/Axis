package grpc

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGRPC struct {
	client userpb.UserServiceClient
}

func NewUserGRPC(client userpb.UserServiceClient) *UserGRPC {
	return &UserGRPC{client: client}
}

func (u UserGRPC) CreateUser(ctx context.Context, request *userpb.CreateUserRequest) (*emptypb.Empty, error) {
	result, err := u.client.CreateUser(ctx, request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (u UserGRPC) GetUserById(ctx context.Context, request *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	result, err := u.client.GetUserById(ctx, request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (u UserGRPC) GetUserByLoginId(ctx context.Context, request *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error) {
	result, err := u.client.GetUserByLoginId(ctx, request)
	if err != nil {
		return nil, err
	}
	return result, nil
}
