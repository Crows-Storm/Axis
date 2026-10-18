package bitget

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/gorilla/websocket"
)

func (c *Client) Start(ctx context.Context, req adapter.SyncMarketDataRequest) (<-chan domain.MarketUpdate, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	args, err := bitgetSubscribeArgs(req)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		return nil, errors.New("market sync already started")
	}
	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.mu.Unlock()

	conn, err := adapter.DialWebsocket(runCtx, c.wsURL)
	if err != nil {
		c.Stop()
		return nil, fmt.Errorf("bitget market ws: %w", err)
	}
	if err := conn.WriteJSON(map[string]any{"op": "subscribe", "args": args}); err != nil {
		_ = conn.Close()
		c.Stop()
		return nil, fmt.Errorf("bitget market subscribe: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	out := make(chan domain.MarketUpdate, 64)
	go adapter.WritePingLoop(runCtx, conn, []byte(`{"op":"ping"}`), 30*time.Second)
	go c.readLoop(runCtx, conn, out)
	return out, nil
}

func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	return nil
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn, out chan<- domain.MarketUpdate) {
	defer close(out)
	defer func() { _ = c.Stop() }()
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		upd, ok := parseBitgetStream(raw)
		if !ok {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case out <- upd:
		}
	}
}

func bitgetSubscribeArgs(req adapter.SyncMarketDataRequest) ([]map[string]string, error) {
	var args []map[string]string
	var candleCh string
	var err error
	if req.HasChannel(domain.ChannelCandle) {
		candleCh, err = toBitgetCandleChannel(req.Interval)
		if err != nil {
			return nil, err
		}
	}
	for _, symbol := range req.Symbols {
		instID, err := requireSymbol(symbol)
		if err != nil {
			return nil, err
		}
		if req.HasChannel(domain.ChannelTicker) {
			args = append(args, map[string]string{"instType": "SPOT", "channel": "ticker", "instId": instID})
		}
		if req.HasChannel(domain.ChannelCandle) {
			args = append(args, map[string]string{"instType": "SPOT", "channel": candleCh, "instId": instID})
		}
		if req.HasChannel(domain.ChannelBook) {
			args = append(args, map[string]string{"instType": "SPOT", "channel": "books", "instId": instID})
		}
	}
	return args, nil
}
