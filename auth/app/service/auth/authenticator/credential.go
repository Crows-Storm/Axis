package authenticator

import (
	"errors"

	"github.com/Crows-Storm/Axis/common/security"
)

type PasswordCredential struct {
	LoginId  string `json:"LoginId" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *PasswordCredential) GetLoginType() security.LoginType { return security.LoginTypePassword }
func (c *PasswordCredential) Validate() error {
	if c.LoginId == "" {
		return errors.New("LoginId is required")
	}
	if len(c.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

type CodeCredential struct {
	LoginType_ security.LoginType `json:"login_type"` // sms_code or email_code
	Account    string             `json:"account"`    // Phone number or Email account
	Code       string             `json:"code"`       // Verification code
}

func (c *CodeCredential) GetLoginType() security.LoginType { return c.LoginType_ }
func (c *CodeCredential) Validate() error {
	if c.Account == "" {
		return errors.New("account is required")
	}
	if len(c.Code) != 6 {
		return errors.New("verification code must be 6 digits")
	}
	return nil
}

type OAuthCredential struct {
	Provider string `json:"provider"` // "wechat", "github", "google"
	Code     string `json:"code"`     // OAuth authorization code
	State    string `json:"state"`    // CSRF state
}

func (c *OAuthCredential) GetLoginType() security.LoginType { return security.LoginTypeOAuth }
func (c *OAuthCredential) Validate() error {
	if c.Provider == "" {
		return errors.New("oauth provider is required")
	}
	if c.Code == "" {
		return errors.New("oauth code is required")
	}
	return nil
}

type QRCodeCredential struct {
	Ticket string `json:"ticket"` // QR code ticket (Generated after authorization is confirmed by the scanned device)
}

func (c *QRCodeCredential) GetLoginType() security.LoginType { return security.LoginTypeQRCode }
func (c *QRCodeCredential) Validate() error {
	if c.Ticket == "" {
		return errors.New("qrcode ticket is required")
	}
	return nil
}

type RegisterCredential struct {
	Inner      security.Credential // Embedded actual authentication credentials (CodeCredential / OAuthCredential)
	LoginId    string              `json:"LoginId"`
	Nickname   string              `json:"nickname"`
	AgreeTerms bool                `json:"agree_terms"`
}

func (c *RegisterCredential) GetLoginType() security.LoginType { return c.Inner.GetLoginType() }
func (c *RegisterCredential) Validate() error {
	if !c.AgreeTerms {
		return errors.New("must agree to terms of service")
	}
	if c.Inner == nil {
		return errors.New("inner credential is required")
	}
	return c.Inner.Validate()
}

type WxMiniProgramCredential struct {
	Code string `json:"code" binding:"required"`
}

func (c *WxMiniProgramCredential) GetLoginType() security.LoginType {
	return security.LoginTypeWxMiniProgram
}

func (c *WxMiniProgramCredential) Validate() error {
	if c.Code == "" {
		return errors.New("wechat mini program code is required")
	}
	if len(c.Code) < 10 || len(c.Code) > 256 {
		return errors.New("invalid wechat mini program code format")
	}
	return nil
}
