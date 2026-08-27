package principal

import (
	"fmt"
	"strings"
)

type RoleCode string

const (
	RoleCodeAdmin RoleCode = "ADMIN"
	//RoleCodeManager RoleCode = "MANAGER"
	RoleCodeUser RoleCode = "USER"
)

var validRoleCodes = map[RoleCode]struct{}{
	RoleCodeAdmin: {},
	//RoleCodeManager: {},
	RoleCodeUser: {},
}

func NewRoleCode(code string) (RoleCode, error) {
	c := RoleCode(strings.ToUpper(strings.TrimSpace(code)))
	if _, ok := validRoleCodes[c]; !ok {
		return "", fmt.Errorf("invalid role code: %s", code)
	}
	return c, nil
}

func (r RoleCode) String() string {
	return string(r)
}

func (r RoleCode) IsAdmin() bool {
	return r == RoleCodeAdmin
}

func (r RoleCode) Equals(other RoleCode) bool {
	return r == other
}
