package domain

type Candle struct {
	OpenTime  int64
	Open      *string
	High      *string
	Low       *string
	Close     *string
	Volume    *string
	CloseTime int64
}

type OrderBook struct {
	Symbol *string
	Bids   []*OrderBookLevel
	Asks   []*OrderBookLevel
}

type OrderBookLevel struct {
	Price    *string
	Quantity *string
}

type MarketChannel string

const (
	ChannelTicker MarketChannel = "ticker"
	ChannelCandle MarketChannel = "candle"
	ChannelBook   MarketChannel = "orderbook"
)

type MarketUpdate struct {
	Ticker    *Ticker
	Candle    *Candle
	OrderBook *OrderBook
}
