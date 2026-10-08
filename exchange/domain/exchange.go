package domain

import (
	"context"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
)

type ExchangeClient interface {
	Platform() commexchange.Platform
}

type AccountExchangeClient interface {
	AccountInfo(ctx context.Context) (*AccountInfo, error)
}

type WalletExchangeClient interface {
	GetBalance(ctx context.Context, asset string) (*Balance, error)
}

type TickerExchangeClient interface {
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetCandles(ctx context.Context, symbol, interval string, limit int) ([]Candle, error)
	GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)
}

type PlaceOrderRequest struct{}
type CancelOrderRequest struct{}

type TransactionExchangeClient interface {
	PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*Order, error)
	CancelOrder(ctx context.Context, req *CancelOrderRequest) error
	GetOrder(ctx context.Context, symbol, orderID string) (*Order, error)
}
