package adapter

import (
	"context"
	"errors"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

var (
	errSymbolsRequired  = errors.New("symbol is required")
	errIntervalRequired = errors.New("interval is required")
)

type ExchangeClient interface {
	Platform() commexchange.Platform

	AccountExchangeClient
	WalletExchangeClient
	TickerExchangeClient
	TransactionExchangeClient
}

type AccountExchangeClient interface {
	AccountInfo(ctx context.Context) (*domain.AccountInfo, error)
}

type WalletExchangeClient interface {
	GetBalance(ctx context.Context, asset string) (*domain.Balance, error)
}

type TickerExchangeClient interface {
	GetTicker(ctx context.Context, symbol string) (*domain.Ticker, error)
	GetCandles(ctx context.Context, symbol string, interval commexchange.Interval, limit int) ([]domain.Candle, error)
	GetOrderBook(ctx context.Context, symbol string, depth int) (*domain.OrderBook, error)
}

type SyncMarketDataRequest struct {
	Symbols  []string
	Interval commexchange.Interval
	Depth    int
	Channels []domain.MarketChannel
}

func (r SyncMarketDataRequest) ChannelSet() []domain.MarketChannel {
	if len(r.Channels) == 0 {
		return []domain.MarketChannel{domain.ChannelTicker, domain.ChannelCandle, domain.ChannelBook}
	}
	return r.Channels
}

func (r SyncMarketDataRequest) HasChannel(ch domain.MarketChannel) bool {
	for _, c := range r.ChannelSet() {
		if c == ch {
			return true
		}
	}
	return false
}

func (r SyncMarketDataRequest) Validate() error {
	if len(r.Symbols) == 0 {
		return errSymbolsRequired
	}
	for _, s := range r.Symbols {
		if s == "" {
			return errSymbolsRequired
		}
	}
	if r.HasChannel(domain.ChannelCandle) && !r.Interval.Valid() {
		return errIntervalRequired
	}
	return nil
}

// SyncMarketData pulls live ticker / candle / order book over websocket.
type SyncMarketData interface {
	Start(ctx context.Context, req SyncMarketDataRequest) (<-chan domain.MarketUpdate, error)
	Stop() error
}

// MarketSpot / MarketPerpetual identify the venue product. Empty Market on
// PlaceOrderRequest is treated as spot.
const (
	MarketSpot      = "spot"
	MarketPerpetual = "perpetual"
	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
	OrderSideBuy    = "BUY"
	OrderSideSell   = "SELL"
)

type PlaceOrderRequest struct {
	Symbol        string
	Side          string
	Type          string
	Market        string
	Quantity      string
	QuoteQuantity string
	Price         string
	TimeInForce   string
	ClientOrderID string
	StopPrice     string
}

type CancelOrderRequest struct {
	Symbol        string
	OrderID       string
	ClientOrderID string
}

type TransactionExchangeClient interface {
	PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*domain.Order, error)
	CancelOrder(ctx context.Context, req *CancelOrderRequest) error
	GetOrder(ctx context.Context, symbol, orderID string) (*domain.Order, error)
}
