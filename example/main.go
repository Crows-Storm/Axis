package main

import (
	"context"
	"encoding/json"
	"log"

	client "github.com/binance/binance-connector-go/clients/wallet"
	commbinance "github.com/binance/binance-connector-go/common/v2/common"
)

func main() {
	AccountInfo()
}

func AccountInfo() {
	configuration := commbinance.NewConfigurationRestAPI(
		commbinance.WithBasePath(commbinance.WalletRestApiProdUrl),
		//commbinance.WithApiKey("FTT2G2foziZqdFI94JTmKd2zsE0iFlVURLmHl9iTqCWrE"),
		//commbinance.WithApiSecret("uEGaB0rRMJUoSGc8zDsoRqHbCBAWY"),
	)
	apiClient := client.NewBinanceWalletClient(
		client.WithRestAPI(configuration),
	)
	resp, err := apiClient.RestApi.AccountAPI.AccountInfo(context.Background()).Execute()
	if err != nil {
		log.Println(err)
		return
	}

	accountInfo, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("Account Info: %s\n", string(accountInfo))

	rateLimitsValue, _ := json.MarshalIndent(resp.RateLimits, "", "  ")
	log.Printf("Rate limits: %s\n", string(rateLimitsValue))

	dataValue, _ := json.MarshalIndent(resp.Data, "", "  ")
	log.Printf("Response: %s\n", string(dataValue))

	asset, err := apiClient.RestApi.AssetAPI.AssetDetail(context.Background()).Asset("USDT").Execute()

	assetInfo, _ := json.MarshalIndent(asset, "", "  ")
	log.Printf("Asset Info: %s\n", assetInfo)
}
