package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/domain/principal"
)

type RefreshTokenCommand struct {
}

func (c *RefreshTokenCommand) Validate() error {
	// no everything now
	return nil
}

type RefreshTokenResult struct {
	NewAccessToken string
}

type RefreshTokenCommandHandler decorator.CommandHandler[RefreshTokenCommand, RefreshTokenResult]

func NewRefreshTokenCommandHandler(
	userService provider.UserService,
	authenticators *auth.AuthAppService,
	metricsClient decorator.MetricsClient,
) RefreshTokenCommandHandler {
	return decorator.ApplyCommandDecorators[RefreshTokenCommand, RefreshTokenResult](
		refreshTokenCommandHandler{
			userService:    userService,
			authenticators: authenticators,
		},
		metricsClient,
	)
}

type refreshTokenCommandHandler struct {
	userService    provider.UserService
	authenticators *auth.AuthAppService
}

func (h refreshTokenCommandHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (RefreshTokenResult, error) {
	_, err := principal.GetUserIDFromContext(ctx) // use UserID to query user info from user GRPC service
	if err != nil {
		return RefreshTokenResult{}, fmt.Errorf("unauthorized: missing principal")
	}
	// TODO: build a new principal from user/role domain
	var p principal.Principal

	refreshToken, err := h.authenticators.RefreshToken(ctx, "", &p)

	if err != nil {
		return RefreshTokenResult{}, fmt.Errorf("refresh token failed: %w", err)
	}
	return RefreshTokenResult{NewAccessToken: refreshToken.AccessToken}, nil
}
