package command

import (
	"context"

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
	userGRPC UserService,
	cacheClient cache.RueidisClient,
	metricsClient decorator.MetricsClient,
) RegisterUserCommandHandler {
	if userGRPC == nil {
		panic("nil userGRPC")
	}
	return decorator.ApplyCommandDecorators[RegisterUserCommand, struct{}](
		registerUserCommandHandler{userGRPC: userGRPC, cacheClient: cacheClient},
		metricsClient,
	)
}

type registerUserCommandHandler struct {
	cacheClient cache.RueidisClient
	userGRPC    UserService
}

func (r registerUserCommandHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (struct{}, error) {
	if cmd.LoginId == "" || cmd.Password == "" || cmd.Email == "" {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "LoginId, password, and email cannot be empty.",
		}
	}

	// call user grpc interface to create user
	_, err := r.userGRPC.CreateUser(ctx, &userpb.CreateUserRequest{
		LoginId:  cmd.LoginId,
		Password: cmd.Password,
		Email:    cmd.Email,
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
