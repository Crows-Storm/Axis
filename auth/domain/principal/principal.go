package principal

import "time"

type Principal struct {
	Id      int64
	LoginId string
	Email   string
	Status  int8

	RoleId     int64
	Permission []*string // menu code

	LastLoginTime time.Time
}
