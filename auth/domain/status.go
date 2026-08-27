package domain

import "fmt"

type Status int8

const (
	StatusDisabled Status = 0
	StatusEnabled  Status = 1
)

func (s Status) IsEnabled() bool {
	return s == StatusEnabled
}

func (s Status) Validate() error {
	if s != StatusDisabled && s != StatusEnabled {
		return fmt.Errorf("invalid status: %d, must be 0 or 1", s)
	}
	return nil
}

func (s Status) String() string {
	switch s {
	case StatusEnabled:
		return "enabled"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}
