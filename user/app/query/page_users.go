package query

import (
	"context"
	"strings"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/pagination"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type PageUsersQuery struct {
	Args    pagination.Args
	LoginID string
	Email   string
	Status  *int8
}

type PageUsersQueryHandler decorator.QueryHandler[PageUsersQuery, *pagination.Connection[*domain.User]]

type pageUsersQueryHandler struct {
	repo domain.Repository
}

func NewPageUsersQueryHandler(repo domain.Repository, metricsClient decorator.MetricsClient) PageUsersQueryHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[PageUsersQuery, *pagination.Connection[*domain.User]](
		pageUsersQueryHandler{repo: repo},
		metricsClient,
	)
}

func (h pageUsersQueryHandler) Handle(ctx context.Context, q PageUsersQuery) (*pagination.Connection[*domain.User], error) {
	if q.Status != nil && *q.Status != 0 && *q.Status != 1 {
		return nil, decorator.CommandExecutedError{Msg: "status must be 0 or 1"}
	}
	conn, err := h.repo.Page(ctx, domain.PageFilter{
		LoginID: strings.TrimSpace(q.LoginID),
		Email:   strings.TrimSpace(q.Email),
		Status:  q.Status,
	}, q.Args)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return &pagination.Connection[*domain.User]{
			Edges:    []*pagination.Edge[*domain.User]{},
			PageInfo: &pagination.PageInfo{},
		}, nil
	}
	return conn, nil
}
