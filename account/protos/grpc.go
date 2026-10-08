package protos

import (
	"context"
	"errors"
	"math/big"

	"github.com/Crows-Storm/Axis/account/app"
	"github.com/Crows-Storm/Axis/account/app/query"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
	"github.com/Crows-Storm/Axis/common/genproto/accountpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	accountpb.UnimplementedAccountServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) GetAccountsByUserID(ctx context.Context, request *accountpb.GetAccountsByUserIDRequest) (*accountpb.GetAccountsByUserIDResponse, error) {
	accounts, err := s.app.Queries.ListAccounts.Handle(ctx, query.ListAccountsQuery{UserID: request.GetUserId()})
	if err != nil {
		return nil, mapQueryError(err)
	}
	out := &accountpb.GetAccountsByUserIDResponse{UserId: request.GetUserId()}
	for _, acc := range accounts {
		item := &accountpb.Account{Id: acc.ID, Type: string(acc.Type)}
		for _, wallet := range acc.Wallet {
			item.Wallets = append(item.Wallets, &accountpb.Balance{
				WalletId: wallet.ID,
				Currency: wallet.Currency.Symbol,
				Amount:   commwallet.FormatMinor(wallet.Balance, wallet.Currency.Decimals),
			})
		}
		out.Accounts = append(out.Accounts, item)
	}
	return out, nil
}

func (s *GRPCServer) GetBalance(ctx context.Context, request *accountpb.GetBalanceRequest) (*accountpb.GetBalanceResponse, error) {
	symbol, ok := commaccount.ParseSymbol(request.GetCurrency())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid currency")
	}
	accounts, err := s.app.Queries.ListAccounts.Handle(ctx, query.ListAccountsQuery{UserID: request.GetUserId()})
	if err != nil {
		return nil, mapQueryError(err)
	}
	var decimals uint8
	total := big.NewInt(0)
	for _, acc := range accounts {
		for _, wallet := range acc.Wallet {
			if wallet.Currency.Symbol != symbol {
				continue
			}
			decimals = wallet.Currency.Decimals
			if wallet.Balance != nil {
				total.Add(total, wallet.Balance)
			}
		}
	}
	return &accountpb.GetBalanceResponse{Amount: commwallet.FormatMinor(total, decimals)}, nil
}

func mapQueryError(err error) error {
	switch {
	case errors.Is(err, commaccount.ErrNotFound):
		return status.Error(codes.NotFound, "account not found")
	case errors.Is(err, commaccount.ErrInvalid):
		return status.Error(codes.InvalidArgument, "invalid account query")
	default:
		return status.Error(codes.Internal, "account query failed")
	}
}
