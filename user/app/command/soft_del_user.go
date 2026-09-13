package command

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/jwt"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type SoftDeleteUserCommand struct {
	ID int64
}

func (s *SoftDeleteUserCommand) Validate() error {
	if s.ID <= 0 {
		return decorator.CommandExecutedError{Msg: "Invalid user ID"}
	}
	return nil
}

type SoftDeleteUserCommandHandler decorator.CommandHandler[SoftDeleteUserCommand, struct{}]

type softDeleteUserCommandHandler struct {
	repo domain.Repository
	jwt  jwt.TokenIssuer
}

func NewSoftDeleteUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
	jwt jwt.TokenIssuer,
) SoftDeleteUserCommandHandler {
	if repo == nil {
		panic("nil User repo")
	}
	return decorator.ApplyCommandDecorators[SoftDeleteUserCommand, struct{}](
		softDeleteUserCommandHandler{repo: repo, jwt: jwt},
		metricsClient,
	)
}

// Handle implementation of `SoftDeleteUserCommand` returns void
func (c softDeleteUserCommandHandler) Handle(ctx context.Context, cmd SoftDeleteUserCommand) (struct{}, error) {
	user, err := c.repo.GetInfo(cmd.ID)
	if err != nil {
		return struct{}{}, domain.NotFoundError{UserId: cmd.ID}
	}
	if err := user.Delete(); err != nil {
		return struct{}{}, domain.NotFoundError{UserId: cmd.ID}
	}

	if err := c.repo.SoftDelete(ctx, cmd.ID); err != nil {
		logger.Warnf("failed to soft delete user: %v", err)
		return struct{}{}, errors.New("failed to soft delete user")
	}

	// TODO: revoke accessToken to blocklist
	//accessToken := ""
	//if err := c.jwt.Revoke(ctx, accessToken); err != nil {
	//	return struct{}{}, err
	//}

	return struct{}{}, nil
}
