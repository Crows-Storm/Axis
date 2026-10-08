package main

import (
	"errors"

	"github.com/Crows-Storm/Axis/account/app"
	"github.com/Crows-Storm/Axis/account/app/command"
	"github.com/Crows-Storm/Axis/account/app/query"
	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/account/protos"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (h HTTPServer) GetMyAccounts(c *gin.Context, params protos.GetMyAccountsParams) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	converted := protos.GetAccountsByUserIdParams{Type: params.Type}
	h.GetAccountsByUserId(c, caller.UserID, converted)
}

func (h HTTPServer) GetAccountsByUserId(c *gin.Context, userID uint64, params protos.GetAccountsByUserIdParams) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	if !canRead(caller, userID) {
		server.ErrorWithCode(c, server.CodeForbidden)
		return
	}
	q := query.ListAccountsQuery{UserID: userID}
	if params.Type != nil {
		q.Type = string(*params.Type)
	}
	accounts, err := h.app.Queries.ListAccounts.Handle(c.Request.Context(), q)
	if err != nil {
		writeQueryError(c, err)
		return
	}
	out := make([]protos.AccountInfo, 0, len(accounts))
	for _, acc := range accounts {
		out = append(out, toAccountInfo(acc))
	}
	server.Success(c, out)
}

func (h HTTPServer) GetAccountById(c *gin.Context, id uint64) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	acc, err := h.app.Queries.GetAccount.Handle(c.Request.Context(), query.GetAccountQuery{ID: id})
	if err != nil {
		writeQueryError(c, err)
		return
	}
	if !canRead(caller, acc.UserID) {
		server.ErrorWithCode(c, server.CodeForbidden)
		return
	}
	server.Success(c, toAccountInfo(acc))
}

func (h HTTPServer) GetWalletById(c *gin.Context, id uint64) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	view, err := h.app.Queries.GetWallet.Handle(c.Request.Context(), query.GetWalletQuery{ID: id})
	if err != nil {
		writeQueryError(c, err)
		return
	}
	if !canRead(caller, view.Account.UserID) {
		server.ErrorWithCode(c, server.CodeForbidden)
		return
	}
	server.Success(c, toWalletInfo(view.Account.ID, view.Wallet))
}

func (h HTTPServer) ListCurrencies(c *gin.Context) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	rows, err := h.app.Queries.ListCurrencies.Handle(c.Request.Context(), query.ListCurrenciesQuery{})
	if err != nil {
		writeQueryError(c, err)
		return
	}
	out := make([]protos.Currency, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCurrency(row))
	}
	server.Success(c, out)
}

func (h HTTPServer) OpenAccount(c *gin.Context) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	var body protos.OpenAccountRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	userID := caller.UserID
	if body.UserId != nil && *body.UserId != 0 {
		if rbac.NormalizeAudience(caller.Audience()) != rbac.AudienceConsole {
			server.ErrorWithCode(c, server.CodeForbidden)
			return
		}
		userID = *body.UserId
	}
	acc, err := h.app.Commands.OpenAccount.Handle(c.Request.Context(), command.OpenAccountCommand{
		UserID: userID,
		Type:   string(body.Type),
	})
	if err != nil {
		writeQueryError(c, err)
		return
	}
	server.Success(c, toAccountInfo(acc))
}

func (h HTTPServer) OpenWallet(c *gin.Context, id uint64) {
	caller := principal.FromContext(c.Request.Context())
	if caller == nil || caller.UserID == 0 {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	acc, err := h.app.Queries.GetAccount.Handle(c.Request.Context(), query.GetAccountQuery{ID: id})
	if err != nil {
		writeQueryError(c, err)
		return
	}
	if !canWrite(caller, acc.UserID) {
		server.ErrorWithCode(c, server.CodeForbidden)
		return
	}
	var body protos.OpenWalletRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	cmd := command.OpenWalletCommand{AccountID: id}
	if body.CurrencyId != nil {
		cmd.CurrencyID = *body.CurrencyId
	}
	if body.Symbol != nil {
		cmd.Symbol = *body.Symbol
	}
	wallet, err := h.app.Commands.OpenWallet.Handle(c.Request.Context(), cmd)
	if err != nil {
		writeQueryError(c, err)
		return
	}
	server.Success(c, toWalletInfo(id, wallet))
}

func canRead(caller *principal.Principal, owner uint64) bool {
	if rbac.NormalizeAudience(caller.Audience()) == rbac.AudienceConsole {
		return true
	}
	return caller.UserID == owner
}

func canWrite(caller *principal.Principal, owner uint64) bool {
	if rbac.NormalizeAudience(caller.Audience()) == rbac.AudienceConsole {
		return true
	}
	return caller.UserID == owner
}

func writeQueryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, commaccount.ErrNotFound):
		server.ErrorWithCode(c, server.CodeNotFound)
	case errors.Is(err, commaccount.ErrInvalid):
		server.ErrorWithCode(c, server.CodeBadRequest)
	default:
		server.ErrorWithCode(c, server.CodeInternalServerError)
	}
}

func toAccountInfo(acc *domain.Account) protos.AccountInfo {
	wallets := make([]protos.WalletInfo, 0, len(acc.Wallet))
	for _, wallet := range acc.Wallet {
		wallets = append(wallets, toWalletInfo(acc.ID, wallet))
	}
	return protos.AccountInfo{
		Id:      acc.ID,
		UserId:  acc.UserID,
		Type:    protos.AccountType(acc.Type),
		Wallets: wallets,
	}
}

func toWalletInfo(accountID uint64, wallet domain.Wallet) protos.WalletInfo {
	balance := "0"
	if wallet.Balance != nil {
		balance = wallet.Balance.String()
	}
	info := protos.WalletInfo{
		Id:        wallet.ID,
		AccountId: accountID,
		Currency:  toCurrency(wallet.Currency),
		Balance:   balance,
	}
	if wallet.Currency.Address != "" {
		addr := wallet.Currency.Address
		info.Address = &addr
	}
	return info
}

func toCurrency(c commaccount.Currency) protos.Currency {
	out := protos.Currency{
		Id:           c.ID,
		Class:        protos.AssetClass(c.Class),
		Symbol:       c.Symbol,
		Name:         c.Name,
		Decimals:     int(c.Decimals),
		SmallestUnit: c.SmallestUnit,
	}
	if c.Address != "" {
		addr := c.Address
		out.Address = &addr
	}
	return out
}
