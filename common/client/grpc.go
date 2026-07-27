package client

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUserGRPCClient(ctx context.Context) (userpb.UserServiceClient, func() error, error) {
	grpcAddress := viper.GetString("user.grpc-addr")
	opts, err := grpcDialOpts(grpcAddress)
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(grpcAddress, opts...)
	if err != nil {
		return nil, nil, err
	}

	return userpb.NewUserServiceClient(conn), conn.Close, err
}

func grpcDialOpts(address string) ([]grpc.DialOption, error) {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}
