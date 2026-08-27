package provider

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type UserService interface {
	CreateUser(ctx context.Context, request *userpb.CreateUserRequest) (*emptypb.Empty, error)

	GetUserById(ctx context.Context, request *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error)
	GetUserByLoginId(ctx context.Context, request *userpb.GetUserByLoginIdRequest) (*userpb.GetUserByLoginIdResponse, error)

	GetByUnionID(ctx context.Context, request *userpb.GetByUnionIDRequest) (*userpb.GetByUnionIDResponse, error)
	// CreateAndBindIdentity Create a new user and bind their identity (register)
	CreateAndBindIdentity(ctx context.Context, request *userpb.CreateAndBindIdentityRequest) (*userpb.CreateAndBindIdentityResponse, error)
	VerifyPassword(ctx context.Context, request *userpb.VerifyPasswordRequest) (*wrapperspb.BoolValue, error)
}
