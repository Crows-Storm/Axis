package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/server/cache"
)

type RegisterUserCommand struct {
	LoginId  string
	Password string
	Email    string
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
	if cmd.LoginId == "" || cmd.Password == "" || cmd.Email == "" {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "LoginId, password, and email cannot be empty.",
		}
	}

	// call user grpc interface to create user
	_, err := r.userService.CreateUser(ctx, &userpb.CreateUserRequest{
		LoginId:  cmd.LoginId,
		Password: cmd.Password,
		Email:    cmd.Email,
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
