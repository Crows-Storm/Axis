package grpc

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type UserGRPC struct {
	client userpb.UserServiceClient
}

func NewUserGRPC(client userpb.UserServiceClient) *UserGRPC {
	return &UserGRPC{client: client}
}

func (u *UserGRPC) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*emptypb.Empty, error) {
	return u.client.CreateUser(ctx, req)
}

func (u *UserGRPC) GetUserById(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	return u.client.GetUserById(ctx, req)
}

func (u *UserGRPC) GetUserByLoginId(ctx context.Context, req *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error) {
	return u.client.GetUserByLoginId(ctx, req)
}

func (u *UserGRPC) GetByUnionID(ctx context.Context, req *userpb.GetByUnionIDRequest) (*userpb.GetByUnionIDResponse, error) {
	return u.client.GetByUnionID(ctx, req)
}

func (u *UserGRPC) CreateAndBindIdentity(ctx context.Context, req *userpb.CreateAndBindIdentityRequest) (*userpb.CreateAndBindIdentityResponse, error) {
	return u.client.CreateAndBindIdentity(ctx, req)
}

func (u *UserGRPC) VerifyPassword(ctx context.Context, req *userpb.VerifyPasswordRequest) (*wrapperspb.BoolValue, error) {
	return u.client.VerifyPassword(ctx, req)
}
