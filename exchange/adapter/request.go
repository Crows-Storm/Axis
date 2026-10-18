package adapter

type AccountDetailRequest struct {
	Asset string
}

// ExchangeInfoRequest pulls the venue trading-pair catalog.
type ExchangeInfoRequest struct {
	Symbol string
}

type CloseOutPositionRequest struct {
	Symbol string
}
