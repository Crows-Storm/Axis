package auth

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/security"
)

type AuthAppService struct {
	authenticators map[security.LoginType]security.Authenticator // 策略注册表
	tokenIssuer    jwt.TokenIssuer
}

func NewAuthAppService(
	ti jwt.TokenIssuer,
	authenticators ...security.Authenticator,
) *AuthAppService {
	m := make(map[security.LoginType]security.Authenticator, len(authenticators))
	for _, a := range authenticators {
		m[a.LoginType()] = a
	}
	return &AuthAppService{
		authenticators: m,
		tokenIssuer:    ti,
	}
}

func (a *AuthAppService) GetTokenIssuer() *jwt.TokenIssuer {
	return &a.tokenIssuer
}

func (a *AuthAppService) Match(loginType security.LoginType) (security.Authenticator, error) {
	strategy, ok := a.authenticators[loginType]
	if !ok {
		return nil, fmt.Errorf("unknown authenticator type: %s", loginType)
	}
	return strategy, nil
}

func (a *AuthAppService) IssueToken(ctx context.Context, principal *domain.Principal) (*jwt.TokenPayload, error) {
	return a.tokenIssuer.Issue(ctx, principal)
}

func (a *AuthAppService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPayload, error) {
	return a.tokenIssuer.Refresh(ctx, refreshToken)
}

func (a *AuthAppService) Logout(ctx context.Context, accessToken string) error {
	token, err := a.tokenIssuer.Preprocessing(ctx, accessToken)
	if err != nil {
		return err
	}
	return a.tokenIssuer.Revoke(ctx, token)
}

var (
	ErrUserNotFound    = errors.New("user not found, please register first")
	ErrAccountDisabled = errors.New("account is disabled")
)
