package command

import (
	"context"
	"errors"
	"strings"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

type RecordAuditEventCommand struct {
	EventID     string         `json:"event_id"`
	EventName   string         `json:"event_name"`
	AggregateID string         `json:"aggregate_id"`
	OccurredAt  string         `json:"occurred_at"`
	Version     int            `json:"version"`
	Payload     []byte         `json:"payload"`
	Metadata    event.Metadata `json:"metadata,omitempty"`
	Headers     map[string]string
	OperatorID  uint64 `json:"operator_id"`
	Topic       string `json:"topic"`
	Partition   int    `json:"partition"`
	Offset      int64  `json:"offset"`
}

func (c RecordAuditEventCommand) Validate() decorator.CommandError {
	if strings.TrimSpace(c.EventID) == "" || strings.TrimSpace(c.EventName) == "" {
		return decorator.InvalidCommand("event_id and event_name are required")
	}
	return nil
}

type RecordAuditEventCommandHandler decorator.CommandHandler[RecordAuditEventCommand, *audit.AuditEvent]

type recordEventCommandHandler struct {
	repo audit.Repository
}

func NewRecordAuditEventCommandHandler(repo audit.Repository, metricsClient decorator.MetricsClient) RecordAuditEventCommandHandler {
	if repo == nil {
		panic("nil audit repository")
	}
	return decorator.ApplyCommandDecorators[RecordAuditEventCommand, *audit.AuditEvent](
		recordEventCommandHandler{repo: repo},
		metricsClient,
	)
}

func (h recordEventCommandHandler) Handle(ctx context.Context, cmd RecordAuditEventCommand) (*audit.AuditEvent, error) {
	rec, err := audit.New(event.Wire{
		EventID:     cmd.EventID,
		EventName:   cmd.EventName,
		AggregateID: cmd.AggregateID,
		OccurredAt:  cmd.OccurredAt,
		Version:     cmd.Version,
		Payload:     cmd.Payload,
		Metadata:    cmd.Metadata,
		Headers:     cmd.Headers,
		Topic:       cmd.Topic,
		Partition:   cmd.Partition,
		Offset:      cmd.Offset,
	}, cmd.OperatorID)
	if err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}

	existing, err := h.repo.GetByEventID(ctx, rec.EventID)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, audit.ErrNotFound) {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.repo.Append(ctx, rec); err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	return rec, nil
}
