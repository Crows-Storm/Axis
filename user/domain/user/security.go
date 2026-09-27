// Package user provides user security update operations related to user passwords
package user

import (
	"fmt"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
)

// The ApplyPassword function is only called during the create and change password operations
func (u *User) ApplyPassword(psw string, isChange bool, hasher PasswordHasher) error {
	if psw == "" {
		return user.ErrInvalidPassword
	}
	// create a new psw, if is change password: The old password has been used to check the new password
	storePsw, err := hasher.Hash(psw)
	if err != nil {
		logger.Warnf("user.ApplyPassword: failed to hash password, %v", err)
		return user.ErrInvalidPassword
	}

	u.Password = storePsw

	message := events.NewUserApplyPasswordMessage(
		u.ID,
		u.LoginID,
	)

	if err := u.RaiseEvent(u, message); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("user.Create: failed to apply domain event, %w", err)
	}
	return nil
}
