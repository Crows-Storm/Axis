package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type RoleUpdatedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	RoleName    string
	Description string
}

func NewRoleUpdatedMessage(aggregateID uint64, name, description string) *RoleUpdatedMessage {
	return &RoleUpdatedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Updated,
		AggregateID: aggregateID,
		RoleName:    name,
		Description: description,
	}
}

func (RoleUpdatedMessage) Name() string {
	return "RoleUpdated"
}

func (RoleUpdatedMessage) Domain() string {
	return "role"
}

type RoleStatusChangedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	Status      int8
}

func NewRoleStatusChangedMessage(aggregateID uint64, status int8) *RoleStatusChangedMessage {
	return &RoleStatusChangedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Updated,
		AggregateID: aggregateID,
		Status:      status,
	}
}

func (RoleStatusChangedMessage) Name() string {
	return "RoleStatusChanged"
}

func (RoleStatusChangedMessage) Domain() string {
	return "role"
}
