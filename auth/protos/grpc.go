package protos

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/query"
	"github.com/Crows-Storm/Axis/common/genproto/authpb"
)

type GRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G *GRPCServer) CheckPermission(ctx context.Context, req *authpb.CheckPermissionRequest) (*authpb.CheckPermissionResponse, error) {
	if req == nil {
		return &authpb.CheckPermissionResponse{}, nil
	}
	result, err := G.app.Queries.CheckMenuPermission.Handle(ctx, query.CheckMenuPermissionQuery{
		UserID:   req.GetUserId(),
		MenuCode: req.GetMenu(),
		Action:   req.GetAction(),
	})
	if err != nil {
		return nil, err
	}
	return &authpb.CheckPermissionResponse{
		Allowed: result.Allowed,
		Mask:    result.Mask,
		Actions: result.Actions,
	}, nil
}
