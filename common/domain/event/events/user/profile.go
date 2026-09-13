package user

import (
	"fmt"

	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

type UserUpdatedPayload struct {
	event.BaseEvent

	Email   string               `json:"email"`
	Changes []domain.FieldChange `json:"changes"`
}

func NewUserUpdatedEvent(
	aggregateID int64,
	aggregateName string,
	payload UserUpdatedPayload,
) (event.DomainEvent, error) {
	if len(payload.Changes) == 0 {
		return nil, fmt.Errorf("no changes to record")
	}

	return event.NewEventBuilder().
		WithName("user.updated").
		WithEventCategory(event.Updated).
		WithAggregate(aggregateID, aggregateName).
		WithPayload(payload).
		Build()
}
