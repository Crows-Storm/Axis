package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/security"
)

type LoginCommand struct {
	LoginType  security.LoginType
	Credential security.Credential
}

func (c *LoginCommand) Validate() error {
	if err := c.Credential.Validate(); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}
	return nil
}

type LoginResult struct {
	UserId   int64
	Username string
	IsNew    bool
	Token    *jwt.TokenPayload
}

type LoginCommandHandler decorator.CommandHandler[LoginCommand, LoginResult]

func NewLoginCommandHandler(
	userService provider.UserService,
	authenticators *auth.AuthAppService,
	metricsClient decorator.MetricsClient,
) LoginCommandHandler {
	return decorator.ApplyCommandDecorators[LoginCommand, LoginResult](
		loginCommandHandler{
			userService:    userService,
			authenticators: authenticators,
		},
		metricsClient,
	)
}

type loginCommandHandler struct {
	userService    provider.UserService
	authenticators *auth.AuthAppService
}

func (h loginCommandHandler) Handle(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
	strategy, err := h.authenticators.Match(cmd.LoginType)
	if err != nil {
		return LoginResult{}, fmt.Errorf("unsupported login type: %s", cmd.LoginType)
	}
	if err := cmd.Validate(); err != nil {
		return LoginResult{}, fmt.Errorf("credential validation failed: %w", err)
	}

	// get Authenticate identity
	identity, err := strategy.Authenticate(ctx, cmd.Credential)
	if err != nil {
		return LoginResult{}, fmt.Errorf("authentication failed: %w", err)
	}

	// TODO: maybe should create strategy to login/register and login by identity.IsNewUser?
	if identity.IsNewUser {
		return LoginResult{}, auth.ErrUserNotFound
	}

	// Obtaining user information for constructing the principal
	userInfo, err := h.userService.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: identity.UserId,
	})

	// TODO: Need to query Role and Permissions from Role and Permission domain, domain query redis
	holdRole := domain.HoldRole{}
	// builder principal
	principal := &domain.Principal{
		UserId:      userInfo.Id,
		Username:    "",
		LoginId:     userInfo.LoginId,
		Email:       userInfo.Email,
		Status:      int8(userInfo.Status),
		Role:        holdRole,
		AuthChannel: string(cmd.LoginType),
		Extra:       nil,
	}
	token, err := h.authenticators.IssueToken(ctx, principal)

	if err != nil {
		return LoginResult{}, fmt.Errorf("issue token failed: %w", err)
	}
	return LoginResult{UserId: identity.UserId, Username: principal.Username, IsNew: false, Token: token}, nil
}
