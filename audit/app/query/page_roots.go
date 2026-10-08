package query

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/pagination"
)

type PageRootsQuery struct {
	Args      pagination.Args
	Service   string
	EventName string
	From      *time.Time
	To        *time.Time
}

type PageRootsQueryHandler decorator.QueryHandler[PageRootsQuery, *pagination.Connection[*audit.AuditEvent]]

type pageRootsQueryHandler struct {
	repo audit.Repository
}

func NewPageRootsQueryHandler(repo audit.Repository, metricsClient decorator.MetricsClient) PageRootsQueryHandler {
	if repo == nil {
		panic("nil audit repository")
	}
	return decorator.ApplyQueryDecorators[PageRootsQuery, *pagination.Connection[*audit.AuditEvent]](
		pageRootsQueryHandler{repo: repo},
		metricsClient,
	)
}

func (h pageRootsQueryHandler) Handle(ctx context.Context, q PageRootsQuery) (*pagination.Connection[*audit.AuditEvent], error) {
	filter := audit.RootFilter{Service: q.Service, EventName: q.EventName}
	if q.From != nil || q.To != nil {
		if q.From == nil || q.To == nil || q.From.IsZero() || q.To.IsZero() || !q.To.After(*q.From) {
			return nil, decorator.CommandExecutedError{Msg: "from and to are required together and to must be after from"}
		}
		filter.From = q.From.UTC()
		filter.To = q.To.UTC()
	}
	conn, err := h.repo.PageRoots(ctx, filter, q.Args)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return pagination.Wrap([]*audit.AuditEvent{}, &pagination.Resolved{Direction: pagination.Forward, Limit: pagination.DefaultLimit}, pagination.NewestFirst("timestamp", "event_id"), 0), nil
	}
	return conn, nil
}
