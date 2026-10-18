package binance

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/domain"
	spotmodels "github.com/binance/binance-connector-go/clients/spot/src/restapi/models"
	walletmodels "github.com/binance/binance-connector-go/clients/wallet/src/restapi/models"
)

func toBinanceInterval(iv commexchange.Interval) (string, error) {
	if !iv.Valid() || iv == commexchange.OneYear {
		return "", fmt.Errorf("unsupported interval %q", iv)
	}
	return string(iv), nil
}

func toDomainAccountInfo(resp spotmodels.GetAccountResponse) *domain.AccountInfo {
	info := &domain.AccountInfo{
		AccountType: resp.AccountType,
		CanTrade:    resp.CanTrade,
	}
	if resp.Uid != nil {
		uid := strconv.FormatInt(*resp.Uid, 10)
		info.UID = &uid
	}
	if len(resp.Permissions) > 0 {
		info.Permissions = make([]*string, 0, len(resp.Permissions))
		for i := range resp.Permissions {
			p := resp.Permissions[i]
			info.Permissions = append(info.Permissions, &p)
		}
	}
	return info
}

func toDomainBalanceFromUserAsset(resp walletmodels.UserAssetResponse, asset string) *domain.Balance {
	want := strings.ToUpper(strings.TrimSpace(asset))
	for _, item := range resp.Items {
		if item.Asset == nil {
			continue
		}
		if want != "" && strings.ToUpper(*item.Asset) != want {
			continue
		}
		return &domain.Balance{
			Asset:  item.Asset,
			Free:   item.Free,
			Locked: item.Locked,
		}
	}
	return nil
}

func toDomainTicker(resp spotmodels.Ticker24hrResponse) *domain.Ticker {
	src := resp.Ticker24hrResponse1
	if src == nil {
		return nil
	}
	ts := src.CloseTime
	return &domain.Ticker{
		Symbol:    src.Symbol,
		LastPrice: src.LastPrice,
		BidPrice:  src.BidPrice,
		AskPrice:  src.AskPrice,
		Timestamp: ts,
	}
}

func toDomainCandles(resp spotmodels.KlinesResponse) []domain.Candle {
	out := make([]domain.Candle, 0, len(resp.Items))
	for _, item := range resp.Items {
		c, ok := klineToCandle(item.Items)
		if !ok {
			continue
		}
		out = append(out, c)
	}
	return out
}

func klineToCandle(items []spotmodels.KlinesItemInner) (domain.Candle, bool) {
	if len(items) < 7 {
		return domain.Candle{}, false
	}
	openTime, ok := klineInt64(items[0])
	if !ok {
		return domain.Candle{}, false
	}
	closeTime, ok := klineInt64(items[6])
	if !ok {
		return domain.Candle{}, false
	}
	return domain.Candle{
		OpenTime:  openTime,
		Open:      klineString(items[1]),
		High:      klineString(items[2]),
		Low:       klineString(items[3]),
		Close:     klineString(items[4]),
		Volume:    klineString(items[5]),
		CloseTime: closeTime,
	}, true
}

