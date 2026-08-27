package principal

type Principal struct {
	UserId      int64
	Username    string
	LoginId     string
	Email       string
	Status      int8
	Role        HoldRole            // a user just can hold a role
	Permissions map[string]struct{} // 权限标识集合（O(1)查找）
	AuthChannel string
	Extra       map[string]string
}

// NewPrincipal The Permissions field is missing because it is hot-updated and stored in Redis.
func NewPrincipal(userId int64, username string, loginId string, email string, status int8, role HoldRole, authChannel string, loginFrom string, extra map[string]string) *Principal {
	return &Principal{
		UserId:      userId,
		Username:    username,
		LoginId:     loginId,
		Email:       email,
		Status:      status,
		Role:        role,
		AuthChannel: authChannel,
		Extra:       extra,
	}
}

// HoldRole is this principal hold the role
type HoldRole struct {
	RoleID    int64
	RoleCode  RoleCode
	RoleLevel int
}

// ---- 鉴权方法 ----

// IsAuthenticated 是否已认证
func (p *Principal) IsAuthenticated() bool {
	return p.UserId > 0
}

// IsAdmin 是否管理员
func (p *Principal) IsAdmin() bool {
	if p.Role.RoleCode.IsAdmin() {
		return true
	}
	return false
}

func (p *Principal) HasRole(roleCode RoleCode) bool {
	if p.Role.RoleCode.Equals(roleCode) {
		return true
	}
	return false
}

// CanOperateRole 判断是否能操作目标角色级别
func (p *Principal) CanOperateRole(targetRoleLevel int) bool {
	if p.IsAdmin() {
		return true
	}
	return p.Role.RoleLevel <= targetRoleLevel
}
