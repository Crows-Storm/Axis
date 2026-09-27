package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/server"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type VerifyLogin struct {
	LoginID   string `json:"LoginID"`
	Password  string `json:"password"`
	RequestId string `json:"requestId"`
}

type VerifyLoginHandler decorator.QueryHandler[VerifyLogin, bool]

type verifyLoginHandler struct {
	userRepo domain.Repository
	hasher   domain.PasswordHasher
}

func NewVerifyLoginHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
	hasher domain.PasswordHasher,
) VerifyLoginHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[VerifyLogin, bool](
		verifyLoginHandler{userRepo: repo, hasher: hasher},
		metricsClient,
	)

}

func (g verifyLoginHandler) Handle(ctx context.Context, query VerifyLogin) (bool, error) {
	// request ID restored from gRPC metadata by inboundMetadataInterceptor
	if query.RequestId == "" {
		// bad request
		return false, errors.New(server.CodeBadRequest.String())
	}
	psw := g.userRepo.GetPasswordByLoginID(ctx, query.LoginID)

	eq, err := g.hasher.Compare(psw, query.Password, query.RequestId)
	// 09e006642e7c6e2d3fa14f3f6e4b2f815a7d02cd8d0669c3bb1f78748bdd1999
	if err != nil {
		return false, err
	}
	return eq, nil
}
