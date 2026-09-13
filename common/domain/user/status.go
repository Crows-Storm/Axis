package user

type Status int8

// Status constants
const (
	Pending Status = iota
	Activated
	Disabled
)

func (s *Status) Value() int8 {
	return int8(*s)
}
