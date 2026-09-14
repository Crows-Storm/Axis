package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/server/cache"
)

type RegisterUserCommand struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c RegisterUserCommand) Validate() error {
	if c.LoginID == "" || c.Password == "" || c.Email == "" {
		return decorator.CommandExecutedError{
			Msg: "LoginID, password, and email cannot be empty.",
		}
	}
	return nil
}

type RegisterUserCommandHandler decorator.CommandHandler[RegisterUserCommand, struct{}]

func NewRegisterUserCommandHandler(
	userService provider.UserService,
	cacheClient cache.RueidisClient,
	metricsClient decorator.MetricsClient,
) RegisterUserCommandHandler {
	if userService == nil {
		panic("nil userService")
	}
	return decorator.ApplyCommandDecorators[RegisterUserCommand, struct{}](
		registerUserCommandHandler{userService: userService, cacheClient: cacheClient},
		metricsClient,
	)
}

type registerUserCommandHandler struct {
	cacheClient cache.RueidisClient
	userService provider.UserService
}

func (r registerUserCommandHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (struct{}, error) {

	// call user grpc interface to create user
	_, err := r.userService.CreateUser(ctx, &userpb.CreateUserRequest{
		LoginID:  cmd.LoginID,
		Password: cmd.Password,
		Email:    cmd.Email,
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
