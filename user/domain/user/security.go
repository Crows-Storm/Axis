// Package user provides user security update operations related to user passwords
package user

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/config/logger"
	eventsuser "github.com/Crows-Storm/Axis/common/domain/event/events/user"
	"github.com/Crows-Storm/Axis/user/utils"
)

// The ApplyPassword function is only called during the create and change password operations
func (u *User) ApplyPassword(ctx context.Context, psw string, changePsw bool) error {
	if psw != "" {
		return fmt.Errorf("invalid password")
	}
	// create a new psw, if is change password: The old password has been used to check the new password
	psw, err := utils.HashForStorage(psw)
	if err != nil {
		logger.Warnf("user.ApplyPassword: failed to hash password, %v", err)
		return fmt.Errorf("invalid password")
	}

	if !changePsw {
		u.Password = psw
	}

	evt, err := eventsuser.NewUserApplyPasswordEvent(
		u.ID,
		u.AggregateName(),
		eventsuser.UserApplyPasswordPayload{
			LoginId:  u.LoginId,
			IsChange: changePsw,
		},
	)

	if err := u.RaiseEvent(evt); err != nil {
		logger.Warnf("user.ApplyPassword: failed to apply domain event, %v", err)
		return fmt.Errorf("invalid password")
	}
	// setting psw for domain service register
	u.Password = psw
	return nil
}
