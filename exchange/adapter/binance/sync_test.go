package binance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestStartValidation(t *testing.T) {
	c := NewClient()
	_, err := c.Start(context.Background(), adapter.SyncMarketDataRequest{})
	require.Error(t, err)
	_, err = c.Start(context.Background(), adapter.SyncMarketDataRequest{
		Symbols: []string{"BTCUSDT"}, Interval: commexchange.OneYear,
	})
	require.Error(t, err)
}

func TestStartTickerWS(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(
			`{"stream":"btcusdt@ticker","data":{"e":"24hrTicker","s":"BTCUSDT","c":"100","b":"99","a":"101","E":9}}`,
		))
		time.Sleep(200 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	c.wsURL = adapter.HTTPToWS(srv.URL)
	out, err := c.Start(context.Background(), adapter.SyncMarketDataRequest{
		Symbols:  []string{"BTCUSDT"},
		Channels: []domain.MarketChannel{domain.ChannelTicker},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Stop() })

	select {
	case u := <-out:
		require.NotNil(t, u.Ticker)
		require.Equal(t, "100", *u.Ticker.LastPrice)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for ticker")
	}
}
