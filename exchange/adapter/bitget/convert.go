package bitget

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

func toBitgetSymbol(symbol string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(symbol), "-", ""))
}

func toBitgetGranularity(iv commexchange.Interval) (string, error) {
	if !iv.Valid() {
		return "", fmt.Errorf("invalid kline interval %q", iv)
	}
	switch iv {
	case commexchange.FiveMinutes:
		return "5min", nil
	case commexchange.FifteenMinutes:
		return "15min", nil
	case commexchange.ThirtyMinutes:
		return "30min", nil
	case commexchange.OneHour:
		return "1h", nil
	case commexchange.OneDay:
		return "1day", nil
	case commexchange.OneWeek:
		return "1week", nil
	case commexchange.OneMonth:
		return "1M", nil
	default:
		return "", fmt.Errorf("bitget does not support interval %q", iv)
	}
}

func toBitgetCandleChannel(iv commexchange.Interval) (string, error) {
	if !iv.Valid() {
		return "", fmt.Errorf("invalid kline interval %q", iv)
	}
	switch iv {
	case commexchange.FiveMinutes:
		return "candle5m", nil
	case commexchange.FifteenMinutes:
		return "candle15m", nil
	case commexchange.ThirtyMinutes:
		return "candle30m", nil
	case commexchange.OneHour:
		return "candle1H", nil
	case commexchange.OneDay:
		return "candle1D", nil
	case commexchange.OneWeek:
		return "candle1W", nil
	case commexchange.OneMonth:
		return "candle1M", nil
	default:
		return "", fmt.Errorf("bitget does not support interval %q", iv)
	}
}

func toBitgetSide(side string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "buy", nil
	case "sell":
		return "sell", nil
	default:
		return "", fmt.Errorf("invalid order side %q", side)
	}
}

func toBitgetOrderType(typ string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(typ)) {
	case adapter.OrderTypeMarket:
		return "market", nil
	case adapter.OrderTypeLimit:
		return "limit", nil
	default:
		return "", fmt.Errorf("invalid order type %q", typ)
	}
}

func toBitgetForce(tif string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(tif)) {
	case "", "GTC":
		return "gtc", nil
	case "IOC":
		return "ioc", nil
	case "FOK":
		return "fok", nil
	case "POST_ONLY":
		return "post_only", nil
	default:
		return "", fmt.Errorf("invalid time in force %q", tif)
	}
}

type accountInfo struct {
	UserID      string   `json:"userId"`
	Authorities []string `json:"authorities"`
}

type spotAsset struct {
	Coin      string `json:"coin"`
	Available string `json:"available"`
	Frozen    string `json:"frozen"`
	Locked    string `json:"locked"`
}

type ticker struct {
	Symbol string `json:"symbol"`
	LastPr string `json:"lastPr"`
	BidPr  string `json:"bidPr"`
	AskPr  string `json:"askPr"`
	Ts     string `json:"ts"`
}

type orderBook struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
	Ts   string     `json:"ts"`
}

type order struct {
	Symbol     string `json:"symbol"`
	OrderID    string `json:"orderId"`
	ClientOid  string `json:"clientOid"`
	Status     string `json:"status"`
	OrderType  string `json:"orderType"`
	Side       string `json:"side"`
	Price      string `json:"price"`
	Size       string `json:"size"`
	BaseVolume string `json:"baseVolume"`
	Force      string `json:"force"`
	CTime      string `json:"cTime"`
	UTime      string `json:"uTime"`
}

func toDomainAccountInfo(info accountInfo) *domain.AccountInfo {
	out := &domain.AccountInfo{}
	if info.UserID != "" {
		uid := info.UserID
		out.UID = &uid
	}
	acct := "spot"
	out.AccountType = &acct
	canTrade := false
	for _, a := range info.Authorities {
		p := a
		out.Permissions = append(out.Permissions, &p)
		if strings.EqualFold(a, "spot") || strings.EqualFold(a, "trade") {
			canTrade = true
		}
	}
	out.CanTrade = &canTrade
	return out
}

func toDomainBalance(assets []spotAsset, coin string) *domain.Balance {
	want := strings.ToUpper(strings.TrimSpace(coin))
	for _, a := range assets {
		if want != "" && strings.ToUpper(a.Coin) != want {
			continue
		}
		ccy := a.Coin
		free := a.Available
		locked := a.Frozen
		if a.Locked != "" && a.Locked != "0" {
			locked = a.Locked
		}
		return &domain.Balance{Asset: &ccy, Free: &free, Locked: &locked}
	}
	return nil
}

func toDomainTicker(t ticker) *domain.Ticker {
	sym, last, bid, ask := t.Symbol, t.LastPr, t.BidPr, t.AskPr
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
		out = append(out, domain.Candle{
			OpenTime:  openTime,
			Open:      &open,
			High:      &high,
			Low:       &low,
			Close:     &closePx,
			Volume:    &vol,
			CloseTime: openTime,
		})
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
	sym, id, cid := o.Symbol, o.OrderID, o.ClientOid
	status, typ, side := o.Status, strings.ToUpper(o.OrderType), strings.ToUpper(o.Side)
	px, sz := o.Price, o.Size
	filled := o.BaseVolume
	tif := strings.ToUpper(o.Force)
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
		TimeInForce:   &tif,
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

func parseBitgetStream(raw []byte) (domain.MarketUpdate, bool) {
	var msg struct {
		Event string `json:"event"`
		Arg   struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		} `json:"arg"`
		Data json.RawMessage `json:"data"`
		Op   string          `json:"op"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return domain.MarketUpdate{}, false
	}
	if msg.Op == "pong" || msg.Op == "ping" || (msg.Event != "" && msg.Event != "update") {
		return domain.MarketUpdate{}, false
	}
	ch := msg.Arg.Channel
	switch {
	case ch == "ticker":
		var rows []ticker
		if json.Unmarshal(msg.Data, &rows) != nil || len(rows) == 0 {
			var alt []struct {
				InstID string `json:"instId"`
				LastPr string `json:"lastPr"`
				BidPr  string `json:"bidPr"`
				AskPr  string `json:"askPr"`
				Ts     string `json:"ts"`
			}
			if json.Unmarshal(msg.Data, &alt) != nil || len(alt) == 0 {
				return domain.MarketUpdate{}, false
			}
			rows = []ticker{{Symbol: alt[0].InstID, LastPr: alt[0].LastPr, BidPr: alt[0].BidPr, AskPr: alt[0].AskPr, Ts: alt[0].Ts}}
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
	case ch == "books" || ch == "books5":
		var rows []orderBook
		if json.Unmarshal(msg.Data, &rows) != nil || len(rows) == 0 {
			return domain.MarketUpdate{}, false
		}
		return domain.MarketUpdate{OrderBook: toDomainOrderBook(msg.Arg.InstID, rows[0])}, true
	default:
		return domain.MarketUpdate{}, false
	}
}
