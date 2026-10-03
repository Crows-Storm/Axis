package user

import (
	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
)

type User struct {
	domain.AggregateRoot

	ID       uint64 `json:"id"`
	Username string `json:"username"`
	LoginID  string `json:"login_id"`
	Password string `json:"-"`
	Email    string `json:"email"`
	//Profile  Profile `json:"profile"`
	Status  commuser.Status `json:"status"`
	Deleted int8            `json:"deleted"`
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}

// TODO: nned design
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

func (u *User) ApplyEvent(evt event.Message) error {
	switch e := evt.(type) {
	case *events.UserCreatedMessage:
		u.ID = e.AggregateID
		u.Username = e.Username
		u.LoginID = e.LoginID
		u.Email = e.Email
		u.Status = commuser.Activated // User Created Message default status is Activated
		u.Deleted = 0
	case *events.UserStatusMessage:
		u.ID = e.AggregateID
		u.Status = e.Status
	}
	return nil
}
