package user

import (
	"fmt"

	eventsuser "github.com/Crows-Storm/Axis/common/domain/event/events/user"
)

func (u *User) Delete() error {
	evt, err := eventsuser.NewUserDeletedEvent(
		u.ID,
		u.AggregateName(),
		eventsuser.UserDeletedPayload{
			LoginId: u.LoginId,
		},
	)
	if err != nil {
		return fmt.Errorf("user.Delete: failed to apply domain event, %w", err)
	}

	if err := u.RaiseEvent(evt); err != nil {
		return fmt.Errorf("user.Delete: failed to apply domain event, %w", err)
	}

	return nil
}
