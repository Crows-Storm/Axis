package client

import (
	"context"

	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
)

// NewUserGRPCClient Obtain the User service client through Consul service discovery.
func NewUserGRPCClient(ctx context.Context) (userpb.UserServiceClient, func() error, error) {
	conn, err := grpcx.DialService("user")
	if err != nil {
		return nil, nil, err
	}

	client := userpb.NewUserServiceClient(conn)
	return client, conn.Close, nil
}
