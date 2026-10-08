package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type UserApplyPasswordMessage struct {
	ID       string // event id
	Category event.EventCategory

	AggregateID uint64
	LoginID     string
}

func NewUserApplyPasswordMessage(aggregateID uint64, loginID string) *UserApplyPasswordMessage {
	return &UserApplyPasswordMessage{
		ID:          event.GenerateEventID(),
		AggregateID: aggregateID,
		LoginID:     loginID,
		Category:    event.Created,
	}
}

func (u UserApplyPasswordMessage) Name() string {
	return "UserApplyPassword"
}

func (u UserApplyPasswordMessage) Domain() string {
	return "user"
}
