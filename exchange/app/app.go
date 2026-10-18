package app

import (
	"github.com/Crows-Storm/Axis/exchange/app/command"
	"github.com/Crows-Storm/Axis/exchange/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	PlaceOrder  command.PlaceOrderHandler
	CancelOrder command.CancelOrderHandler
}

type Queries struct {
	GetAccountInfo query.GetAccountInfoHandler
	GetBalance     query.GetBalanceHandler
	GetTicker      query.GetTickerHandler
	GetCandles     query.GetCandlesHandler
	GetOrderBook   query.GetOrderBookHandler
	GetOrder       query.GetOrderHandler
}
