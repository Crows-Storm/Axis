package protos

import (
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/common/genproto/authpb"
)

type GRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}
