package bitget

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/domain"
)

func (c *Client) Platform() commexchange.Platform {
	return commexchange.Bitget
}

func (c *Client) AccountInfo(ctx context.Context) (*domain.AccountInfo, error) {
	var data accountInfo
	if err := c.get(ctx, "/api/v2/spot/account/info", nil, true, &data); err != nil {
		return nil, fmt.Errorf("bitget account info: %w", err)
	}
	return toDomainAccountInfo(data), nil
}

func (c *Client) GetBalance(ctx context.Context, asset string) (*domain.Balance, error) {
	asset = strings.TrimSpace(asset)
	if asset == "" {
		return nil, errors.New("asset is required")
	}
	q := url.Values{"coin": {strings.ToUpper(asset)}}
	var data []spotAsset
	if err := c.get(ctx, "/api/v2/spot/account/assets", q, true, &data); err != nil {
		return nil, fmt.Errorf("bitget user asset: %w", err)
	}
	bal := toDomainBalance(data, asset)
	if bal == nil {
		return nil, fmt.Errorf("bitget user asset: %s not found", asset)
	}
	return bal, nil
}

func (c *Client) GetTicker(ctx context.Context, symbol string) (*domain.Ticker, error) {
	sym, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	var data []ticker
	if err := c.get(ctx, "/api/v2/spot/market/tickers", url.Values{"symbol": {sym}}, false, &data); err != nil {
		return nil, fmt.Errorf("bitget ticker: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("bitget ticker: empty response")
	}
	return toDomainTicker(data[0]), nil
}

func (c *Client) GetCandles(ctx context.Context, symbol string, interval commexchange.Interval, limit int) ([]domain.Candle, error) {
	sym, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	gran, err := toBitgetGranularity(interval)
	if err != nil {
		return nil, fmt.Errorf("bitget klines interval: %w", err)
	}
	q := url.Values{"symbol": {sym}, "granularity": {gran}}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var data [][]string
	if err := c.get(ctx, "/api/v2/spot/market/candles", q, false, &data); err != nil {
		return nil, fmt.Errorf("bitget klines: %w", err)
	}
	return toDomainCandles(data), nil
}

func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*domain.OrderBook, error) {
	sym, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	q := url.Values{"symbol": {sym}, "type": {"step0"}}
	if depth > 0 {
		q.Set("limit", strconv.Itoa(depth))
	}
	var data orderBook
	if err := c.get(ctx, "/api/v2/spot/market/orderbook", q, false, &data); err != nil {
		return nil, fmt.Errorf("bitget depth: %w", err)
	}
	return toDomainOrderBook(sym, data), nil
}

func (c *Client) PlaceOrder(ctx context.Context, req *adapter.PlaceOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, errors.New("place order request is required")
	}
	if err := rejectNonSpot(req.Market); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.StopPrice) != "" {
		return nil, errors.New("bitget spot adapter does not support stop orders")
	}
	sym, err := requireSymbol(req.Symbol)
	if err != nil {
		return nil, err
	}
	side, err := toBitgetSide(req.Side)
	if err != nil {
		return nil, err
	}
	ordType, err := toBitgetOrderType(req.Type)
	if err != nil {
		return nil, err
	}
	force, err := toBitgetForce(req.TimeInForce)
	if err != nil {
		return nil, err
	}
	size := req.Quantity
	if size == "" {
		size = req.QuoteQuantity
	}
	if size == "" {
		return nil, errors.New("quantity or quoteQuantity is required")
	}
	body := map[string]string{
		"symbol":    sym,
		"side":      side,
		"orderType": ordType,
		"force":     force,
		"size":      size,
	}
	if req.Price != "" {
		body["price"] = req.Price
	}
	if req.ClientOrderID != "" {
		body["clientOid"] = req.ClientOrderID
	}
	var data order
	if err := c.post(ctx, "/api/v2/spot/trade/place-order", body, &data); err != nil {
		return nil, fmt.Errorf("bitget new order: %w", err)
	}
	return toDomainOrder(data), nil
}

func (c *Client) CancelOrder(ctx context.Context, req *adapter.CancelOrderRequest) error {
	if req == nil {
		return errors.New("cancel order request is required")
	}
	sym, err := requireSymbol(req.Symbol)
	if err != nil {
		return err
	}
	if req.OrderID == "" && req.ClientOrderID == "" {
		return errors.New("orderID or clientOrderID is required")
	}
	body := map[string]string{"symbol": sym}
	if req.OrderID != "" {
		body["orderId"] = req.OrderID
	}
	if req.ClientOrderID != "" {
		body["clientOid"] = req.ClientOrderID
	}
	if err := c.post(ctx, "/api/v2/spot/trade/cancel-order", body, nil); err != nil {
		return fmt.Errorf("bitget cancel order: %w", err)
	}
	return nil
}

func (c *Client) GetOrder(ctx context.Context, symbol, orderID string) (*domain.Order, error) {
	sym, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(orderID) == "" {
		return nil, errors.New("orderID is required")
	}
	q := url.Values{"symbol": {sym}, "orderId": {orderID}}
	var data []order
	if err := c.get(ctx, "/api/v2/spot/trade/orderInfo", q, true, &data); err != nil {
		return nil, fmt.Errorf("bitget get order: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("bitget get order: empty response")
	}
	return toDomainOrder(data[0]), nil
}

func (c *Client) CloseOutPosition(_ context.Context, _ *adapter.CloseOutPositionRequest) error {
	return errors.New("bitget spot adapter does not support close-out position")
}

func requireSymbol(symbol string) (string, error) {
	s := toBitgetSymbol(symbol)
	if s == "" {
		return "", errors.New("symbol is required")
	}
	return s, nil
}

func rejectNonSpot(market string) error {
	m := strings.ToLower(strings.TrimSpace(market))
	if m != "" && m != adapter.MarketSpot {
		return fmt.Errorf("bitget spot adapter does not support market %q", market)
	}
	return nil
}

var (
	_ adapter.ExchangeClient            = (*Client)(nil)
	_ adapter.AccountExchangeClient     = (*Client)(nil)
	_ adapter.WalletExchangeClient      = (*Client)(nil)
	_ adapter.TickerExchangeClient      = (*Client)(nil)
	_ adapter.TransactionExchangeClient = (*Client)(nil)
	_ adapter.SyncMarketData            = (*Client)(nil)
)
