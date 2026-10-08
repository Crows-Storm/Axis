package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type RoleCreatedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	Code        string
	RoleName    string
	Description string
}

// NewRoleCreatedMessage builds the event raised when a Role aggregate is created.
func NewRoleCreatedMessage(aggregateID uint64, code, name, description string) *RoleCreatedMessage {
	return &RoleCreatedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Created,
		AggregateID: aggregateID,
		Code:        code,
		RoleName:    name,
		Description: description,
	}
}

func (RoleCreatedMessage) Name() string {
	return "RoleCreated"
}

func (RoleCreatedMessage) Domain() string {
	return "role"
}
