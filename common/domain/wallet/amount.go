package wallet

import (
	"fmt"
	"math/big"
	"strings"
)

// AmountScale is the number of minor units in one major unit.
// One unit is 1e8 minors so both fiat and crypto fit in int64.
const AmountScale int64 = 100_000_000

// FormatDecimal renders a minor-unit amount without binary floating point.
func FormatDecimal(minor int64) string {
	negative := minor < 0
	if negative {
		minor = -minor
	}
	whole := minor / AmountScale
	frac := minor % AmountScale
	text := fmt.Sprintf("%d.%08d", whole, frac)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" {
		text = "0"
	}
	if negative {
		return "-" + text
	}
	return text
}

// ToFloat converts minor units for the wallet HTTP contract, which uses JSON numbers.
func ToFloat(minor int64) float64 {
	return float64(minor) / float64(AmountScale)
}

// FormatMinor renders an integer amount in Currency.SmallestUnit using decimals.
func FormatMinor(amount *big.Int, decimals uint8) string {
	if amount == nil {
		return "0"
	}
	negative := amount.Sign() < 0
	abs := new(big.Int).Abs(amount)
	if decimals == 0 {
		if negative {
			return "-" + abs.String()
		}
		return abs.String()
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	frac := new(big.Int)
	whole, frac := new(big.Int).QuoRem(abs, scale, frac)
	fracText := frac.String()
	for len(fracText) < int(decimals) {
		fracText = "0" + fracText
	}
	fracText = strings.TrimRight(fracText, "0")
	text := whole.String()
	if fracText != "" {
		text += "." + fracText
	}
	if negative {
		return "-" + text
	}
	return text
}
