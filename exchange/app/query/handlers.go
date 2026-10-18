package query

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

type GetAccountInfoQuery struct {
	Platform string
}

type GetBalanceQuery struct {
	Platform string
	Asset    string
}

type GetTickerQuery struct {
	Platform string
	Symbol   string
}

type GetCandlesQuery struct {
	Platform string
	Symbol   string
	Interval commexchange.Interval
	Limit    int
}

type GetOrderBookQuery struct {
	Platform string
	Symbol   string
	Depth    int
}

type GetOrderQuery struct {
	Platform string
	Symbol   string
	OrderID  string
}

type GetAccountInfoHandler decorator.QueryHandler[GetAccountInfoQuery, *domain.AccountInfo]
type GetBalanceHandler decorator.QueryHandler[GetBalanceQuery, *domain.Balance]
type GetTickerHandler decorator.QueryHandler[GetTickerQuery, *domain.Ticker]
type GetCandlesHandler decorator.QueryHandler[GetCandlesQuery, []domain.Candle]
type GetOrderBookHandler decorator.QueryHandler[GetOrderBookQuery, *domain.OrderBook]
type GetOrderHandler decorator.QueryHandler[GetOrderQuery, *domain.Order]

func NewGetAccountInfoHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetAccountInfoHandler {
	return decorator.ApplyQueryDecorators[GetAccountInfoQuery, *domain.AccountInfo](accountInfoHandler{venues: venues}, metrics)
}

func NewGetBalanceHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetBalanceHandler {
	return decorator.ApplyQueryDecorators[GetBalanceQuery, *domain.Balance](balanceHandler{venues: venues}, metrics)
}

func NewGetTickerHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetTickerHandler {
	return decorator.ApplyQueryDecorators[GetTickerQuery, *domain.Ticker](tickerHandler{venues: venues}, metrics)
}

func NewGetCandlesHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetCandlesHandler {
	return decorator.ApplyQueryDecorators[GetCandlesQuery, []domain.Candle](candlesHandler{venues: venues}, metrics)
}

func NewGetOrderBookHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetOrderBookHandler {
	return decorator.ApplyQueryDecorators[GetOrderBookQuery, *domain.OrderBook](orderBookHandler{venues: venues}, metrics)
}

func NewGetOrderHandler(venues adapter.Registry, metrics decorator.MetricsClient) GetOrderHandler {
	return decorator.ApplyQueryDecorators[GetOrderQuery, *domain.Order](orderHandler{venues: venues}, metrics)
}

type accountInfoHandler struct{ venues adapter.Registry }
type balanceHandler struct{ venues adapter.Registry }
type tickerHandler struct{ venues adapter.Registry }
type candlesHandler struct{ venues adapter.Registry }
type orderBookHandler struct{ venues adapter.Registry }
type orderHandler struct{ venues adapter.Registry }

func (h accountInfoHandler) Handle(ctx context.Context, q GetAccountInfoQuery) (*domain.AccountInfo, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	info, err := c.AccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("account info: %w", err)
	}
	return info, nil
}

func (h balanceHandler) Handle(ctx context.Context, q GetBalanceQuery) (*domain.Balance, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	bal, err := c.GetBalance(ctx, q.Asset)
	if err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}
	return bal, nil
}

func (h tickerHandler) Handle(ctx context.Context, q GetTickerQuery) (*domain.Ticker, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	tk, err := c.GetTicker(ctx, q.Symbol)
	if err != nil {
		return nil, fmt.Errorf("ticker: %w", err)
	}
	return tk, nil
}

func (h candlesHandler) Handle(ctx context.Context, q GetCandlesQuery) ([]domain.Candle, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	rows, err := c.GetCandles(ctx, q.Symbol, q.Interval, q.Limit)
	if err != nil {
		return nil, fmt.Errorf("candles: %w", err)
	}
	return rows, nil
}

func (h orderBookHandler) Handle(ctx context.Context, q GetOrderBookQuery) (*domain.OrderBook, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	book, err := c.GetOrderBook(ctx, q.Symbol, q.Depth)
	if err != nil {
		return nil, fmt.Errorf("order book: %w", err)
	}
	return book, nil
}

func (h orderHandler) Handle(ctx context.Context, q GetOrderQuery) (*domain.Order, error) {
	c, err := h.venues.Client(q.Platform)
	if err != nil {
		return nil, err
	}
	ord, err := c.GetOrder(ctx, q.Symbol, q.OrderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return ord, nil
}
