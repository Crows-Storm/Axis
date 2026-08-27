package menu

import (
	"fmt"
	"regexp"
	"strings"
)

// Permission 权限标识值对象，格式: module:resource:action
// 例如: system:user:add, system:user:edit, system:role:list
type Permission string

var permissionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*(?::[a-z][a-z0-9_]*)?$`)

func NewPermission(perm string) (Permission, error) {
	p := Permission(strings.TrimSpace(strings.ToLower(perm)))
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}

func (p Permission) Validate() error {
	if p == "" {
		return nil // permission 可以为空（目录类型菜单）
	}
	if !permissionPattern.MatchString(string(p)) {
		return fmt.Errorf("invalid permission format: %s, expected format like 'module:resource:action'", p)
	}
	return nil
}

func (p Permission) String() string {
	return string(p)
}

func (p Permission) IsEmpty() bool {
	return p == ""
}

// Matches 权限匹配，支持通配符
// "system:user:*" 可以匹配 "system:user:add"
func (p Permission) Matches(target Permission) bool {
	if p == target {
		return true
	}
	pStr := string(p)
	tStr := string(target)
	if strings.HasSuffix(pStr, ":*") {
		prefix := strings.TrimSuffix(pStr, "*")
		return strings.HasPrefix(tStr, prefix)
	}
	return false
}

// Module 提取模块部分
func (p Permission) Module() string {
	parts := strings.SplitN(string(p), ":", 3)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
