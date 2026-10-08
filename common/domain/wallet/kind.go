package wallet

import "strings"

// Type is the product bucket a balance lives in.
type Type string

const (
	Spot    Type = "spot"
	Futures Type = "futures"
	Savings Type = "savings"
	Card    Type = "card"
	Escrow  Type = "escrow"
)

func ParseType(raw string) (Type, bool) {
	kind := Type(strings.TrimSpace(raw))
	switch kind {
	case Spot, Futures, Savings, Card, Escrow:
		return kind, true
	default:
		return "", false
	}
}
