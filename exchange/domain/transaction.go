package domain

type Order struct {
	Symbol        *string
	OrderID       *string
	ClientOrderID *string
	Status        *string
	Type          *string
	Side          *string
	Price         *string
	OrigQty       *string
	ExecutedQty   *string
	TimeInForce   *string
	TransactTime  *int64
}
