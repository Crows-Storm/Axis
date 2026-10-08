package rbac

import "errors"

var (
	ErrForbidden       = errors.New("forbidden")
	ErrUnavailable     = errors.New("permission store unavailable")
	ErrUnauthenticated = errors.New("unauthenticated")
)
