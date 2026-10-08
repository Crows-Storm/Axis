package events

import (
	"github.com/Crows-Storm/Axis/common/domain/event"
	domain "github.com/Crows-Storm/Axis/common/domain/user"
)

type UserStatusMessage struct {
	ID       string
	Category event.EventCategory

	AggregateID uint64
	Status      domain.Status
}

func NewUserStatusMessage(aggregateID uint64, status domain.Status) *UserStatusMessage {
	category := event.Disabled
	if status != domain.Disabled {
		category = event.Activated
	}
	return &UserStatusMessage{
		ID:          event.GenerateEventID(),
		AggregateID: aggregateID,
		Status:      status,
		Category:    category,
	}
}

func (u UserStatusMessage) Name() string {
	return "UserStatusChange"
}

func (u UserStatusMessage) Domain() string {
	return "user"
}
