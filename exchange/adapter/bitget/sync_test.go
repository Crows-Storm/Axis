package bitget

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestStartValidation(t *testing.T) {
	c := NewClient()
	_, err := c.Start(context.Background(), adapter.SyncMarketDataRequest{})
	require.Error(t, err)
}

func TestStartTickerWS(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(
			`{"arg":{"channel":"ticker","instId":"BTCUSDT"},"data":[{"instId":"BTCUSDT","lastPr":"100","bidPr":"99","askPr":"101","ts":"9"}]}`,
		))
		time.Sleep(200 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	c.wsURL = "ws" + srv.URL[len("http"):]
	out, err := c.Start(context.Background(), adapter.SyncMarketDataRequest{
		Symbols:  []string{"BTCUSDT"},
		Channels: []domain.MarketChannel{domain.ChannelTicker},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Stop() })

	select {
	case u := <-out:
		require.Equal(t, "100", *u.Ticker.LastPrice)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for ticker")
	}
}
