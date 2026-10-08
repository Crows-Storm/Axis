package query

import (
	"context"
	"strings"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type GetEventGraphQuery struct {
	EventID string `json:"event_id"`
}

type GetEventGraphQueryHandler decorator.QueryHandler[GetEventGraphQuery, *audit.EventGraph]

type getEventGraphQueryHandler struct {
	repo audit.Repository
}

func NewGetEventGraphQueryHandler(repo audit.Repository, metricsClient decorator.MetricsClient) GetEventGraphQueryHandler {
	if repo == nil {
		panic("nil audit repository")
	}
	return decorator.ApplyQueryDecorators[GetEventGraphQuery, *audit.EventGraph](
		getEventGraphQueryHandler{repo: repo},
		metricsClient,
	)
}

func (h getEventGraphQueryHandler) Handle(ctx context.Context, q GetEventGraphQuery) (*audit.EventGraph, error) {
	eventID := strings.TrimSpace(q.EventID)
	if eventID == "" {
		return nil, decorator.CommandExecutedError{Msg: "event_id is required"}
	}
	focus, err := h.repo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	rows, err := h.repo.ListByCorrelation(ctx, focus.Metadata.CorrelationId)
	if err != nil {
		return nil, err
	}
	return audit.Forest(focus, rows), nil
}
