package account

import "strings"

type AssetClass string

const (
	AssetClassCrypto    AssetClass = "crypto"
	AssetClassFiat      AssetClass = "fiat"
	AssetClassStock     AssetClass = "stock"
	AssetClassCommodity AssetClass = "commodity"
)

func ParseAssetClass(raw string) (AssetClass, bool) {
	class := AssetClass(strings.ToLower(strings.TrimSpace(raw)))
	switch class {
	case AssetClassCrypto, AssetClassFiat, AssetClassStock, AssetClassCommodity:
		return class, true
	default:
		return "", false
	}
}
