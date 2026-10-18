package okx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

func toOKXInstID(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if s == "" || strings.Contains(s, "-") {
		return s
	}
	for _, q := range []string{"USDT", "USDC", "USD", "BTC", "ETH"} {
		if strings.HasSuffix(s, q) && len(s) > len(q) {
			return s[:len(s)-len(q)] + "-" + q
		}
	}
	return s
}

func toOKXBar(iv commexchange.Interval) (string, error) {
	if !iv.Valid() {
		return "", fmt.Errorf("invalid kline interval %q", iv)
	}
	switch iv {
	case commexchange.OneSecond:
		return "1s", nil
	case commexchange.FiveMinutes:
		return "5m", nil
	case commexchange.FifteenMinutes:
		return "15m", nil
	case commexchange.ThirtyMinutes:
		return "30m", nil
	case commexchange.OneHour:
		return "1H", nil
	case commexchange.OneDay:
		return "1D", nil
	case commexchange.OneWeek:
		return "1W", nil
	case commexchange.OneMonth:
		return "1M", nil
	case commexchange.OneYear:
		return "", fmt.Errorf("okx does not support interval %q", iv)
	default:
		return "", fmt.Errorf("invalid kline interval %q", iv)
	}
}

func toOKXCandleChannel(iv commexchange.Interval) (string, error) {
	bar, err := toOKXBar(iv)
	if err != nil {
		return "", err
	}
	return "candle" + bar, nil
}

func toOKXSide(side string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "buy", nil
	case "sell":
		return "sell", nil
	default:
		return "", fmt.Errorf("invalid order side %q", side)
	}
}

func toOKXOrdType(typ, tif string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(typ)) {
	case adapter.OrderTypeMarket:
		return "market", nil
	case adapter.OrderTypeLimit:
		switch strings.ToUpper(strings.TrimSpace(tif)) {
		case "", "GTC":
			return "limit", nil
		case "IOC":
			return "ioc", nil
		case "FOK":
			return "fok", nil
		case "POST_ONLY":
			return "post_only", nil
		default:
			return "", fmt.Errorf("invalid time in force %q", tif)
		}
	default:
		return "", fmt.Errorf("invalid order type %q", typ)
	}
}

type accountConfig struct {
	UID    string `json:"uid"`
	AcctLv string `json:"acctLv"`
	Perm   string `json:"perm"`
}

type balanceBook struct {
	Details []balanceDetail `json:"details"`
}

type balanceDetail struct {
	Ccy       string `json:"ccy"`
	AvailBal  string `json:"availBal"`
	FrozenBal string `json:"frozenBal"`
}

type ticker struct {
	InstID string `json:"instId"`
	Last   string `json:"last"`
	BidPx  string `json:"bidPx"`
	AskPx  string `json:"askPx"`
	Ts     string `json:"ts"`
}

type orderBook struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
	Ts   string     `json:"ts"`
}

type order struct {
	InstID  string `json:"instId"`
	OrdID   string `json:"ordId"`
	ClOrdID string `json:"clOrdId"`
	State   string `json:"state"`
	OrdType string `json:"ordType"`
	Side    string `json:"side"`
	Px      string `json:"px"`
	Sz      string `json:"sz"`
	AccFill string `json:"accFillSz"`
	CTime   string `json:"cTime"`
	UTime   string `json:"uTime"`
}

func toDomainAccountInfo(cfg accountConfig) *domain.AccountInfo {
	info := &domain.AccountInfo{}
	if cfg.UID != "" {
		uid := cfg.UID
		info.UID = &uid
	}
	if cfg.AcctLv != "" {
		lv := cfg.AcctLv
		info.AccountType = &lv
	}
	canTrade := strings.Contains(cfg.Perm, "trade")
	info.CanTrade = &canTrade
	if cfg.Perm != "" {
		for _, p := range strings.Split(cfg.Perm, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			perm := p
			info.Permissions = append(info.Permissions, &perm)
		}
	}
	return info
}

