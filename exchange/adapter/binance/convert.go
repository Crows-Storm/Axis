package binance

import (
	"github.com/Crows-Storm/Axis/exchange/domain"
	sportmodels "github.com/binance/binance-connector-go/clients/spot/src/restapi/models"
	walletmodels "github.com/binance/binance-connector-go/clients/wallet/src/restapi/models"
)

// TODO: be Convert
func toDomainAccountInfo(repo sportmodels.GetAccountResponse) *domain.AccountInfo {
	return &domain.AccountInfo{
		CanTrade: repo.CanTrade,
	}
}

func toDomainBalance(repo walletmodels.AssetDetailResponse, asset string) *domain.Balance {
	return &domain.Balance{}
}
