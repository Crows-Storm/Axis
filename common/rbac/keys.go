package rbac

import (
	"fmt"
	"strings"
)

const (
	roleLoadedField = "_"
	emptyRoleMember = "0"
)

func userRolesKey(userID uint64) string {
	return fmt.Sprintf("rbac:user:%d:roles", userID)
}

func roleKey(roleID uint64) string {
	return fmt.Sprintf("rbac:role:%d", roleID)
}

func routeKey(audience Audience, method, path string) string {
	return fmt.Sprintf("rbac:route:%s:%s:%s", audience, strings.ToUpper(method), path)
}
