package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
		LoginID:  request.LoginID,
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
		ID: request.Id,
	})
	if err != nil {
		return nil, err
	}
	// get current Principal context
	//p := grpcx.PrincipalFromContext(ctx)
	//fmt.Print(p)

	return &userpb.GetUserByIdResponse{
		Id:      result.ID,
		LoginID: result.LoginID,
		Email:   result.Email,
		Status:  int32(result.Status),
	}, nil
}

func (G GRPCServer) GetUserByLoginID(ctx context.Context, request *userpb.GetUserByLoginIDRequest) (*userpb.GetUserByLoginIDResponse, error) {
	result, err := G.app.Queries.GetUser.Handle(ctx, query.GetUserQuery{
		ID:      0,
		LoginID: request.LoginID,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserByLoginIDResponse{
		Id:      result.ID,
		LoginID: result.LoginID,
		Email:   result.Email,
		Status:  int32(result.Status),
	}, nil
}

func (G GRPCServer) CreateAndBindIdentity(ctx context.Context, request *userpb.CreateAndBindIdentityRequest) (*userpb.CreateAndBindIdentityResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) VerifyPassword(ctx context.Context, request *userpb.VerifyPasswordRequest) (*wrapperspb.BoolValue, error) {
	ok, err := G.app.Queries.VerifyLogin.Handle(ctx, query.VerifyLogin{
		LoginID:   request.LoginID,
		Password:  request.Password,
		RequestId: grpcx.RequestIDFromContext(ctx),
	})
	if err != nil {
		return &wrapperspb.BoolValue{Value: ok}, err
	}
	return &wrapperspb.BoolValue{Value: ok}, nil
}
