package binance

import (
	"context"
	"testing"

	"github.com/Crows-Storm/Axis/common/contextx"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/stretchr/testify/require"
)

func TestClientValidation(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	_, err := c.AccountInfo(ctx)
	require.Error(t, err)

	_, err = c.GetBalance(ctx, "  ")
	require.EqualError(t, err, "asset is required")

	_, err = c.GetTicker(ctx, "")
	require.EqualError(t, err, "symbol is required")

	_, err = c.GetCandles(ctx, "BTCUSDT", commexchange.Interval("bad"), 0)
	require.Error(t, err)

	_, err = c.PlaceOrder(ctx, nil)
	require.EqualError(t, err, "place order request is required")

	_, err = c.PlaceOrder(ctx, &adapter.PlaceOrderRequest{
		Symbol: "BTCUSDT",
		Side:   adapter.OrderSideBuy,
		Type:   adapter.OrderTypeLimit,
		Market: adapter.MarketPerpetual,
	})
	require.ErrorContains(t, err, "does not support market")

	err = c.CancelOrder(ctx, &adapter.CancelOrderRequest{Symbol: "BTCUSDT"})
	require.EqualError(t, err, "orderID or clientOrderID is required")

	_, err = c.GetOrder(ctx, "BTCUSDT", "x")
	require.Error(t, err)

	err = c.CloseOutPosition(ctx, &adapter.CloseOutPositionRequest{Symbol: "BTCUSDT"})
	require.ErrorContains(t, err, "does not support close-out")
}

func TestSignedEndpointsNeedCredentials(t *testing.T) {
	c := NewClient()
	ctx := contextx.WithCredentials(context.Background(), &contextx.Credentials{
		APIKey:    "test-key",
		SecretKey: "",
	})
	_, err := c.GetBalance(ctx, "USDT")
	require.EqualError(t, err, "missing api credentials in context")
}
