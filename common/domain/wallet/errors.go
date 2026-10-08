package wallet

import "errors"

var (
	ErrNotFound = errors.New("wallet not found")
	ErrInvalid  = errors.New("invalid wallet")
)
