package rbac

import "strings"

type Audience string

const (
	AudienceApp     Audience = "app"
	AudienceConsole Audience = "console"
)

func ParseAudience(raw string) (Audience, bool) {
	switch Audience(strings.ToLower(strings.TrimSpace(raw))) {
	case AudienceApp:
		return AudienceApp, true
	case AudienceConsole:
		return AudienceConsole, true
	default:
		return "", false
	}
}

func NormalizeAudience(raw string) Audience {
	if a, ok := ParseAudience(raw); ok {
		return a
	}
	return AudienceApp
}
