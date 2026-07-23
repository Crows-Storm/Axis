package client

import (
	"context"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
)

func NewGRPCClient(ctx context.Context) (userpb.UserServiceClient, close func() error, err error) {
	
}
