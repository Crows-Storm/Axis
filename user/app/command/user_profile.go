package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

// The SaveProfileCommand Handler create and update profile for user
type SaveProfileCommand struct {
	ID string `json:"id"` // user id
}

func (s SaveProfileCommand) Validate() decorator.CommandError {
	//TODO implement me
	panic("implement me")
}

type SaveProfileResult struct {
}

type SaveProfileCommandHandler decorator.CommandHandler[SaveProfileCommand, SaveProfileResult]

// TODO: need implement it
func NewSaveProfileCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
	bus decorator.EventBus,
) SaveProfileCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[SaveProfileCommand, SaveProfileResult](
		saveProfileCommandHandler{repo: repo, bus: bus},
		metricsClient,
	)
}

type saveProfileCommandHandler struct {
	repo domain.Repository
	bus  decorator.EventBus
}

func (s saveProfileCommandHandler) Handle(ctx context.Context, cmd SaveProfileCommand) (result SaveProfileResult, err error) {
	//TODO implement me
	panic("implement me")
}
