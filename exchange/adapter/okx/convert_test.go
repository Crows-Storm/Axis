package okx

import (
	"testing"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/stretchr/testify/require"
)

func TestToOKXInstIDAndBar(t *testing.T) {
	require.Equal(t, "BTC-USDT", toOKXInstID("btcusdt"))
	require.Equal(t, "BTC-USDT", toOKXInstID("BTC-USDT"))
	bar, err := toOKXBar(commexchange.OneHour)
	require.NoError(t, err)
	require.Equal(t, "1H", bar)
	_, err = toOKXBar(commexchange.Interval("bad"))
	require.Error(t, err)
	ch, err := toOKXCandleChannel(commexchange.OneHour)
	require.NoError(t, err)
	require.Equal(t, "candle1H", ch)
}

func TestOKXConverts(t *testing.T) {
	info := toDomainAccountInfo(accountConfig{UID: "9", AcctLv: "2", Perm: "read_only,trade"})
	require.Equal(t, "9", *info.UID)
	require.True(t, *info.CanTrade)
	require.Len(t, info.Permissions, 2)

	bal := toDomainBalance([]balanceDetail{
		{Ccy: "BTC", AvailBal: "1", FrozenBal: "0"},
		{Ccy: "USDT", AvailBal: "10", FrozenBal: "2"},
	}, "usdt")
	require.Equal(t, "10", *bal.Free)
	require.Nil(t, toDomainBalance(nil, "ETH"))

	tk := toDomainTicker(ticker{InstID: "BTC-USDT", Last: "1", BidPx: "0.9", AskPx: "1.1", Ts: "8"})
	require.Equal(t, int64(8), *tk.Timestamp)

	candles := toDomainCandles([][]string{{"1", "10", "12", "9", "11", "100"}, {"bad"}})
	require.Len(t, candles, 1)
	require.Equal(t, "11", *candles[0].Close)

	book := toDomainOrderBook("BTC-USDT", orderBook{Bids: [][]string{{"99", "1"}}, Asks: [][]string{{"101", "2"}, {"x"}}})
	require.Len(t, book.Asks, 1)

	ord := toDomainOrder(order{InstID: "BTC-USDT", OrdID: "88", State: "live", OrdType: "limit", Side: "buy", CTime: "3"})
	require.Equal(t, "88", *ord.OrderID)
	require.Equal(t, "BUY", *ord.Side)
}

func TestParseOKXStream(t *testing.T) {
	upd, ok := parseOKXStream([]byte(`{"arg":{"channel":"tickers","instId":"BTC-USDT"},"data":[{"instId":"BTC-USDT","last":"100","bidPx":"99","askPx":"101","ts":"9"}]}`))
	require.True(t, ok)
	require.Equal(t, "100", *upd.Ticker.LastPrice)

	upd, ok = parseOKXStream([]byte(`{"event":"subscribe"}`))
	require.False(t, ok)
}