func toDomainBalance(details []balanceDetail, asset string) *domain.Balance {
	want := strings.ToUpper(strings.TrimSpace(asset))
	for _, d := range details {
		if want != "" && strings.ToUpper(d.Ccy) != want {
			continue
		}
		ccy, free, locked := d.Ccy, d.AvailBal, d.FrozenBal
		return &domain.Balance{Asset: &ccy, Free: &free, Locked: &locked}
	}
	return nil
}

func toDomainTicker(t ticker) *domain.Ticker {
	sym, last, bid, ask := t.InstID, t.Last, t.BidPx, t.AskPx
	out := &domain.Ticker{Symbol: &sym, LastPrice: &last, BidPrice: &bid, AskPrice: &ask}
	if ts, err := strconv.ParseInt(t.Ts, 10, 64); err == nil {
		out.Timestamp = &ts
	}
	return out
}

func toDomainCandles(rows [][]string) []domain.Candle {
	out := make([]domain.Candle, 0, len(rows))
	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		openTime, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		open, high, low, closePx, vol := row[1], row[2], row[3], row[4], row[5]
		c := domain.Candle{
			OpenTime: openTime,
			Open:     &open,
			High:     &high,
			Low:      &low,
			Close:    &closePx,
			Volume:   &vol,
		}
		if openTime > 0 {
			c.CloseTime = openTime
		}
		out = append(out, c)
	}
	return out
}

func toDomainOrderBook(symbol string, book orderBook) *domain.OrderBook {
	sym := symbol
	return &domain.OrderBook{
		Symbol: &sym,
		Bids:   toLevels(book.Bids),
		Asks:   toLevels(book.Asks),
	}
}

func toLevels(rows [][]string) []*domain.OrderBookLevel {
	out := make([]*domain.OrderBookLevel, 0, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		p, q := row[0], row[1]
		out = append(out, &domain.OrderBookLevel{Price: &p, Quantity: &q})
	}
	return out
}

func toDomainOrder(o order) *domain.Order {
	sym, id, cid := o.InstID, o.OrdID, o.ClOrdID
	status, typ, side := o.State, strings.ToUpper(o.OrdType), strings.ToUpper(o.Side)
	px, sz, filled := o.Px, o.Sz, o.AccFill
	out := &domain.Order{
		Symbol:        &sym,
		OrderID:       &id,
		ClientOrderID: &cid,
		Status:        &status,
		Type:          &typ,
		Side:          &side,
		Price:         &px,
		OrigQty:       &sz,
		ExecutedQty:   &filled,
	}
	if ts, err := strconv.ParseInt(firstNonEmpty(o.UTime, o.CTime), 10, 64); err == nil {
		out.TransactTime = &ts
	}
	return out
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseOKXStream(raw []byte) (domain.MarketUpdate, bool) {
	if string(raw) == "pong" || string(raw) == "ping" {
		return domain.MarketUpdate{}, false
	}
	var msg struct {
		Event string `json:"event"`
		Arg   struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		} `json:"arg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return domain.MarketUpdate{}, false
	}
	if msg.Event != "" && msg.Event != "update" {
		return domain.MarketUpdate{}, false
	}
	ch := msg.Arg.Channel
	switch {
	case ch == "tickers":
		var rows []ticker
		if json.Unmarshal(msg.Data, &rows) != nil || len(rows) == 0 {
			return domain.MarketUpdate{}, false
		}
		return domain.MarketUpdate{Ticker: toDomainTicker(rows[0])}, true
	case strings.HasPrefix(ch, "candle"):
		var rows [][]string
		if json.Unmarshal(msg.Data, &rows) != nil {
			return domain.MarketUpdate{}, false
		}
		cs := toDomainCandles(rows)
		if len(cs) == 0 {
			return domain.MarketUpdate{}, false
		}
		return domain.MarketUpdate{Candle: &cs[0]}, true
	case ch == "books" || ch == "books5" || ch == "books50":
		var rows []orderBook
		if json.Unmarshal(msg.Data, &rows) != nil || len(rows) == 0 {
			return domain.MarketUpdate{}, false
		}
		return domain.MarketUpdate{OrderBook: toDomainOrderBook(msg.Arg.InstID, rows[0])}, true
	default:
		return domain.MarketUpdate{}, false
	}
}
