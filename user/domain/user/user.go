package user

import (
	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
)

type User struct {
	domain.AggregateRoot

	ID       uint64            `json:"id"`
	Username string            `json:"username"`
	LoginID  string            `json:"login_id"`
	Password string            `json:"-"`
	Email    string            `json:"email"`
	Profile  map[Scope]Profile `json:"profile"` // is user the platform profile, default nil
	Status   commuser.Status   `json:"status"`
	Deleted  int8              `json:"deleted"`
}

// Desensitization password
//func (u *User) Desensitization() {
//	u.password = ""
//}

// The Scope is Profile scope
// SystemProfile for system of setting
// ExchangeProfile for CEX open-api of client
type Scope string

const (
	System   Scope = "system"
	Exchange Scope = "exchange"
)

type Profile interface {
	Scope() Scope
}

// The SystemProfile store in redis: data struct: hash key: profile:user hashkey: userid value SystemProfile json.raw
// O(1)
type SystemProfile struct {
	Language int8 // CN: 1  EN: 2
	Theme    int8 // 1: Light  2: Dark
}

func (s *SystemProfile) Scope() Scope {
	return System
}

// The ExchangeProfile just role = USER can setting
// maybe i can setting more exchange
// NOTE: store in redis & mongo
// NOTE: The ExchangeProfile store in redis: data struct: hash key: profile:exchange:user:{{platform}} hashkey: userid value ExchangeSecret json.raw
// O(1)
type ExchangeProfile struct {
	Name     string                    // this exhcange the name or short-description
	Platform map[string]ExchangeSecret // {platform:ExchangeSecret}
}

// The ExchangeSecret will send to MQ, consume in exchange service, to sync user account data
// called exchange service use this profile, to request CEX open-api
// NOTE: a user -> more account -> more wallet -> more ledger -> more report
type ExchangeSecret struct {
	Tag        string `json:"tag"`         // is api key the tag
	Key        string `json:"key"`         // PpdFMg43aJlSH6CTuWaL6ArQ864ziIyzKwRazXM3SfA0a4qSeFY9udaA43aYIyOm
	PrivateKey string `json:"private_key"` // z4FDZ3CSnAAnEHFh9NKZDj5l23fj5fzmB2zZRQNcarNAqdzADKF4TmbLFbGkRxao
	Passphrase string `json:"passphrase"`  // just a string salt
}

func (s *ExchangeProfile) Scope() Scope {
	return Exchange
}

// DefaultProfile use English and USD
func NewDefaultProfile() Profile {
	return &SystemProfile{
		Language: 2,
		Theme:    1,
	}
}

func (u *User) ApplyEvent(evt event.Message) error {
	switch e := evt.(type) {
	case *events.UserCreatedMessage:
		u.ID = e.AggregateID
		u.Username = e.Username
		u.LoginID = e.LoginID
		u.Email = e.Email
		u.Profile = nil
		u.Status = commuser.Activated // User Created Message default status is Activated
		u.Deleted = 0
	case *events.UserStatusMessage:
		u.ID = e.AggregateID
		u.Status = e.Status
	}
	return nil
}
