package user

import (
	"time"
)

type User struct {
	Id       int64  `json:"id"`
	LoginId  string `json:"loginId"`
	Password string `json:"-"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`
	Deleted  int8   `json:"deleted"`

	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}

func (u *User) Create() {
	u.CreateTime = time.Now()
	u.UpdateTime = time.Now()
}