func klineInt64(item spotmodels.KlinesItemInner) (int64, bool) {
	if item.Int64 != nil {
		return *item.Int64, true
	}
	if item.String != nil && *item.String != "" {
		n, err := strconv.ParseInt(*item.String, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func klineString(item spotmodels.KlinesItemInner) *string {
	if item.String != nil {
		return item.String
	}
	if item.Int64 != nil {
		s := strconv.FormatInt(*item.Int64, 10)
		return &s
	}
	return nil
}

func toDomainOrderBook(symbol string, resp spotmodels.DepthResponse) *domain.OrderBook {
	sym := symbol
	book := &domain.OrderBook{
		Symbol: &sym,
		Bids:   toOrderBookLevels(resp.Bids),
		Asks:   toOrderBookLevels(resp.Asks),
	}
	return book
}

func toOrderBookLevels(rows [][]string) []*domain.OrderBookLevel {
	out := make([]*domain.OrderBookLevel, 0, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		price := row[0]
		qty := row[1]
		out = append(out, &domain.OrderBookLevel{
			Price:    &price,
			Quantity: &qty,
		})
	}
	return out
}

func toDomainOrderFromNew(resp spotmodels.NewOrderResponse) *domain.Order {
	return &domain.Order{
		Symbol:        resp.Symbol,
		OrderID:       int64PtrToString(resp.OrderId),
		ClientOrderID: resp.ClientOrderId,
		Status:        resp.Status,
		Type:          resp.Type,
		Side:          resp.Side,
		Price:         resp.Price,
		OrigQty:       resp.OrigQty,
		ExecutedQty:   resp.ExecutedQty,
		TimeInForce:   resp.TimeInForce,
		TransactTime:  resp.TransactTime,
	}
}

func toDomainOrderFromGet(resp spotmodels.GetOrderResponse) *domain.Order {
	return &domain.Order{
		Symbol:        resp.Symbol,
		OrderID:       int64PtrToString(resp.OrderId),
		ClientOrderID: resp.ClientOrderId,
		Status:        resp.Status,
		Type:          resp.Type,
		Side:          resp.Side,
		Price:         resp.Price,
		OrigQty:       resp.OrigQty,
		ExecutedQty:   resp.ExecutedQty,
		TimeInForce:   resp.TimeInForce,
		TransactTime:  resp.Time,
	}
}

func int64PtrToString(v *int64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}

func parseBinanceStream(raw []byte) (domain.MarketUpdate, bool) {
	var wrap struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}
	payload := raw
	stream := ""
	if json.Unmarshal(raw, &wrap) == nil && wrap.Stream != "" && len(wrap.Data) > 0 {
		payload = wrap.Data
		stream = wrap.Stream
	}
	ev := eventName(payload)
	switch {
	case strings.Contains(stream, "@kline_") || ev == "kline":
		return parseBinanceKline(payload)
	case strings.Contains(stream, "@ticker") || ev == "24hrTicker" || ev == "24hrMiniTicker":
		return parseBinanceTicker(payload)
	case strings.Contains(stream, "@depth") || ev == "depthUpdate":
		return parseBinanceDepth(payload, stream)
	default:
		return domain.MarketUpdate{}, false
	}
}

func eventName(raw []byte) string {
	var head struct {
		Event string `json:"e"`
	}
	_ = json.Unmarshal(raw, &head)
	return head.Event
}

func parseBinanceTicker(raw []byte) (domain.MarketUpdate, bool) {
	var src struct {
		Symbol    string `json:"s"`
		LastPrice string `json:"c"`
		BidPrice  string `json:"b"`
		AskPrice  string `json:"a"`
	}
	if err := json.Unmarshal(raw, &src); err != nil || src.Symbol == "" {
		return domain.MarketUpdate{}, false
	}
	sym, last, bid, ask := src.Symbol, src.LastPrice, src.BidPrice, src.AskPrice
	tk := &domain.Ticker{Symbol: &sym, LastPrice: &last, BidPrice: &bid, AskPrice: &ask}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) == nil {
		if v, ok := fields["E"]; ok {
			var ts int64
			if json.Unmarshal(v, &ts) == nil {
				tk.Timestamp = &ts
			}
		}
	}
	return domain.MarketUpdate{Ticker: tk}, true
}

func parseBinanceKline(raw []byte) (domain.MarketUpdate, bool) {
	var src struct {
		K struct {
			OpenTime  int64  `json:"t"`
			CloseTime int64  `json:"T"`
			Open      string `json:"o"`
			High      string `json:"h"`
			Low       string `json:"l"`
			Close     string `json:"c"`
			Volume    string `json:"v"`
		} `json:"k"`
	}
	if err := json.Unmarshal(raw, &src); err != nil || src.K.Open == "" {
		return domain.MarketUpdate{}, false
	}
	o, h, l, c, v := src.K.Open, src.K.High, src.K.Low, src.K.Close, src.K.Volume
	return domain.MarketUpdate{Candle: &domain.Candle{
		OpenTime: src.K.OpenTime, CloseTime: src.K.CloseTime,
		Open: &o, High: &h, Low: &l, Close: &c, Volume: &v,
	}}, true
}

func parseBinanceDepth(raw []byte, stream string) (domain.MarketUpdate, bool) {
	var src struct {
		Symbol string     `json:"s"`
		Bids   [][]string `json:"b"`
		Asks   [][]string `json:"a"`
		Bids2  [][]string `json:"bids"`
		Asks2  [][]string `json:"asks"`
	}
	if err := json.Unmarshal(raw, &src); err != nil {
		return domain.MarketUpdate{}, false
	}
	bids, asks := src.Bids, src.Asks
	if len(bids) == 0 {
		bids = src.Bids2
	}
	if len(asks) == 0 {
		asks = src.Asks2
	}
	sym := src.Symbol
	if sym == "" {
		sym = strings.ToUpper(strings.Split(stream, "@")[0])
	}
	return domain.MarketUpdate{OrderBook: toDomainOrderBook(sym, spotmodels.DepthResponse{Bids: bids, Asks: asks})}, true
}
