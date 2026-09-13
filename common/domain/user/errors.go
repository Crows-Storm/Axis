package user

import "errors"

var (
	ErrInvalidUser               = errors.New("invalid user")
	ErrInvalidEmail              = errors.New("invalid email format")
	ErrUsernameTooShort          = errors.New("username must be at least 3 characters")
	ErrPasswordTooWeak           = errors.New("password must be at least 8 characters")
	ErrAlreadyActive             = errors.New("user already active")
	ErrBlockedUserCannotActivate = errors.New("blocked user cannot be activated")
	ErrCannotModifySuspendedUser = errors.New("cannot modify suspended user")
	ErrWrongPassword             = errors.New("wrong password")
	ErrEmailSame                 = errors.New("new email is same as old")
	ErrNotSuspended              = errors.New("user is not suspended")
	ErrBlockedUserCannotSuspend  = errors.New("blocked user cannot be suspended")
	ErrUserAlreadyExists         = errors.New("user already exists")
)
