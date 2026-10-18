package exchange

import "strings"

type Platform string

const (
	Binance Platform = "Binance"
	Okx     Platform = "OKX"
	Bitget  Platform = "Bitget"
)

func ParsePlatform(raw string) (Platform, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "binance":
		return Binance, true
	case "okx":
		return Okx, true
	case "bitget":
		return Bitget, true
	default:
		return "", false
	}
}
