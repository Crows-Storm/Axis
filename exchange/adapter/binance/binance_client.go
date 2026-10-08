package binance

import (
	"context"
	"errors"
	"fmt"

	"github.com/Crows-Storm/Axis/common/contextx"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/domain"
	spotclient "github.com/binance/binance-connector-go/clients/spot"
	walletclient "github.com/binance/binance-connector-go/clients/wallet"
	commbinance "github.com/binance/binance-connector-go/common/v2/common"
)

type Client struct{}

func NewClient() *Client { return &Client{} }

func (c *Client) Platform() commexchange.Platform {
	return commexchange.Binance
}

func (c *Client) AccountInfo(ctx context.Context) (*domain.AccountInfo, error) {
	sdk, err := c.accountSDK(ctx)
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
	sdk, err := c.walletSDK(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := sdk.RestApi.AssetAPI.AssetDetail(ctx).Asset(asset).Execute()
	if err != nil {
		return nil, fmt.Errorf("binance asset detail: %w", err)
	}
	return toDomainBalance(resp.Data, asset), nil
}

func (c *Client) accountSDK(ctx context.Context) (*spotclient.BinanceSpotClient, error) {
	cfg, err := configuration(ctx)
	if err != nil {
		return nil, err
	}
	return spotclient.NewBinanceSpotClient(
		spotclient.WithRestAPI(cfg),
	), nil
}

func (c *Client) walletSDK(ctx context.Context) (*walletclient.BinanceWalletClient, error) {
	cfg, err := configuration(ctx)
	if err != nil {
		return nil, err
	}
	return walletclient.NewBinanceWalletClient(
		walletclient.WithRestAPI(cfg),
	), nil
}

func configuration(ctx context.Context) (*commbinance.ConfigurationRestAPI, error) {
	cred, ok := contextx.CredentialsFromContext(ctx)
	if !ok || cred.APIKey == "" || cred.SecretKey == "" {
		return nil, errors.New("missing api credentials in context")
	}
	return commbinance.NewConfigurationRestAPI(
		commbinance.WithBasePath(commbinance.WalletRestApiProdUrl),
		commbinance.WithApiKey(cred.APIKey),
		commbinance.WithApiSecret(cred.SecretKey),
	), nil
}

var (
	_ domain.ExchangeClient        = (*Client)(nil)
	_ domain.AccountExchangeClient = (*Client)(nil)
	_ domain.WalletExchangeClient  = (*Client)(nil)
)
