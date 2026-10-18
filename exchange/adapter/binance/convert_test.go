package binance

import (
	"testing"

	spotmodels "github.com/binance/binance-connector-go/clients/spot/src/restapi/models"
	walletmodels "github.com/binance/binance-connector-go/clients/wallet/src/restapi/models"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestToDomainAccountInfo(t *testing.T) {
	got := toDomainAccountInfo(spotmodels.GetAccountResponse{
		CanTrade:    ptr(true),
		AccountType: ptr("SPOT"),
		Uid:         ptr(int64(42)),
		Permissions: []string{"SPOT"},
	})
	require.NotNil(t, got.CanTrade)
	require.True(t, *got.CanTrade)
	require.Equal(t, "SPOT", *got.AccountType)
	require.Equal(t, "42", *got.UID)
	require.Len(t, got.Permissions, 1)
	require.Equal(t, "SPOT", *got.Permissions[0])
}

func TestToDomainBalanceFromUserAsset(t *testing.T) {
	resp := walletmodels.UserAssetResponse{
		Items: []walletmodels.UserAssetResponseInner{
			{Asset: ptr("BTC"), Free: ptr("1"), Locked: ptr("0")},
			{Asset: ptr("USDT"), Free: ptr("10.5"), Locked: ptr("2")},
		},
	}
	got := toDomainBalanceFromUserAsset(resp, "usdt")
	require.NotNil(t, got)
	require.Equal(t, "USDT", *got.Asset)
	require.Equal(t, "10.5", *got.Free)
	require.Equal(t, "2", *got.Locked)
	require.Nil(t, toDomainBalanceFromUserAsset(resp, "ETH"))
}

func TestToDomainTicker(t *testing.T) {
	require.Nil(t, toDomainTicker(spotmodels.Ticker24hrResponse{}))
	got := toDomainTicker(spotmodels.Ticker24hrResponse{
		Ticker24hrResponse1: &spotmodels.Ticker24hrResponse1{
			Symbol:    ptr("BTCUSDT"),
			LastPrice: ptr("100"),
			BidPrice:  ptr("99"),
			AskPrice:  ptr("101"),
			CloseTime: ptr(int64(9)),
		},
	})
	require.Equal(t, "BTCUSDT", *got.Symbol)
	require.Equal(t, "100", *got.LastPrice)
	require.Equal(t, int64(9), *got.Timestamp)
}

func TestToDomainCandlesAndOrderBook(t *testing.T) {
	open := int64(1)
	closeT := int64(2)
	candles := toDomainCandles(spotmodels.KlinesResponse{
		Items: []spotmodels.KlinesItem{
			{Items: []spotmodels.KlinesItemInner{
				{Int64: &open},
				{String: ptr("10")},
				{String: ptr("12")},
				{String: ptr("9")},
				{String: ptr("11")},
				{String: ptr("100")},
				{Int64: &closeT},
			}},
			{Items: []spotmodels.KlinesItemInner{{Int64: &open}}},
		},
	})
	require.Len(t, candles, 1)
	require.Equal(t, int64(1), candles[0].OpenTime)
	require.Equal(t, "11", *candles[0].Close)

	book := toDomainOrderBook("BTCUSDT", spotmodels.DepthResponse{
		Bids: [][]string{{"99", "1"}},
		Asks: [][]string{{"101", "2"}, {"bad"}},
	})
	require.Equal(t, "BTCUSDT", *book.Symbol)
	require.Len(t, book.Bids, 1)
	require.Len(t, book.Asks, 1)
	require.Equal(t, "101", *book.Asks[0].Price)
}

func TestToDomainOrder(t *testing.T) {
	id := int64(88)
	newOrd := toDomainOrderFromNew(spotmodels.NewOrderResponse{
		Symbol:        ptr("BTCUSDT"),
		OrderId:       &id,
		ClientOrderId: ptr("c1"),
		Status:        ptr("NEW"),
		Type:          ptr("LIMIT"),
		Side:          ptr("BUY"),
		Price:         ptr("1"),
		OrigQty:       ptr("2"),
		ExecutedQty:   ptr("0"),
		TimeInForce:   ptr("GTC"),
		TransactTime:  ptr(int64(3)),
	})
	require.Equal(t, "88", *newOrd.OrderID)
	require.Equal(t, int64(3), *newOrd.TransactTime)

	got := toDomainOrderFromGet(spotmodels.GetOrderResponse{
		Symbol:  ptr("BTCUSDT"),
		OrderId: &id,
		Status:  ptr("FILLED"),
		Time:    ptr(int64(4)),
	})
	require.Equal(t, "FILLED", *got.Status)
	require.Equal(t, int64(4), *got.TransactTime)
}

func TestParseBinanceStream(t *testing.T) {
	upd, ok := parseBinanceStream([]byte(`{"stream":"btcusdt@ticker","data":{"e":"24hrTicker","s":"BTCUSDT","c":"100","b":"99","a":"101","E":9}}`))
	require.True(t, ok)
	require.Equal(t, "100", *upd.Ticker.LastPrice)

	upd, ok = parseBinanceStream([]byte(`{"e":"kline","k":{"t":1,"T":2,"o":"10","h":"12","l":"9","c":"11","v":"100"}}`))
	require.True(t, ok)
	require.Equal(t, "11", *upd.Candle.Close)
}
