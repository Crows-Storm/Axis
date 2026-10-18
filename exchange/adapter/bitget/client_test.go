package bitget

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
		require.Equal(t, "/api/v2/spot/market/tickers", r.URL.Path)
		require.Equal(t, "BTCUSDT", r.URL.Query().Get("symbol"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "00000",
			"data": []map[string]string{{
				"symbol": "BTCUSDT",
				"lastPr": "100",
				"bidPr":  "99",
				"askPr":  "101",
				"ts":     "9",
			}},
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	c.baseURL = srv.URL
	got, err := c.GetTicker(context.Background(), "btc-usdt")
	require.NoError(t, err)
	require.Equal(t, "100", *got.LastPrice)
}

func TestSignBitgetStable(t *testing.T) {
	a := signBitget("secret", "1690000000000", http.MethodGet, "/api/v2/spot/account/assets?coin=USDT", nil)
	b := signBitget("secret", "1690000000000", http.MethodGet, "/api/v2/spot/account/assets?coin=USDT", nil)
	require.Equal(t, a, b)
	require.NotEmpty(t, a)
}
