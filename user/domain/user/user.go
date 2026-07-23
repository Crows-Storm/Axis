package user

import "time"

type User struct {
	Id       int64
	LoginId  string
	Password string
	Email    string
	Status   int8
	Deleted  int8

	CreateTime time.Time
	UpdateTime time.Time
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}
