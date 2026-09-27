package events

import (
	"github.com/Crows-Storm/Axis/common/domain/event"
	domain "github.com/Crows-Storm/Axis/common/domain/user"
)

type UserCreatedMessage struct {
	ID       string // event id
	Category event.EventCategory

	AggregateID uint64
	LoginID     string
	Email       string
}

func NewUserCreatedMessage(aggregateID uint64, loginID, email string) *UserCreatedMessage {
	return &UserCreatedMessage{
		ID:          event.GenerateEventID(),
		AggregateID: aggregateID,
		LoginID:     loginID,
		Email:       email,
		Category:    event.Created,
	}
}

func (u UserCreatedMessage) Name() string {
	return "UserCreated"
}

type UserStatusMessage struct {
	ID       string // event id
	Category event.EventCategory

	AggregateID uint64
	Status      domain.Status
}

func NewUserStatusMessage(aggregateID uint64, status domain.Status) *UserCreatedMessage {
	// what category by domain status
	var category = event.Disabled
	if status != domain.Disabled {
		category = event.Activated
	}
	return &UserCreatedMessage{
		ID:          event.GenerateEventID(),
		AggregateID: aggregateID,
		Category:    category,
	}
}

func (u UserStatusMessage) Name() string {
	return "UserStatusChange"
}
