package okx

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
	return commexchange.Okx
}

func (c *Client) AccountInfo(ctx context.Context) (*domain.AccountInfo, error) {
	var data []accountConfig
	if err := c.get(ctx, "/api/v5/account/config", nil, true, &data); err != nil {
		return nil, fmt.Errorf("okx account info: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("okx account info: empty response")
	}
	return toDomainAccountInfo(data[0]), nil
}

func (c *Client) GetBalance(ctx context.Context, asset string) (*domain.Balance, error) {
	asset = strings.TrimSpace(asset)
	if asset == "" {
		return nil, errors.New("asset is required")
	}
	q := url.Values{"ccy": {strings.ToUpper(asset)}}
	var data []balanceBook
	if err := c.get(ctx, "/api/v5/account/balance", q, true, &data); err != nil {
		return nil, fmt.Errorf("okx balance: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("okx balance: %s not found", asset)
	}
	bal := toDomainBalance(data[0].Details, asset)
	if bal == nil {
		return nil, fmt.Errorf("okx balance: %s not found", asset)
	}
	return bal, nil
}

func (c *Client) GetTicker(ctx context.Context, symbol string) (*domain.Ticker, error) {
	instID, err := requireInstID(symbol)
	if err != nil {
		return nil, err
	}
	var data []ticker
	if err := c.get(ctx, "/api/v5/market/ticker", url.Values{"instId": {instID}}, false, &data); err != nil {
		return nil, fmt.Errorf("okx ticker: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("okx ticker: empty response")
	}
	return toDomainTicker(data[0]), nil
}

func (c *Client) GetCandles(ctx context.Context, symbol string, interval commexchange.Interval, limit int) ([]domain.Candle, error) {
	instID, err := requireInstID(symbol)
	if err != nil {
		return nil, err
	}
	bar, err := toOKXBar(interval)
	if err != nil {
		return nil, fmt.Errorf("okx klines interval: %w", err)
	}
	q := url.Values{"instId": {instID}, "bar": {bar}}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var data [][]string
	if err := c.get(ctx, "/api/v5/market/candles", q, false, &data); err != nil {
		return nil, fmt.Errorf("okx klines: %w", err)
	}
	return toDomainCandles(data), nil
}

func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*domain.OrderBook, error) {
	instID, err := requireInstID(symbol)
	if err != nil {
		return nil, err
	}
	q := url.Values{"instId": {instID}}
	if depth > 0 {
		q.Set("sz", strconv.Itoa(depth))
	}
	var data []orderBook
	if err := c.get(ctx, "/api/v5/market/books", q, false, &data); err != nil {
		return nil, fmt.Errorf("okx depth: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("okx depth: empty response")
	}
	return toDomainOrderBook(instID, data[0]), nil
}

func (c *Client) PlaceOrder(ctx context.Context, req *adapter.PlaceOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, errors.New("place order request is required")
	}
	if err := rejectNonSpot(req.Market); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.StopPrice) != "" {
		return nil, errors.New("okx spot adapter does not support stop orders")
	}
	instID, err := requireInstID(req.Symbol)
	if err != nil {
		return nil, err
	}
	side, err := toOKXSide(req.Side)
	if err != nil {
		return nil, err
	}
	ordType, err := toOKXOrdType(req.Type, req.TimeInForce)
	if err != nil {
		return nil, err
	}
	body := map[string]string{
		"instId":  instID,
		"tdMode":  "cash",
		"side":    side,
		"ordType": ordType,
	}
	switch {
	case req.QuoteQuantity != "":
		body["sz"] = req.QuoteQuantity
		body["tgtCcy"] = "quote_ccy"
	case req.Quantity != "":
		body["sz"] = req.Quantity
	default:
		return nil, errors.New("quantity or quoteQuantity is required")
	}
	if req.Price != "" {
		body["px"] = req.Price
	}
	if req.ClientOrderID != "" {
		body["clOrdId"] = req.ClientOrderID
	}
	var data []order
	if err := c.post(ctx, "/api/v5/trade/order", body, &data); err != nil {
		return nil, fmt.Errorf("okx new order: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("okx new order: empty response")
	}
	return toDomainOrder(data[0]), nil
}

func (c *Client) CancelOrder(ctx context.Context, req *adapter.CancelOrderRequest) error {
	if req == nil {
		return errors.New("cancel order request is required")
	}
	instID, err := requireInstID(req.Symbol)
	if err != nil {
		return err
	}
	if req.OrderID == "" && req.ClientOrderID == "" {
		return errors.New("orderID or clientOrderID is required")
	}
	body := map[string]string{"instId": instID}
	if req.OrderID != "" {
		body["ordId"] = req.OrderID
	}
	if req.ClientOrderID != "" {
		body["clOrdId"] = req.ClientOrderID
	}
	if err := c.post(ctx, "/api/v5/trade/cancel-order", body, nil); err != nil {
		return fmt.Errorf("okx cancel order: %w", err)
	}
	return nil
}

func (c *Client) GetOrder(ctx context.Context, symbol, orderID string) (*domain.Order, error) {
	instID, err := requireInstID(symbol)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(orderID) == "" {
		return nil, errors.New("orderID is required")
	}
	q := url.Values{"instId": {instID}, "ordId": {orderID}}
	var data []order
	if err := c.get(ctx, "/api/v5/trade/order", q, true, &data); err != nil {
		return nil, fmt.Errorf("okx get order: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("okx get order: empty response")
	}
	return toDomainOrder(data[0]), nil
}

func (c *Client) CloseOutPosition(_ context.Context, _ *adapter.CloseOutPositionRequest) error {
	return errors.New("okx spot adapter does not support close-out position")
}

func requireInstID(symbol string) (string, error) {
	id := toOKXInstID(symbol)
	if id == "" {
		return "", errors.New("symbol is required")
	}
	return id, nil
}

func rejectNonSpot(market string) error {
	m := strings.ToLower(strings.TrimSpace(market))
	if m != "" && m != adapter.MarketSpot {
		return fmt.Errorf("okx spot adapter does not support market %q", market)
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
