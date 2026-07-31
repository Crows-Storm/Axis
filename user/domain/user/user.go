package user

import "time"

type User struct {
	Id       int64  `json:"id"`
	LoginId  string `json:"login_id"`
	Password string `json:"-"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`
	Deleted  int8   `json:"-"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}
