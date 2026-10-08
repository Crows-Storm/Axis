package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type UserAssignedToRoleMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	UserID      uint64
}

// NewUserAssignedToRoleMessage builds the event raised when a user is bound to a role.
func NewUserAssignedToRoleMessage(aggregateID, userID uint64) *UserAssignedToRoleMessage {
	return &UserAssignedToRoleMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Updated,
		AggregateID: aggregateID,
		UserID:      userID,
	}
}

func (UserAssignedToRoleMessage) Name() string {
	return "UserAssignedToRole"
}

func (UserAssignedToRoleMessage) Domain() string {
	return "role"
}

type UserUnassignedFromRoleMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	UserID      uint64
}

func NewUserUnassignedFromRoleMessage(aggregateID, userID uint64) *UserUnassignedFromRoleMessage {
	return &UserUnassignedFromRoleMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Deleted,
		AggregateID: aggregateID,
		UserID:      userID,
	}
}

func (UserUnassignedFromRoleMessage) Name() string {
	return "UserUnassignedFromRole"
}

func (UserUnassignedFromRoleMessage) Domain() string {
	return "role"
}
