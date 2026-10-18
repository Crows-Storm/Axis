package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

type PlaceOrder struct {
	Platform string
	Request  adapter.PlaceOrderRequest
}

type CancelOrder struct {
	Platform string
	Request  adapter.CancelOrderRequest
}

func (c PlaceOrder) Validate() decorator.CommandError {
	if c.Platform == "" {
		return decorator.InvalidCommand("platform is required")
	}
	if c.Request.Symbol == "" {
		return decorator.InvalidCommand("symbol is required")
	}
	return nil
}

func (c CancelOrder) Validate() decorator.CommandError {
	if c.Platform == "" {
		return decorator.InvalidCommand("platform is required")
	}
	if c.Request.Symbol == "" {
		return decorator.InvalidCommand("symbol is required")
	}
	if c.Request.OrderID == "" && c.Request.ClientOrderID == "" {
		return decorator.InvalidCommand("orderID or clientOrderID is required")
	}
	return nil
}

type PlaceOrderHandler decorator.CommandHandler[PlaceOrder, *domain.Order]
type CancelOrderHandler decorator.CommandHandler[CancelOrder, struct{}]

func NewPlaceOrderHandler(venues adapter.Registry, metrics decorator.MetricsClient) PlaceOrderHandler {
	return decorator.ApplyCommandDecorators[PlaceOrder, *domain.Order](placeOrderHandler{venues: venues}, metrics)
}

func NewCancelOrderHandler(venues adapter.Registry, metrics decorator.MetricsClient) CancelOrderHandler {
	return decorator.ApplyCommandDecorators[CancelOrder, struct{}](cancelOrderHandler{venues: venues}, metrics)
}

type placeOrderHandler struct{ venues adapter.Registry }
type cancelOrderHandler struct{ venues adapter.Registry }

func (h placeOrderHandler) Handle(ctx context.Context, cmd PlaceOrder) (*domain.Order, error) {
	c, err := h.venues.Client(cmd.Platform)
	if err != nil {
		return nil, err
	}
	ord, err := c.PlaceOrder(ctx, &cmd.Request)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}
	return ord, nil
}

func (h cancelOrderHandler) Handle(ctx context.Context, cmd CancelOrder) (struct{}, error) {
	c, err := h.venues.Client(cmd.Platform)
	if err != nil {
		return struct{}{}, err
	}
	if err := c.CancelOrder(ctx, &cmd.Request); err != nil {
		return struct{}{}, fmt.Errorf("cancel order: %w", err)
	}
	return struct{}{}, nil
}
