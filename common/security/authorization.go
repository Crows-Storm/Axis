package security

import "context"

type LoginType string

const (
	LoginTypePassword      LoginType = "password"
	LoginTypeSMSCode       LoginType = "sms_code"
	LoginTypeEmailCode     LoginType = "email_code"
	LoginTypeOAuth         LoginType = "oauth"
	LoginTypeQRCode        LoginType = "qrcode"
	LoginTypeWxMiniProgram LoginType = "wx_mini_program"
)

// AuthenticatedIdentity TODO: I thing store to mongo? maybe
type AuthenticatedIdentity struct {
	UserId    int64
	UnionId   string            // Unified identity identifier (phone number/email address/OAuth unionID)
	Channel   string            // Source: "phone", "email", "wechat", "github"
	IsNewUser bool              // is first login?
	Extra     map[string]string // Extended information (avatar, nickname, etc., from third parties)
}

// Authenticator authenticator interface (strategy mode)
//
// Each login method implements this interface and has a single responsibility:
// Verify credentials → Return identity (does not care about registration, permissions, Token)
type Authenticator interface {
	Authenticate(ctx context.Context, credential Credential) (*AuthenticatedIdentity, error)

	LoginType() LoginType
}

type Credential interface {
	GetLoginType() LoginType
	Validate() error
}
