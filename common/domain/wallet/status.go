package wallet

import "strings"

// Status is the lifecycle of a wallet. Values match the wallet HTTP contract.
type Status string

const (
	Active  Status = "active"
	Frozen  Status = "frozen"
	Closed  Status = "closed"
	Pending Status = "pending"
)

func ParseStatus(raw string) (Status, bool) {
	status := Status(strings.TrimSpace(raw))
	switch status {
	case Active, Frozen, Closed, Pending:
		return status, true
	default:
		return "", false
	}
}
