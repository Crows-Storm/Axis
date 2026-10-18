package okx

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
	args, err := okxSubscribeArgs(req)
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
		return nil, fmt.Errorf("okx market ws: %w", err)
	}
	if err := conn.WriteJSON(map[string]any{"op": "subscribe", "args": args}); err != nil {
		_ = conn.Close()
		c.Stop()
		return nil, fmt.Errorf("okx market subscribe: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	out := make(chan domain.MarketUpdate, 64)
	go adapter.WritePingLoop(runCtx, conn, []byte("ping"), 20*time.Second)
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
		upd, ok := parseOKXStream(raw)
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

func okxSubscribeArgs(req adapter.SyncMarketDataRequest) ([]map[string]string, error) {
	var args []map[string]string
	var candleCh string
	var err error
	if req.HasChannel(domain.ChannelCandle) {
		candleCh, err = toOKXCandleChannel(req.Interval)
		if err != nil {
			return nil, err
		}
	}
	for _, symbol := range req.Symbols {
		instID, err := requireInstID(symbol)
		if err != nil {
			return nil, err
		}
		if req.HasChannel(domain.ChannelTicker) {
			args = append(args, map[string]string{"channel": "tickers", "instId": instID})
		}
		if req.HasChannel(domain.ChannelCandle) {
			args = append(args, map[string]string{"channel": candleCh, "instId": instID})
		}
		if req.HasChannel(domain.ChannelBook) {
			ch := "books"
			if req.Depth == 5 {
				ch = "books5"
			}
			args = append(args, map[string]string{"channel": ch, "instId": instID})
		}
	}
	return args, nil
}
