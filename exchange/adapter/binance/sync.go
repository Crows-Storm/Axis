package binance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/gorilla/websocket"
)

func (c *Client) Start(ctx context.Context, req adapter.SyncMarketDataRequest) (<-chan domain.MarketUpdate, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	streams, err := binanceStreams(req)
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

	url := strings.TrimRight(c.wsURL, "/") + "/stream?streams=" + strings.Join(streams, "/")
	conn, err := adapter.DialWebsocket(runCtx, url)
	if err != nil {
		c.Stop()
		return nil, fmt.Errorf("binance market ws: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	out := make(chan domain.MarketUpdate, 64)
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
		upd, ok := parseBinanceStream(raw)
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

func binanceStreams(req adapter.SyncMarketDataRequest) ([]string, error) {
	var streams []string
	for _, symbol := range req.Symbols {
		s := strings.ToLower(strings.TrimSpace(symbol))
		if req.HasChannel(domain.ChannelTicker) {
			streams = append(streams, s+"@ticker")
		}
		if req.HasChannel(domain.ChannelCandle) {
			iv, err := toBinanceInterval(req.Interval)
			if err != nil {
				return nil, err
			}
			streams = append(streams, s+"@kline_"+iv)
		}
		if req.HasChannel(domain.ChannelBook) {
			depth := req.Depth
			if depth <= 0 {
				depth = 20
			}
			streams = append(streams, fmt.Sprintf("%s@depth%d", s, depth))
		}
	}
	return streams, nil
}
