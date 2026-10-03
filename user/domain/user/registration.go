package user

import (
	"fmt"

	domain "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
	"github.com/Crows-Storm/Axis/common/util"
)

// The Create is factory function
// create user and raise user domain event
func Create(loginID, email, username string) (*User, error) {
	// aggregate root data
	aggregateId := util.GenerateFlakeID()

	// init domain
	u := &User{}

	// created a domain event and apply it
	if err := u.RaiseEvent(u, events.NewUserCreatedMessage(aggregateId, loginID, email, username)); err != nil { // setting event to root.events and increment version
		return nil, fmt.Errorf("user.Create: failed to apply domain event, %w", err)
	}
	return u, nil
}

// The Disable is factory function
func (u *User) Disable() error {
	message := events.NewUserStatusMessage(u.ID, domain.Disabled)
	if err := u.RaiseEvent(u, message); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("user.Disable: failed to apply domain event, %w", err)
	}
	return nil
}

func (u *User) Activation() error {
	message := events.NewUserStatusMessage(u.ID, domain.Activated)
	if err := u.RaiseEvent(u, message); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("user.Activation: failed to apply domain event, %w", err)
	}
	return nil
}
