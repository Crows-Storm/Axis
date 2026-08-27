package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
)

type GetUserQuery struct {
	Id int64
}

func (g GetUserQuery) Validate() error {
	if g.Id <= 0 {
		return errors.New("invalid id")
	}
	return nil
}

type GetUserQueryResult struct {
	Id      int64
	LoginId string
	Email   string
	Status  int32
}

type GetUserQueryHandler decorator.QueryHandler[GetUserQuery, GetUserQueryResult]

type getUserQueryHandler struct {
	userService provider.UserService
}

func NewGetUserQueryHandler(
	userService provider.UserService,
	metricsClient decorator.MetricsClient,
) GetUserQueryHandler {
	if userService == nil {
		panic("nil User GRPC Service")
	}
	return decorator.ApplyQueryDecorators[GetUserQuery, GetUserQueryResult](
		getUserQueryHandler{userService: userService},
		metricsClient,
	)
}

func (g getUserQueryHandler) Handle(ctx context.Context, query GetUserQuery) (GetUserQueryResult, error) {
	if err := query.Validate(); err != nil {
		return GetUserQueryResult{}, err
	}

	u, err := g.userService.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: query.Id,
	})

	if err != nil {
		return GetUserQueryResult{}, err
	}
	return GetUserQueryResult{
		Id:      u.Id,
		LoginId: u.LoginId,
		Email:   u.Email,
		Status:  u.Status,
	}, err
}
