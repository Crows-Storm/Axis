package user

type Status int8

// Status constants
const (
	Activated Status = iota
	Disabled
)

func (s *Status) Value() int8 {
	return int8(*s)
}

type Delete int8

// Delete constants
const (
	Undeleted Delete = iota
	Deleted
)

func (s *Delete) Value() int8 {
	return int8(*s)
}
