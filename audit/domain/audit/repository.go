package audit

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/pagination"
)

type RootFilter struct {
	Service   string
	EventName string
	From      time.Time
	To        time.Time
}

type Repository interface {
	Append(ctx context.Context, rec *AuditEvent) error
	GetByEventID(ctx context.Context, eventID string) (*AuditEvent, error)
	PageRoots(ctx context.Context, filter RootFilter, args pagination.Args) (*pagination.Connection[*AuditEvent], error)
	ListByCorrelation(ctx context.Context, correlationID string) ([]*AuditEvent, error)
}
