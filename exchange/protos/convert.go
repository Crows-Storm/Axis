package protos

import (
	"github.com/Crows-Storm/Axis/common/genproto/exchangepb"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func toPBAccountInfo(platform string, in *domain.AccountInfo) *exchangepb.AccountInfo {
	if in == nil {
		return &exchangepb.AccountInfo{Platform: platform}
	}
	out := &exchangepb.AccountInfo{
		Platform:    platform,
		Uid:         deref(in.UID),
		AccountType: deref(in.AccountType),
		CanTrade:    in.CanTrade != nil && *in.CanTrade,
	}
	for _, p := range in.Permissions {
		if p != nil && *p != "" {
			out.Permissions = append(out.Permissions, *p)
		}
	}
	return out
}

func toPBBalance(in *domain.Balance) *exchangepb.Balance {
	if in == nil {
		return &exchangepb.Balance{}
	}
	return &exchangepb.Balance{Asset: deref(in.Asset), Free: deref(in.Free), Locked: deref(in.Locked)}
}

func toPBTicker(in *domain.Ticker) *exchangepb.Ticker {
	if in == nil {
		return &exchangepb.Ticker{}
	}
	return &exchangepb.Ticker{
		Symbol:    deref(in.Symbol),
		LastPrice: deref(in.LastPrice),
		BidPrice:  deref(in.BidPrice),
		AskPrice:  deref(in.AskPrice),
		Timestamp: derefInt(in.Timestamp),
	}
}

func toPBCandles(rows []domain.Candle) []*exchangepb.Candle {
	out := make([]*exchangepb.Candle, 0, len(rows))
	for _, c := range rows {
		out = append(out, &exchangepb.Candle{
			OpenTime:  c.OpenTime,
			Open:      deref(c.Open),
			High:      deref(c.High),
			Low:       deref(c.Low),
			Close:     deref(c.Close),
			Volume:    deref(c.Volume),
			CloseTime: c.CloseTime,
		})
	}
	return out
}

func toPBOrderBook(in *domain.OrderBook) *exchangepb.OrderBook {
	if in == nil {
		return &exchangepb.OrderBook{}
	}
	return &exchangepb.OrderBook{
		Symbol: deref(in.Symbol),
		Bids:   toPBLevels(in.Bids),
		Asks:   toPBLevels(in.Asks),
	}
}

func toPBLevels(rows []*domain.OrderBookLevel) []*exchangepb.OrderBookLevel {
	out := make([]*exchangepb.OrderBookLevel, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &exchangepb.OrderBookLevel{Price: deref(row.Price), Quantity: deref(row.Quantity)})
	}
	return out
}

func toPBOrder(in *domain.Order) *exchangepb.Order {
	if in == nil {
		return &exchangepb.Order{}
	}
	return &exchangepb.Order{
		Symbol:        deref(in.Symbol),
		OrderId:       deref(in.OrderID),
		ClientOrderId: deref(in.ClientOrderID),
		Status:        deref(in.Status),
		Type:          deref(in.Type),
		Side:          deref(in.Side),
		Price:         deref(in.Price),
		OrigQty:       deref(in.OrigQty),
		ExecutedQty:   deref(in.ExecutedQty),
		TimeInForce:   deref(in.TimeInForce),
		TransactTime:  derefInt(in.TransactTime),
	}
}
