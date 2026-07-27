package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService interface {
	CreateUser(ctx context.Context, request *userpb.CreateUserRequest) (*emptypb.Empty, error)

	GetUserById(ctx context.Context, request *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error)
	GetUserByLoginId(ctx context.Context, request *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error)
}
