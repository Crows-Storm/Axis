package adapter

import (
	"context"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func HTTPToWS(raw string) string {
	switch {
	case strings.HasPrefix(raw, "https://"):
		return "wss://" + strings.TrimPrefix(raw, "https://")
	case strings.HasPrefix(raw, "http://"):
		return "ws://" + strings.TrimPrefix(raw, "http://")
	default:
		return raw
	}
}

func DialWebsocket(ctx context.Context, rawURL string) (*websocket.Conn, error) {
	d := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, resp, err := d.DialContext(ctx, rawURL, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func WritePingLoop(ctx context.Context, conn *websocket.Conn, payload []byte, every time.Duration) {
	if every <= 0 {
		every = 20 * time.Second
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}
