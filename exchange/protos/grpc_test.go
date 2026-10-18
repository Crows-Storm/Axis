package protos

import (
	"context"
	"testing"

	"github.com/Crows-Storm/Axis/common/contextx"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/common/genproto/exchangepb"
	"github.com/Crows-Storm/Axis/common/server/metrics"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/app"
	"github.com/Crows-Storm/Axis/exchange/app/command"
	"github.com/Crows-Storm/Axis/exchange/app/query"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

type stubClient struct {
	platform commexchange.Platform
	ticker   *domain.Ticker
}

func (s stubClient) Platform() commexchange.Platform { return s.platform }

func (s stubClient) AccountInfo(context.Context) (*domain.AccountInfo, error) {
	uid := "1"
	return &domain.AccountInfo{UID: &uid}, nil
}

func (s stubClient) GetBalance(context.Context, string) (*domain.Balance, error) {
	a, f := "USDT", "10"
	return &domain.Balance{Asset: &a, Free: &f}, nil
}

func (s stubClient) GetTicker(context.Context, string) (*domain.Ticker, error) {
	return s.ticker, nil
}

func (s stubClient) GetCandles(context.Context, string, commexchange.Interval, int) ([]domain.Candle, error) {
	return nil, nil
}

func (s stubClient) GetOrderBook(context.Context, string, int) (*domain.OrderBook, error) {
	return &domain.OrderBook{}, nil
}

func (s stubClient) PlaceOrder(context.Context, *adapter.PlaceOrderRequest) (*domain.Order, error) {
	id := "88"
	return &domain.Order{OrderID: &id}, nil
}

func (s stubClient) CancelOrder(context.Context, *adapter.CancelOrderRequest) error { return nil }

func (s stubClient) GetOrder(context.Context, string, string) (*domain.Order, error) {
	id := "88"
	return &domain.Order{OrderID: &id}, nil
}

func testApp() app.Application {
	last := "100"
	venues := adapter.NewRegistry(stubClient{
		platform: commexchange.Binance,
		ticker:   &domain.Ticker{LastPrice: &last},
	})
	m := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			PlaceOrder:  command.NewPlaceOrderHandler(venues, m),
			CancelOrder: command.NewCancelOrderHandler(venues, m),
		},
		Queries: app.Queries{
			GetAccountInfo: query.NewGetAccountInfoHandler(venues, m),
			GetBalance:     query.NewGetBalanceHandler(venues, m),
			GetTicker:      query.NewGetTickerHandler(venues, m),
			GetCandles:     query.NewGetCandlesHandler(venues, m),
			GetOrderBook:   query.NewGetOrderBookHandler(venues, m),
			GetOrder:       query.NewGetOrderHandler(venues, m),
		},
	}
}

func TestGRPCPingAndTicker(t *testing.T) {
	s := NewGRPCServer(testApp())
	st, err := s.Ping(context.Background(), &exchangepb.PingRequest{})
	require.NoError(t, err)
	require.Equal(t, "exchange", st.GetService())

	tk, err := s.GetTicker(context.Background(), &exchangepb.GetTickerRequest{
		Platform: "binance",
		Symbol:   "BTCUSDT",
	})
	require.NoError(t, err)
	require.Equal(t, "100", tk.GetLastPrice())

	_, err = s.GetTicker(context.Background(), &exchangepb.GetTickerRequest{Platform: "unknown"})
	require.Error(t, err)
}

func TestAttachCredentialsFromMetadata(t *testing.T) {
	md := metadata.Pairs(
		contextx.MetadataExchangeAPIKey, "k",
		contextx.MetadataExchangeAPISecret, "s",
		contextx.MetadataExchangePassphrase, "p",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = attachCredentials(ctx)
	cred, ok := contextx.CredentialsFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "k", cred.APIKey)
	require.Equal(t, "s", cred.SecretKey)
}

func TestPlaceAndCancelOrder(t *testing.T) {
	s := NewGRPCServer(testApp())
	ord, err := s.PlaceOrder(context.Background(), &exchangepb.PlaceOrderRequest{
		Platform: "Binance",
		Symbol:   "BTCUSDT",
		Side:     "BUY",
		Type:     "LIMIT",
	})
	require.NoError(t, err)
	require.Equal(t, "88", ord.GetOrderId())

	_, err = s.CancelOrder(context.Background(), &exchangepb.CancelOrderRequest{
		Platform: "binance",
		Symbol:   "BTCUSDT",
		OrderId:  "88",
	})
	require.NoError(t, err)
}
