// Package user provides the user domain model and operations.
// It implements Domain-Driven Design patterns including:
// - Aggregate roots
// - Domain events
// - Repository pattern
package user

import (
	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/domain/event/events/user"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
)

const DomainName = "User"

type User struct {
	// Aggregate Root
	domain.AggregateRoot `json:"aggregate_root"`

	ID       int64           `json:"id"` // aggregate id
	LoginId  string          `json:"login_id"`
	Email    string          `json:"email"`
	Status   commuser.Status `json:"status"` // user the status
	Password string          `json:"-"`
	//Profile  Profile `json:"profile"`
}

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

func (u *User) AggregateName() string {
	return DomainName
}

// ApplyEvent implements the domain.Aggregate
// here apply event to the domain, setting all domain field and event recap to the apply
func (u *User) ApplyEvent(event event.DomainEvent) error {
	switch e := event.(type) {
	case *user.UserCreatedPayload:
		u.ID = e.AggregateId()
		u.LoginId = e.LoginId
		u.Email = e.Email
		u.Status = 1

	case *user.UserUpdatedPayload:

		u.Email = e.Email

	case *user.UserApplyPasswordPayload:
		// nothing
	}
	return nil
}
