package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/sirupsen/logrus"
)

type RegisterUserCommand struct {
	LoginId  string
	Password string
	Email    string
}

type RegisterUserCommandHandler decorator.CommandHandler[RegisterUserCommand, struct{}]

func NewRegisterUserCommandHandler(
	userGRPC UserService,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) RegisterUserCommandHandler {
	if userGRPC == nil {
		panic("nil userGRPC")
	}
	return decorator.ApplyCommandDecorators[RegisterUserCommand, struct{}](
		registerUserCommandHandler{userGRPC: userGRPC},
		logger,
		metricsClient,
	)
}

type registerUserCommandHandler struct {
	userGRPC UserService
}

func (r registerUserCommandHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (struct{}, error) {
	// do something: send verification to sms/email

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
