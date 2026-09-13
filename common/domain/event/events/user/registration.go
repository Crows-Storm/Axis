package user

import (
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/domain/user"
)

type UserCreatedPayload struct {
	event.BaseEvent

	LoginId string `json:"login_id"`
	Email   string `json:"email"`
}

func NewUserCreatedEvent(
	aggregateID int64,
	aggregateName string,
	payload UserCreatedPayload,
) (event.DomainEvent, error) {

	return event.NewEventBuilder().
		WithName("user.created").
		WithEventCategory(event.Created).
		WithAggregate(aggregateID, aggregateName).
		WithPayload(payload).
		Build()
}

type UserChangeStatusPayload struct {
	event.BaseEvent

	LoginId   string      `json:"login_id"`
	OldStatus user.Status `json:"old_status"`
	NewStatus user.Status `json:"new_status"`
}

func NewUserChangeStatusEvent(
	aggregateID int64,
	aggregateName string,
	payload UserChangeStatusPayload,
) (event.DomainEvent, error) {
	return event.NewEventBuilder().
		WithName("user.activated").
		WithEventCategory(event.Activated).
		WithAggregate(aggregateID, aggregateName).
		WithPayload(payload).
		Build()
}
