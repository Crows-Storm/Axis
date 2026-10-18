package binance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Crows-Storm/Axis/common/contextx"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
	spotclient "github.com/binance/binance-connector-go/clients/spot"
	spotmodels "github.com/binance/binance-connector-go/clients/spot/src/restapi/models"
	walletclient "github.com/binance/binance-connector-go/clients/wallet"
	commbinance "github.com/binance/binance-connector-go/common/v2/common"
	"github.com/gorilla/websocket"
)

type Client struct {
	wsURL  string
	mu     sync.Mutex
	conn   *websocket.Conn
	cancel context.CancelFunc
}

func NewClient() *Client {
	return &Client{wsURL: "wss://stream.binance.com:9443"}
}

func (c *Client) Platform() commexchange.Platform {
	return commexchange.Binance
}

func (c *Client) AccountInfo(ctx context.Context) (*domain.AccountInfo, error) {
	sdk, err := c.spotSDK(ctx, true)
	if err != nil {
		return nil, err
	}
	resp, err := sdk.RestApi.AccountAPI.GetAccount(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("binance account info: %w", err)
	}
	return toDomainAccountInfo(resp.Data), nil
}

func (c *Client) GetBalance(ctx context.Context, asset string) (*domain.Balance, error) {
	asset = strings.TrimSpace(asset)
	if asset == "" {
		return nil, errors.New("asset is required")
	}
	sdk, err := c.walletSDK(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := sdk.RestApi.AssetAPI.UserAsset(ctx).Asset(asset).Execute()
	if err != nil {
		return nil, fmt.Errorf("binance user asset: %w", err)
	}
	bal := toDomainBalanceFromUserAsset(resp.Data, asset)
	if bal == nil {
		return nil, fmt.Errorf("binance user asset: %s not found", asset)
	}
	return bal, nil
}

func (c *Client) GetTicker(ctx context.Context, symbol string) (*domain.Ticker, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	sdk, err := c.spotSDK(ctx, false)
	if err != nil {
		return nil, err
	}
	resp, err := sdk.RestApi.MarketAPI.Ticker24hr(ctx).Symbol(symbol).Execute()
	if err != nil {
		return nil, fmt.Errorf("binance ticker: %w", err)
	}
	ticker := toDomainTicker(resp.Data)
	if ticker == nil {
		return nil, errors.New("binance ticker: empty 24hr response")
	}
	return ticker, nil
}

func (c *Client) GetCandles(ctx context.Context, symbol string, interval commexchange.Interval, limit int) ([]domain.Candle, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	mapped, err := toBinanceInterval(interval)
	if err != nil {
		return nil, fmt.Errorf("binance klines interval: %w", err)
	}
	iv, err := spotmodels.NewKlinesIntervalParameterFromValue(mapped)
	if err != nil {
		return nil, fmt.Errorf("binance klines interval: %w", err)
	}
	sdk, err := c.spotSDK(ctx, false)
	if err != nil {
		return nil, err
	}
	req := sdk.RestApi.MarketAPI.Klines(ctx).Symbol(symbol).Interval(*iv)
	if limit > 0 {
		req = req.Limit(int32(limit))
	}
	resp, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("binance klines: %w", err)
	}
	return toDomainCandles(resp.Data), nil
}

func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*domain.OrderBook, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	sdk, err := c.spotSDK(ctx, false)
	if err != nil {
		return nil, err
	}
	req := sdk.RestApi.MarketAPI.Depth(ctx).Symbol(symbol)
	if depth > 0 {
		req = req.Limit(int32(depth))
	}
	resp, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("binance depth: %w", err)
	}
	return toDomainOrderBook(symbol, resp.Data), nil
}

func (c *Client) PlaceOrder(ctx context.Context, req *adapter.PlaceOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, errors.New("place order request is required")
	}
	market := strings.ToLower(strings.TrimSpace(req.Market))
	if market != "" && market != adapter.MarketSpot {
		return nil, fmt.Errorf("binance spot adapter does not support market %q", req.Market)
	}
	if strings.TrimSpace(req.Symbol) == "" {
		return nil, errors.New("symbol is required")
	}
	side, err := spotmodels.NewNewOrderSideParameterFromValue(strings.ToUpper(strings.TrimSpace(req.Side)))
	if err != nil {
		return nil, fmt.Errorf("order side: %w", err)
	}
	typ, err := spotmodels.NewNewOrderTypeParameterFromValue(strings.ToUpper(strings.TrimSpace(req.Type)))
	if err != nil {
		return nil, fmt.Errorf("order type: %w", err)
	}
	sdk, err := c.spotSDK(ctx, true)
	if err != nil {
		return nil, err
	}
	order := sdk.RestApi.TradeAPI.NewOrder(ctx).
		Symbol(req.Symbol).
		Side(*side).
		Type(*typ).
		NewOrderRespType(spotmodels.NewOrderNewOrderRespTypeParameterResult)
	if tif := strings.TrimSpace(req.TimeInForce); tif != "" {
		v, err := spotmodels.NewNewOrderTimeInForceParameterFromValue(strings.ToUpper(tif))
		if err != nil {
			return nil, fmt.Errorf("time in force: %w", err)
		}
		order = order.TimeInForce(*v)
	} else if *typ == spotmodels.NewOrderTypeParameterLimit {
		order = order.TimeInForce(spotmodels.NewOrderTimeInForceParameterGtc)
	}
	if req.Quantity != "" {
		qty, err := parseDecimal(req.Quantity, "quantity")
		if err != nil {
			return nil, err
		}
		order = order.Quantity(qty)
	}
	if req.QuoteQuantity != "" {
		qty, err := parseDecimal(req.QuoteQuantity, "quoteQuantity")
		if err != nil {
			return nil, err
		}
		order = order.QuoteOrderQty(qty)
	}
	if req.Price != "" {
		price, err := parseDecimal(req.Price, "price")
		if err != nil {
			return nil, err
		}
		order = order.Price(price)
	}
	if req.StopPrice != "" {
		stop, err := parseDecimal(req.StopPrice, "stopPrice")
		if err != nil {
			return nil, err
		}
		order = order.StopPrice(stop)
	}
	if req.ClientOrderID != "" {
		order = order.NewClientOrderId(req.ClientOrderID)
	}
	resp, err := order.Execute()
	if err != nil {
		return nil, fmt.Errorf("binance new order: %w", err)
	}
	return toDomainOrderFromNew(resp.Data), nil
}

