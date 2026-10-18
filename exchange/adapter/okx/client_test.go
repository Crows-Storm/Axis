package okx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	_, err = c.GetOrder(ctx, "BTCUSDT", "")
	require.EqualError(t, err, "orderID is required")

	err = c.CloseOutPosition(ctx, &adapter.CloseOutPositionRequest{Symbol: "BTCUSDT"})
	require.ErrorContains(t, err, "does not support close-out")
}

func TestSignedEndpointsNeedCredentials(t *testing.T) {
	c := NewClient()
	ctx := contextx.WithCredentials(context.Background(), &contextx.Credentials{
		APIKey:    "test-key",
		SecretKey: "secret",
	})
	_, err := c.GetBalance(ctx, "USDT")
	require.ErrorContains(t, err, "missing api credentials in context")
}

func TestGetTickerHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v5/market/ticker", r.URL.Path)
		require.Equal(t, "BTC-USDT", r.URL.Query().Get("instId"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "0",
			"data": []map[string]string{{
				"instId": "BTC-USDT",
				"last":   "100",
				"bidPx":  "99",
				"askPx":  "101",
				"ts":     "9",
			}},
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	c.baseURL = srv.URL
	got, err := c.GetTicker(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	require.Equal(t, "100", *got.LastPrice)
	require.Equal(t, int64(9), *got.Timestamp)
}

func TestSignOKXStable(t *testing.T) {
	a := signOKX("secret", "2020-12-08T09:08:57.715Z", http.MethodGet, "/api/v5/account/balance?ccy=BTC", nil)
	b := signOKX("secret", "2020-12-08T09:08:57.715Z", http.MethodGet, "/api/v5/account/balance?ccy=BTC", nil)
	require.Equal(t, a, b)
	require.NotEmpty(t, a)
}
