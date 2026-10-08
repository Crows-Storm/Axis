package app

import (
	"github.com/Crows-Storm/Axis/account/app/command"
	"github.com/Crows-Storm/Axis/account/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	OpenAccount command.OpenAccountCommandHandler
	OpenWallet  command.OpenWalletCommandHandler
}

type Queries struct {
	GetAccount     query.GetAccountQueryHandler
	ListAccounts   query.ListAccountsQueryHandler
	GetWallet      query.GetWalletQueryHandler
	ListCurrencies query.ListCurrenciesQueryHandler
}
