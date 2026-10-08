package wallet

import "strings"

// Currency is a balance unit shared by wallet and ledger.
type Currency string

const (
	USD  Currency = "USD"
	EUR  Currency = "EUR"
	CNY  Currency = "CNY"
	JPY  Currency = "JPY"
	GBP  Currency = "GBP"
	BTC  Currency = "BTC"
	ETH  Currency = "ETH"
	USDT Currency = "USDT"
)

func ParseCurrency(raw string) (Currency, bool) {
	currency := Currency(strings.ToUpper(strings.TrimSpace(raw)))
	switch currency {
	case USD, EUR, CNY, JPY, GBP, BTC, ETH, USDT:
		return currency, true
	default:
		return "", false
	}
}
