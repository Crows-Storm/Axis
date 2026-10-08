package command

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type OpenWalletCommand struct {
	AccountID  uint64
	CurrencyID uint64
	Symbol     string
}

func (c OpenWalletCommand) Validate() decorator.CommandError {
	if c.AccountID == 0 {
		return decorator.InvalidCommand("account id is required")
	}
	if c.CurrencyID == 0 && c.Symbol == "" {
		return decorator.InvalidCommand("currency_id or symbol is required")
	}
	return nil
}

type OpenWalletCommandHandler decorator.CommandHandler[OpenWalletCommand, domain.Wallet]

type openWalletCommandHandler struct {
	repo domain.Repository
	bus  decorator.EventBus
}

func NewOpenWalletCommandHandler(repo domain.Repository, metrics decorator.MetricsClient, bus decorator.EventBus) OpenWalletCommandHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyCommandDecorators[OpenWalletCommand, domain.Wallet](
		openWalletCommandHandler{repo: repo, bus: bus},
		metrics,
	)
}

func (h openWalletCommandHandler) Handle(ctx context.Context, cmd OpenWalletCommand) (domain.Wallet, error) {
	acc, err := h.repo.Get(ctx, cmd.AccountID)
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("open wallet: %w", err)
	}
	currency, err := h.repo.GetCurrency(ctx, cmd.CurrencyID, cmd.Symbol)
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("open wallet: %w", err)
	}
	wallet, err := acc.OpenWallet(currency)
	if err != nil {
		return domain.Wallet{}, err
	}
	if err := h.repo.Save(ctx, acc); err != nil {
		return domain.Wallet{}, fmt.Errorf("open wallet: %w", err)
	}
	if err := publishAccountEvents(ctx, h.bus, acc); err != nil {
		return domain.Wallet{}, err
	}
	return wallet, nil
}
