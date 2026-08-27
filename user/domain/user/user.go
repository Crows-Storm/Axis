package user

import (
	"time"
)

type User struct {
	Id       int64  `json:"id"`
	LoginId  string `json:"loginId"`
	Password string `json:"-"`
	Email    string `json:"email"`
	//Profile  Profile `json:"profile"`
	Status  int8 `json:"status"`
	Deleted int8 `json:"deleted"`

	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}

// Profile is User basic profit
type Profile struct {
	Language     int8  // CN: 1  EN: 2
	CurrencyId   int64 // currency id
	CurrencyName string
}

// DefaultProfile use English and USD
func DefaultProfile() Profile {
	return Profile{
		Language:     2,
		CurrencyId:   1,
		CurrencyName: "USD",
	}
}

func (u *User) Create() {
	u.CreateTime = time.Now()
	u.UpdateTime = time.Now()
}