func (c *Client) CancelOrder(ctx context.Context, req *adapter.CancelOrderRequest) error {
	if req == nil {
		return errors.New("cancel order request is required")
	}
	if strings.TrimSpace(req.Symbol) == "" {
		return errors.New("symbol is required")
	}
	if req.OrderID == "" && req.ClientOrderID == "" {
		return errors.New("orderID or clientOrderID is required")
	}
	sdk, err := c.spotSDK(ctx, true)
	if err != nil {
		return err
	}
	cancel := sdk.RestApi.TradeAPI.DeleteOrder(ctx).Symbol(req.Symbol)
	if req.OrderID != "" {
		id, err := strconv.ParseInt(req.OrderID, 10, 64)
		if err != nil {
			return fmt.Errorf("orderID: %w", err)
		}
		cancel = cancel.OrderId(id)
	}
	if req.ClientOrderID != "" {
		cancel = cancel.OrigClientOrderId(req.ClientOrderID)
	}
	if _, err := cancel.Execute(); err != nil {
		return fmt.Errorf("binance cancel order: %w", err)
	}
	return nil
}

func (c *Client) GetOrder(ctx context.Context, symbol, orderID string) (*domain.Order, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	if strings.TrimSpace(orderID) == "" {
		return nil, errors.New("orderID is required")
	}
	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("orderID: %w", err)
	}
	sdk, err := c.spotSDK(ctx, true)
	if err != nil {
		return nil, err
	}
	resp, err := sdk.RestApi.AccountAPI.GetOrder(ctx).Symbol(symbol).OrderId(id).Execute()
	if err != nil {
		return nil, fmt.Errorf("binance get order: %w", err)
	}
	return toDomainOrderFromGet(resp.Data), nil
}

func (c *Client) CloseOutPosition(_ context.Context, _ *adapter.CloseOutPositionRequest) error {
	return errors.New("binance spot adapter does not support close-out position")
}

func parseDecimal(raw, field string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", field, err)
	}
	return v, nil
}

func (c *Client) spotSDK(ctx context.Context, signed bool) (*spotclient.BinanceSpotClient, error) {
	cfg, err := configuration(ctx, commbinance.SpotRestApiProdUrl, signed)
	if err != nil {
		return nil, err
	}
	return spotclient.NewBinanceSpotClient(
		spotclient.WithRestAPI(cfg),
	), nil
}

func (c *Client) walletSDK(ctx context.Context) (*walletclient.BinanceWalletClient, error) {
	cfg, err := configuration(ctx, commbinance.WalletRestApiProdUrl, true)
	if err != nil {
		return nil, err
	}
	return walletclient.NewBinanceWalletClient(
		walletclient.WithRestAPI(cfg),
	), nil
}

func configuration(ctx context.Context, basePath string, signed bool) (*commbinance.ConfigurationRestAPI, error) {
	cred, ok := contextx.CredentialsFromContext(ctx)
	if signed && (!ok || cred.APIKey == "" || cred.SecretKey == "") {
		return nil, errors.New("missing api credentials in context")
	}
	opts := []commbinance.ConfigurationRestAPIOption{
		commbinance.WithBasePath(basePath),
	}
	if ok && cred.APIKey != "" {
		opts = append(opts,
			commbinance.WithApiKey(cred.APIKey),
			commbinance.WithApiSecret(cred.SecretKey),
		)
	}
	return commbinance.NewConfigurationRestAPI(opts...), nil
}

var (
	_ adapter.ExchangeClient            = (*Client)(nil)
	_ adapter.AccountExchangeClient     = (*Client)(nil)
	_ adapter.WalletExchangeClient      = (*Client)(nil)
	_ adapter.TickerExchangeClient      = (*Client)(nil)
	_ adapter.TransactionExchangeClient = (*Client)(nil)
	_ adapter.SyncMarketData            = (*Client)(nil)
)
