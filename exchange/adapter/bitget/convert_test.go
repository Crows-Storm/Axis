package bitget

import (
	"testing"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/stretchr/testify/require"
)

func TestSymbolAndGranularity(t *testing.T) {
	require.Equal(t, "BTCUSDT", toBitgetSymbol("btc-usdt"))
	g, err := toBitgetGranularity(commexchange.OneHour)
	require.NoError(t, err)
	require.Equal(t, "1h", g)
	_, err = toBitgetGranularity(commexchange.OneSecond)
	require.Error(t, err)
	ch, err := toBitgetCandleChannel(commexchange.OneHour)
	require.NoError(t, err)
	require.Equal(t, "candle1H", ch)
}

func TestBitgetConverts(t *testing.T) {
	info := toDomainAccountInfo(accountInfo{UserID: "7", Authorities: []string{"spot", "p2p"}})
	require.Equal(t, "7", *info.UID)
	require.True(t, *info.CanTrade)

	bal := toDomainBalance([]spotAsset{
		{Coin: "USDT", Available: "10", Frozen: "2"},
	}, "usdt")
	require.Equal(t, "10", *bal.Free)

	tk := toDomainTicker(ticker{Symbol: "BTCUSDT", LastPr: "1", BidPr: "0.9", AskPr: "1.1", Ts: "8"})
	require.Equal(t, int64(8), *tk.Timestamp)

	candles := toDomainCandles([][]string{{"1", "10", "12", "9", "11", "100"}})
	require.Equal(t, "11", *candles[0].Close)

	ord := toDomainOrder(order{Symbol: "BTCUSDT", OrderID: "88", Status: "live", OrderType: "limit", Side: "buy", Force: "gtc", CTime: "3"})
	require.Equal(t, "88", *ord.OrderID)
	require.Equal(t, "GTC", *ord.TimeInForce)
}

func TestParseBitgetStream(t *testing.T) {
	upd, ok := parseBitgetStream([]byte(`{"arg":{"channel":"ticker","instId":"BTCUSDT"},"data":[{"instId":"BTCUSDT","lastPr":"100","bidPr":"99","askPr":"101","ts":"9"}]}`))
	require.True(t, ok)
	require.Equal(t, "100", *upd.Ticker.LastPrice)
}
