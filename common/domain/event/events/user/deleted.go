package user

import "github.com/Crows-Storm/Axis/common/domain/event"

type UserDeletedPayload struct {
	event.BaseEvent

	LoginId string `json:"login_id"`
}

func NewUserDeletedEvent(
	aggregateID int64,
	aggregateName string,
	payload UserDeletedPayload,
) (event.DomainEvent, error) {
	return event.NewEventBuilder().
		WithName("user.deleted").
		WithEventCategory(event.Deleted). // change/create password is a mark category
		WithAggregate(aggregateID, aggregateName).
		WithPayload(payload).
		Build()
}
