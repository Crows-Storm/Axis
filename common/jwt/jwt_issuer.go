package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	domain "github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/util"
	"github.com/golang-jwt/jwt/v5"
)

// JWTIssuer JWT Token issuer implement
type JWTIssuer struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessTTL       time.Duration
	refreshTTL      time.Duration
	tokenRepository TokenRepository // blacklist/whitelist store
}

// TokenRepository sure is a repository interface, query in redis
type TokenRepository interface {
	// StoreRefreshToken store refreshToken（use in refresh + cancel）
	StoreRefreshToken(ctx context.Context, jti string, userID int64, ttl time.Duration) error
	// CheckRefreshToken check refreshToken is effective
	CheckRefreshToken(ctx context.Context, jti string) (userID int64, err error)
	// RevokeRefreshToken cancel refreshToken
	RevokeRefreshToken(ctx context.Context, jti string) error
	// BlacklistAccessToken do accessToken add to blacklist（logout）
	BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error
	// IsBlacklisted check accessToken in the blacklist
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

func NewJWTIssuer(cfg config.JWTConfig, repository TokenRepository) *JWTIssuer {
	return &JWTIssuer{
		accessSecret:    []byte(cfg.AccessSecret),
		refreshSecret:   []byte(cfg.RefreshSecret),
		accessTTL:       cfg.AccessTTL,
		refreshTTL:      cfg.RefreshTTL,
		tokenRepository: repository,
	}
}

// JWT Claim

type AuthClaims struct {
	jwt.RegisteredClaims

	Token    string          `json:"token"` // jwt token
	UserId   int64           `json:"userId"`
	Username string          `json:"username"`
	LoginId  string          `json:"loginId"`
	Email    string          `json:"email"`
	Status   int8            `json:"status"`
	Role     domain.HoldRole `json:"role"` // a user just can hold a role
	//Permissions map[string]struct{} `json:"permissions"` // Permission identifier set (O(1) lookup)
	AuthChannel string            `json:"authChannel"` // password/sms_code/email_code/oauth/qrcode
	Extra       map[string]string `json:"extra"`
}

func (j *JWTIssuer) Issue(ctx context.Context, principal *domain.Principal) (*TokenPayload, error) {
	if principal == nil {
		return nil, errors.New("principal is nil")
	}
	now := time.Now()

	// Access Token
	accessJTI := generateJTI()
	expiresAt := now.Add(j.accessTTL)
	accessClaims := AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        accessJTI,
			Issuer:    "auth",
			Subject:   fmt.Sprintf("%d", principal.UserId),
		},
		UserId:      principal.UserId,
		Username:    principal.Username,
		LoginId:     principal.LoginId,
		Email:       principal.Email,
		Status:      principal.Status,
		Role:        principal.Role,
		AuthChannel: principal.AuthChannel,
		Extra:       principal.Extra,
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString(j.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token failed: %w", err)
	}

	// Refresh Token
	refreshJTI := generateJTI()
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        refreshJTI,
		Subject:   fmt.Sprintf("%d", principal.UserId),
		Issuer:    "auth",
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString(j.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token failed: %w", err)
	}

	// store refreshToken mapping
	if err := j.tokenRepository.StoreRefreshToken(ctx, refreshJTI, principal.UserId, j.refreshTTL); err != nil {
		return nil, fmt.Errorf("store refresh token failed: %w", err)
	}

	return &TokenPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    j.accessTTL,
	}, nil
}

// ---- Refresh ----

func (j *JWTIssuer) Refresh(ctx context.Context, refreshToken string) (*TokenPayload, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return j.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired refresh token")
	}

	userID, err := j.tokenRepository.CheckRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, errors.New("refresh token has been revoked")
	}

	_ = j.tokenRepository.RevokeRefreshToken(ctx, claims.ID)

	// TODO reload Principal to issue new Token (UserProvider needs to be injected externally here to simplify processing)
	// In actual production, the latest permissions should be reloaded from DB
	_ = userID
	return nil, errors.New("refresh requires UserProvider injection — see production note")
}

// Revoke
func (j *JWTIssuer) Revoke(ctx context.Context, accessToken string) error {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
		return j.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil // Token It is invalid and does not need to be revoked
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		return j.tokenRepository.BlacklistAccessToken(ctx, claims.ID, remaining)
	}
	return nil
}

func (j *JWTIssuer) Parse(ctx context.Context, accessToken string) (*domain.Principal, error) {
	//if !strings.HasPrefix(accessToken, BearerPrefix) {
	//	return nil, errors.New("invalid or expired access token")
	//}

	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired access token")
	}

	blacklisted, _ := j.tokenRepository.IsBlacklisted(ctx, claims.ID)
	if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	return domain.NewPrincipal(
		claims.UserId,      // UserId
		claims.Username,    // Username
		claims.LoginId,     // LoginId
		claims.Email,       // Email
		claims.Status,      // Status
		claims.Role,        // Role
		claims.AuthChannel, // AuthChannel
		"",                 // LoginFrom (get in claims)
		claims.Extra,       // Extra
	), nil
}

// The Preprocessing of tokens
func (j *JWTIssuer) Preprocessing(ctx context.Context, token string) (string, error) {
	// cut off bearer prefix
	bearerPrefix := "Bearer "
	if strings.Contains(token, bearerPrefix) {
		_, after, ok := strings.Cut(token, bearerPrefix)
		if ok {
			token = after
		}
	} else {
		return "", errors.New("invalid or expired token")
	}

	// can do other process

	return token, nil
}

// ================= helpers =================

// generateJTI generates a unique JWT ID (JTI) using a version 1 UUID.
// The resulting ID is used to uniquely identify a JWT token for tracking,
// revocation, and preventing replay attacks.
func generateJTI() string {
	return util.UUIDV1().String()
}
