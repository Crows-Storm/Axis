package user

import "github.com/Crows-Storm/Axis/common/domain/event"

type UserApplyPasswordPayload struct {
	event.BaseEvent

	LoginId  string `json:"login_id"`
	IsChange bool   `json:"is_change"`

	// can not record anything about password !!! so this payload is a mark event
	//OldPassword string `json:"old_password,omitempty"`
	//NewPassword string `json:"new_password,omitempty"`
}

func NewUserApplyPasswordEvent(
	aggregateID int64,
	aggregateName string,
	payload UserApplyPasswordPayload,
) (event.DomainEvent, error) {
	return event.NewEventBuilder().
		WithName("user.apply_password").
		WithEventCategory(event.Mark). // change/create password is a mark category
		WithAggregate(aggregateID, aggregateName).
		WithPayload(payload).
		Build()
}
