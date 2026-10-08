package service

import (
	"context"

	"github.com/Crows-Storm/Axis/audit/app"
	"github.com/Crows-Storm/Axis/audit/app/command"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/server/kafka"
)

// ConsumeRecord stores one audit-topic delivery. The caller has no principal;
// operator id stays on the payload.
func ConsumeRecord(ctx context.Context, application app.Application, delivery kafka.Delivery) error {
	wire, err := decorator.DecodeWire(delivery.Topic, delivery.Partition, delivery.Offset, delivery.Value, delivery.Headers)
	if err != nil {
		return err
	}
	ctx = contextx.BindMQ(ctx, delivery.Headers, "", "")
	_, err = application.Commands.RecordEvent.Handle(ctx, command.RecordAuditEventCommand{
		EventID:     wire.EventID,
		EventName:   wire.EventName,
		AggregateID: wire.AggregateID,
		OccurredAt:  wire.OccurredAt,
		Version:     wire.Version,
		Payload:     wire.Payload,
		Metadata:    wire.Metadata,
		Headers:     wire.Headers,
		Topic:       wire.Topic,
		Partition:   wire.Partition,
		Offset:      wire.Offset,
	})
	return err
}
