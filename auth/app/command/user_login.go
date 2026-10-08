package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/security"
)

type LoginCommand struct {
	LoginType  security.LoginType
	Audience   string
	Credential security.Credential
}

func (c *LoginCommand) Validate() error {
	if err := c.Credential.Validate(); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}
	return nil
}

type LoginResult struct {
	UserID   uint64            `json:"user_id"`
	Username string            `json:"username"`
	IsNew    bool              `json:"is_new"`
	Token    *jwt.TokenPayload `json:"token"`
}

type LoginCommandHandler decorator.CommandHandler[LoginCommand, LoginResult]

func NewLoginCommandHandler(
	userService provider.UserService,
	authenticators *auth.AuthAppService,
	roleRepo role.Repository,
	metricsClient decorator.MetricsClient,
) LoginCommandHandler {
	return decorator.ApplyCommandDecorators[LoginCommand, LoginResult](
		loginCommandHandler{
			userService:    userService,
			authenticators: authenticators,
			roleRepo:       roleRepo,
		},
		metricsClient,
	)
}

type loginCommandHandler struct {
	userService    provider.UserService
	authenticators *auth.AuthAppService
	roleRepo       role.Repository
}

func (h loginCommandHandler) Handle(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
	strategy, err := h.authenticators.Match(cmd.LoginType)
	if err != nil {
		return LoginResult{}, fmt.Errorf("unsupported login type: %s", cmd.LoginType)
	}
	//if err := cmd.Validate(); err != nil {
	//	return LoginResult{}, fmt.Errorf("credential validation failed: %w", err)
	//}

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
		Id: identity.UserID,
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("get user: %w", err)
	}

	holdRole := domain.HoldRole{}
	if h.roleRepo != nil {
		roles, roleErr := h.roleRepo.GetByUserID(ctx, identity.UserID)
		if roleErr != nil {
			return LoginResult{}, fmt.Errorf("get user roles: %w", roleErr)
		}
		if len(roles) > 0 && roles[0] != nil {
			holdRole = domain.HoldRole{
				RoleID:   int64(roles[0].ID),
				RoleCode: domain.RoleCode(roles[0].Code),
			}
		}
	}

	principal := &domain.Principal{
		UserID:      userInfo.Id,
		Username:    "",
		LoginID:     userInfo.LoginID,
		Email:       userInfo.Email,
		Status:      int8(userInfo.Status),
		Role:        holdRole,
		AuthChannel: string(cmd.LoginType),
		Extra: map[string]string{
			domain.ExtraAudience: audienceOrApp(cmd.Audience),
		},
	}
	token, err := h.authenticators.IssueToken(ctx, principal)

	if err != nil {
		return LoginResult{}, fmt.Errorf("issue token failed: %w", err)
	}
	return LoginResult{UserID: identity.UserID, Username: principal.Username, IsNew: false, Token: token}, nil
}

func audienceOrApp(raw string) string {
	if raw == "" {
		return "app"
	}
	return raw
}
