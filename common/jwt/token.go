package jwt

import (
	"context"
	"time"

	domain "github.com/Crows-Storm/Axis/common/domain/principal"
)

type TokenPayload struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"` // "Bearer"
	ExpiresIn    time.Duration `json:"expires_in"` // access_token expires in
}

// TokenIssuer Token Issuance/Verification Interface
type TokenIssuer interface {
	Issue(ctx context.Context, principal *domain.Principal) (*TokenPayload, error)

	Refresh(ctx context.Context, refreshToken string) (*TokenPayload, error)

	Revoke(ctx context.Context, accessToken string) error

	Parse(ctx context.Context, accessToken string) (*domain.Principal, error)

	Preprocessing(ctx context.Context, token string) (string, error)
}
