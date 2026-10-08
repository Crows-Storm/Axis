package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type AccountOpenedMessage struct {
	ID       string
	Category event.EventCategory

	AggregateID uint64
	UserID      uint64
	AccountType string
}

func NewAccountOpenedMessage(aggregateID, userID uint64, accountType string) *AccountOpenedMessage {
	return &AccountOpenedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Created,
		AggregateID: aggregateID,
		UserID:      userID,
		AccountType: accountType,
	}
}

func (AccountOpenedMessage) Name() string   { return "AccountOpened" }
func (AccountOpenedMessage) Domain() string { return "account" }
