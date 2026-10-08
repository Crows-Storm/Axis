package account

import "strings"

// Currency is a row from sys_currency.
type Currency struct {
	ID           uint64
	Class        AssetClass
	Symbol       string
	Name         string
	Address      string
	Decimals     uint8
	SmallestUnit string
}

func ParseSymbol(raw string) (string, bool) {
	symbol := strings.ToUpper(strings.TrimSpace(raw))
	if symbol == "" || len(symbol) > 32 {
		return "", false
	}
	return symbol, true
}
