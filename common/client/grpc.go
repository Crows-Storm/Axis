package client

import (
	"context"

	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/genproto/accountpb"
	"github.com/Crows-Storm/Axis/common/genproto/authpb"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
)

// NewUserGRPCClient Obtain the User service client through Consul service discovery.
func NewUserGRPCClient(ctx context.Context) (userpb.UserServiceClient, func() error, error) {
	conn, err := grpcx.DialService("user")
	if err != nil {
		return nil, nil, err
	}

	client := userpb.NewUserServiceClient(conn)
	// The connection stays in the grpcx pool. CloseAll owns its lifetime.
	return client, func() error { return nil }, nil
}

func NewAuthGRPCClient(_ context.Context) (authpb.AuthServiceClient, func() error, error) {
	conn, err := grpcx.DialService("auth")
	if err != nil {
		return nil, nil, err
	}
	return authpb.NewAuthServiceClient(conn), func() error { return nil }, nil
}

func NewAccountGRPCClient(_ context.Context) (accountpb.AccountServiceClient, func() error, error) {
	conn, err := grpcx.DialService("account")
	if err != nil {
		return nil, nil, err
	}
	return accountpb.NewAccountServiceClient(conn), func() error { return nil }, nil
}
