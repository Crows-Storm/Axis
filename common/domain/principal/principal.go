package principal

type Principal struct {
	UserID   uint64   `json:"user_id"`
	Username string   `json:"username"`
	LoginID  string   `json:"login_id"`
	Email    string   `json:"email"`
	Status   int8     `json:"status"`
	Role     HoldRole `json:"role"` // a user just can hold a role
	// Permissions map[string]struct{} `json:"permissions"`
	AuthChannel string            `json:"auth_channel"`
	Extra       map[string]string `json:"extra"`
}

const ExtraAudience = "audience"

// NewPrincipal The Permissions field is missing because it is hot-updated and stored in Redis.
func NewPrincipal(userId uint64, username string, LoginID string, email string, status int8, role HoldRole, authChannel string, loginFrom string, extra map[string]string) *Principal {
	return &Principal{
		UserID:      userId,
		Username:    username,
		LoginID:     LoginID,
		Email:       email,
		Status:      status,
		Role:        role,
		AuthChannel: authChannel,
		Extra:       extra,
	}
}

// HoldRole is this principal hold the role
type HoldRole struct {
	RoleID    int64    `json:"role_id"`
	RoleCode  RoleCode `json:"role_code"`
	RoleLevel int      `json:"role_level"`
}

// ---- 鉴权方法 ----

// IsAuthenticated 是否已认证
func (p *Principal) IsAuthenticated() bool {
	return p.UserID > 0
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

func (p *Principal) Audience() string {
	if p == nil || p.Extra == nil {
		return "app"
	}
	v := p.Extra[ExtraAudience]
	if v == "" {
		return "app"
	}
	return v
}
