package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type RoleMenuGrantedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	MenuID      uint64
	Mask        uint64
}

// NewRoleMenuGrantedMessage builds the event raised when bits are OR-merged onto a menu grant.
func NewRoleMenuGrantedMessage(aggregateID, menuID, mask uint64) *RoleMenuGrantedMessage {
	return &RoleMenuGrantedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Updated,
		AggregateID: aggregateID,
		MenuID:      menuID,
		Mask:        mask,
	}
}

func (RoleMenuGrantedMessage) Name() string {
	return "RoleMenuGranted"
}

func (RoleMenuGrantedMessage) Domain() string {
	return "role"
}

type RoleMenuMaskReplacedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	MenuID      uint64
	Mask        uint64
}

// NewRoleMenuMaskReplacedMessage builds the event raised when a menu grant is replaced.
func NewRoleMenuMaskReplacedMessage(aggregateID, menuID, mask uint64) *RoleMenuMaskReplacedMessage {
	return &RoleMenuMaskReplacedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Updated,
		AggregateID: aggregateID,
		MenuID:      menuID,
		Mask:        mask,
	}
}

func (RoleMenuMaskReplacedMessage) Name() string {
	return "RoleMenuMaskReplaced"
}

func (RoleMenuMaskReplacedMessage) Domain() string {
	return "role"
}

type RoleMenuRevokedMessage struct {
	ID          string
	Category    event.EventCategory
	AggregateID uint64
	MenuID      uint64
}

func NewRoleMenuRevokedMessage(aggregateID, menuID uint64) *RoleMenuRevokedMessage {
	return &RoleMenuRevokedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Deleted,
		AggregateID: aggregateID,
		MenuID:      menuID,
	}
}

func (RoleMenuRevokedMessage) Name() string {
	return "RoleMenuRevoked"
}

func (RoleMenuRevokedMessage) Domain() string {
	return "role"
}
