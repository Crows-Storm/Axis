package command

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
)

type OpenAccountCommand struct {
	UserID uint64
	Type   string
}

func (c OpenAccountCommand) Validate() decorator.CommandError {
	if c.UserID == 0 || c.Type == "" {
		return decorator.InvalidCommand("user id and type are required")
	}
	if _, ok := commwallet.ParseType(c.Type); !ok {
		return decorator.InvalidCommand("unknown account type")
	}
	return nil
}

type OpenAccountCommandHandler decorator.CommandHandler[OpenAccountCommand, *domain.Account]

type openAccountCommandHandler struct {
	repo domain.Repository
	bus  decorator.EventBus
}

func NewOpenAccountCommandHandler(repo domain.Repository, metrics decorator.MetricsClient, bus decorator.EventBus) OpenAccountCommandHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyCommandDecorators[OpenAccountCommand, *domain.Account](
		openAccountCommandHandler{repo: repo, bus: bus},
		metrics,
	)
}

func (h openAccountCommandHandler) Handle(ctx context.Context, cmd OpenAccountCommand) (*domain.Account, error) {
	kind, _ := commwallet.ParseType(cmd.Type)
	acc, err := domain.Open(cmd.UserID, kind)
	if err != nil {
		return nil, fmt.Errorf("open account: %w", err)
	}
	if err := h.repo.Save(ctx, acc); err != nil {
		return nil, fmt.Errorf("open account: %w", err)
	}
	if err := publishAccountEvents(ctx, h.bus, acc); err != nil {
		return nil, err
	}
	return acc, nil
}
