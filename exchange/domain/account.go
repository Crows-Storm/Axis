package domain

type AccountInfo struct {
	AccountType *string
	UID         *string
	Permissions []*string
	CanTrade    *bool
}

type Balance struct {
	Asset  *string
	Free   *string
	Locked *string
}

type Ticker struct {
	Symbol    *string
	LastPrice *string
	BidPrice  *string
	AskPrice  *string
	Timestamp *int64
}
