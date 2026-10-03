package events

import (
	"github.com/Crows-Storm/Axis/common/domain/event"
)

type UserCreatedMessage struct {
	ID       string // event id
	Category event.EventCategory

	AggregateID uint64
	LoginID     string
	Username    string
	Email       string
}

func NewUserCreatedMessage(aggregateID uint64, loginID, email, username string) *UserCreatedMessage {
	return &UserCreatedMessage{
		ID:          event.GenerateEventID(),
		AggregateID: aggregateID,
		LoginID:     loginID,
		Username:    username,
		Email:       email,
		Category:    event.Created,
	}
}

func (u UserCreatedMessage) Name() string {
	return "UserCreated"
}
