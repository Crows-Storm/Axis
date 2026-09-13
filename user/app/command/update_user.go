package command

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type UpdateUserCommand struct {
	ID    int64
	Email *string `json:"email"` // value can nil
	// ... other fields
}

func (u *UpdateUserCommand) Validate() error {
	if u.ID == 0 {
		return errors.New("user ID cannot be nil")
	}

	// all field cannot be nil
	if u.Email == nil {
		return errors.New("you haven't changed anything")
	}
	return nil
}

type UpdateUserCommandHandler decorator.CommandHandler[UpdateUserCommand, struct{}]

type updateUserCommandHandler struct {
	repo domain.Repository
}

func NewUpdateUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) UpdateUserCommandHandler {
	if repo == nil {
		panic("nil User repo")
	}
	return decorator.ApplyCommandDecorators[UpdateUserCommand, struct{}](
		updateUserCommandHandler{
			repo: repo,
		},
		metricsClient,
	)
}

func (u updateUserCommandHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (struct{}, error) {
	// get domain from repo
	//user, err := u.repo.GetInfo(cmd.ID)
	//if err != nil {
	//	logger.Debugf("Not found user by id %d: %v", cmd.ID, err)
	//	var notFoundError = domain.NotFoundError{UserId: cmd.ID}
	//	return struct{}{}, fmt.Errorf("the UpdateUserCommand command failed to Execute, because: %v", notFoundError)
	//}
	//
	//tracker := commdomain.NewDiffTracker()
	//
	//// check email
	//if cmd.Email != nil {
	//	tracker.TrackStringChange("email", user.Email, *cmd.Email)
	//	if err := user.ChangeEmail(*cmd.Email); err != nil {
	//		return struct{}{}, err
	//	}
	//}
	//
	//// call domain service to save and publish domain events
	//if err := u.repo.Update(ctx, user); err != nil {
	//	return struct{}{}, decorator.CommandExecutedError{Msg: fmt.Sprintf("The UpdateUserCommand command failed to Execute, because: %v", err)}
	//}
	return struct{}{}, nil
}
