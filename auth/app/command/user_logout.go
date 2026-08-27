package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/domain/principal"
)

// LogoutCommand / LogoutCommandHandler
// Logout (Revoke Token)
type LogoutCommand struct {
	AccessToken string
}

type LogoutResult struct{}

type LogoutCommandHandler decorator.CommandHandler[LogoutCommand, LogoutResult]

func NewLogoutCommandHandler(
	authApp *auth.AuthAppService,
	metricsClient decorator.MetricsClient,
) LogoutCommandHandler {
	if authApp == nil {
		panic("nil authApp service")
	}

	return decorator.ApplyCommandDecorators[LogoutCommand, LogoutResult](
		logoutCommandHandler{
			authApp: authApp,
		},
		metricsClient,
	)
}

type logoutCommandHandler struct {
	authApp *auth.AuthAppService
}

func (h logoutCommandHandler) Handle(ctx context.Context, cmd LogoutCommand) (LogoutResult, error) {

	v := principal.FromContext(ctx)
	_ = fmt.Sprintf("%v", v)

	if err := h.authApp.Logout(ctx, cmd.AccessToken); err != nil {
		return LogoutResult{}, fmt.Errorf("revoke token failed: %w", err)
	}
	return LogoutResult{}, nil
}
