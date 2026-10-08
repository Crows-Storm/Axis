package role

type Status int8

const (
	Disabled Status = 0
	Active   Status = 1
)

// Value returns the persisted tinyint representation of Status.
func (s *Status) Value() int8 {
	return int8(*s)
}

// IsActive reports whether the role status is Active (1).
func (s Status) IsActive() bool {
	return s == Active
}
