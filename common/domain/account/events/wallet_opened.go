package events

import "github.com/Crows-Storm/Axis/common/domain/event"

type WalletOpenedMessage struct {
	ID       string
	Category event.EventCategory

	AggregateID uint64
	WalletID    uint64
	CurrencyID  uint64
	Symbol      string
}

func NewWalletOpenedMessage(accountID, walletID, currencyID uint64, symbol string) *WalletOpenedMessage {
	return &WalletOpenedMessage{
		ID:          event.GenerateEventID(),
		Category:    event.Created,
		AggregateID: accountID,
		WalletID:    walletID,
		CurrencyID:  currencyID,
		Symbol:      symbol,
	}
}

func (WalletOpenedMessage) Name() string   { return "WalletOpened" }
func (WalletOpenedMessage) Domain() string { return "account" }
