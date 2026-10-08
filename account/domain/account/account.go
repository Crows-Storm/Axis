package account

import (
	"fmt"
	"math/big"

	"github.com/Crows-Storm/Axis/common/domain"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	"github.com/Crows-Storm/Axis/common/domain/account/events"
	"github.com/Crows-Storm/Axis/common/domain/event"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
	"github.com/Crows-Storm/Axis/common/util"
)

type Account struct {
	domain.AggregateRoot

	ID     uint64
	UserID uint64
	Wallet map[uint64]Wallet `json:"wallet,omitempty"`
	Type   commwallet.Type
}

type Wallet struct {
	ID       uint64
	Currency commaccount.Currency
	Balance  *big.Int
}

func Open(userID uint64, kind commwallet.Type) (*Account, error) {
	if userID == 0 {
		return nil, fmt.Errorf("open account: %w", commaccount.ErrInvalid)
	}
	if _, ok := commwallet.ParseType(string(kind)); !ok {
		return nil, fmt.Errorf("open account: %w", commaccount.ErrInvalid)
	}
	acc := &Account{}
	if err := acc.RaiseEvent(acc, events.NewAccountOpenedMessage(util.GenerateFlakeID(), userID, string(kind))); err != nil {
		return nil, fmt.Errorf("open account: %w", err)
	}
	return acc, nil
}

func (a *Account) OpenWallet(currency commaccount.Currency) (Wallet, error) {
	if a == nil || a.ID == 0 {
		return Wallet{}, fmt.Errorf("open wallet: %w", commaccount.ErrInvalid)
	}
	if currency.ID == 0 || currency.Symbol == "" {
		return Wallet{}, fmt.Errorf("open wallet: %w", commaccount.ErrInvalid)
	}
	for _, existing := range a.Wallet {
		if existing.Currency.ID == currency.ID {
			return Wallet{}, fmt.Errorf("open wallet: %w", commaccount.ErrInvalid)
		}
	}
	walletID := util.GenerateFlakeID()
	if err := a.RaiseEvent(a, events.NewWalletOpenedMessage(a.ID, walletID, currency.ID, currency.Symbol)); err != nil {
		return Wallet{}, fmt.Errorf("open wallet: %w", err)
	}
	opened := a.Wallet[walletID]
	opened.Currency = currency
	a.Wallet[walletID] = opened
	return opened, nil
}

func Reconstruct(id, userID uint64, kind commwallet.Type, wallets map[uint64]Wallet) *Account {
	if wallets == nil {
		wallets = map[uint64]Wallet{}
	}
	return &Account{ID: id, UserID: userID, Type: kind, Wallet: wallets}
}

func (a *Account) ApplyEvent(evt event.Message) error {
	switch e := evt.(type) {
	case *events.AccountOpenedMessage:
		a.ID = e.AggregateID
		a.UserID = e.UserID
		a.Type = commwallet.Type(e.AccountType)
		a.Wallet = map[uint64]Wallet{}
	case *events.WalletOpenedMessage:
		if a.Wallet == nil {
			a.Wallet = map[uint64]Wallet{}
		}
		a.Wallet[e.WalletID] = Wallet{
			ID: e.WalletID,
			Currency: commaccount.Currency{
				ID:     e.CurrencyID,
				Symbol: e.Symbol,
			},
			Balance: big.NewInt(0),
		}
	}
	return nil
}

func (a *Account) WalletByID(id uint64) (Wallet, bool) {
	if a == nil {
		return Wallet{}, false
	}
	w, ok := a.Wallet[id]
	return w, ok
}
