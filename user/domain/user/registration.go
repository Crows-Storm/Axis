// Package user Provides user creation and activation operations. e.g. NewUser(), Activate(), ValidatePassword()
package user

import (
	"context"
	"fmt"

	eventsuser "github.com/Crows-Storm/Axis/common/domain/event/events/user"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/util"
)

// The Create is factory function
// create user and raise user domain event
func Create(ctx context.Context, loginId, email string) (*User, error) {
	// aggregate root data
	aggregateId := util.GenerateFlakeID()

	// init domain
	var u User

	evt, err := eventsuser.NewUserCreatedEvent(
		aggregateId,
		fmt.Sprintf("%s", u.AggregateName()),
		eventsuser.UserCreatedPayload{
			LoginId: loginId,
			Email:   email,
		},
	)
	if err != nil {
		return nil, err
	}

	// raise user domain event
	if err := u.RaiseEvent(evt); err != nil {
		return nil, fmt.Errorf("user.Create: failed to apply domain event, %w", err)
	}
	return &u, nil
}

func (u *User) Disable(ctx context.Context) error {
	evt, err := eventsuser.NewUserChangeStatusEvent(
		u.ID,
		fmt.Sprintf("%s", u.AggregateName()),
		eventsuser.UserChangeStatusPayload{
			LoginId:   u.LoginId,
			OldStatus: u.Status,
			NewStatus: commuser.Disabled,
		},
	)
	if err != nil {
		return err
	}

	// raise user domain event
	if err := u.RaiseEvent(evt); err != nil {
		return fmt.Errorf("user.Disable: failed to apply domain event, %w", err)
	}
	return nil
}

func (u *User) Activation(ctx context.Context) error {
	evt, err := eventsuser.NewUserChangeStatusEvent(
		u.ID,
		fmt.Sprintf("%s", u.AggregateName()),
		eventsuser.UserChangeStatusPayload{
			LoginId:   u.LoginId,
			OldStatus: u.Status,
			NewStatus: commuser.Activated,
		},
	)
	if err != nil {
		return err
	}

	// raise user domain event
	if err := u.RaiseEvent(evt); err != nil {
		return fmt.Errorf("user.Activation: failed to apply domain event, %w", err)
	}
	return nil
}
