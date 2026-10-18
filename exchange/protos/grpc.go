package protos

import (
	"context"
	"errors"
	"strings"

	"github.com/Crows-Storm/Axis/common/contextx"
	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/common/genproto/exchangepb"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/app"
	"github.com/Crows-Storm/Axis/exchange/app/command"
	"github.com/Crows-Storm/Axis/exchange/app/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceRole = "exchange-integration"

type GRPCServer struct {
	exchangepb.UnimplementedExchangeServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) Ping(_ context.Context, _ *exchangepb.PingRequest) (*exchangepb.Status, error) {
	return &exchangepb.Status{Service: "exchange", Role: serviceRole}, nil
}

func (s *GRPCServer) GetAccountInfo(ctx context.Context, req *exchangepb.GetAccountInfoRequest) (*exchangepb.AccountInfo, error) {
	ctx = attachCredentials(ctx)
	info, err := s.app.Queries.GetAccountInfo.Handle(ctx, query.GetAccountInfoQuery{Platform: req.GetPlatform()})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBAccountInfo(req.GetPlatform(), info), nil
}

func (s *GRPCServer) GetBalance(ctx context.Context, req *exchangepb.GetBalanceRequest) (*exchangepb.Balance, error) {
	ctx = attachCredentials(ctx)
	bal, err := s.app.Queries.GetBalance.Handle(ctx, query.GetBalanceQuery{
		Platform: req.GetPlatform(),
		Asset:    req.GetAsset(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBBalance(bal), nil
}

func (s *GRPCServer) GetTicker(ctx context.Context, req *exchangepb.GetTickerRequest) (*exchangepb.Ticker, error) {
	ctx = attachCredentials(ctx)
	tk, err := s.app.Queries.GetTicker.Handle(ctx, query.GetTickerQuery{
		Platform: req.GetPlatform(),
		Symbol:   req.GetSymbol(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBTicker(tk), nil
}

func (s *GRPCServer) GetCandles(ctx context.Context, req *exchangepb.GetCandlesRequest) (*exchangepb.GetCandlesResponse, error) {
	ctx = attachCredentials(ctx)
	iv := commexchange.Interval(req.GetInterval())
	rows, err := s.app.Queries.GetCandles.Handle(ctx, query.GetCandlesQuery{
		Platform: req.GetPlatform(),
		Symbol:   req.GetSymbol(),
		Interval: iv,
		Limit:    int(req.GetLimit()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &exchangepb.GetCandlesResponse{Candles: toPBCandles(rows)}, nil
}

func (s *GRPCServer) GetOrderBook(ctx context.Context, req *exchangepb.GetOrderBookRequest) (*exchangepb.OrderBook, error) {
	ctx = attachCredentials(ctx)
	book, err := s.app.Queries.GetOrderBook.Handle(ctx, query.GetOrderBookQuery{
		Platform: req.GetPlatform(),
		Symbol:   req.GetSymbol(),
		Depth:    int(req.GetDepth()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBOrderBook(book), nil
}

func (s *GRPCServer) PlaceOrder(ctx context.Context, req *exchangepb.PlaceOrderRequest) (*exchangepb.Order, error) {
	ctx = attachCredentials(ctx)
	ord, err := s.app.Commands.PlaceOrder.Handle(ctx, command.PlaceOrder{
		Platform: req.GetPlatform(),
		Request: adapter.PlaceOrderRequest{
			Symbol:        req.GetSymbol(),
			Side:          req.GetSide(),
			Type:          req.GetType(),
			Market:        req.GetMarket(),
			Quantity:      req.GetQuantity(),
			QuoteQuantity: req.GetQuoteQuantity(),
			Price:         req.GetPrice(),
			TimeInForce:   req.GetTimeInForce(),
			ClientOrderID: req.GetClientOrderId(),
			StopPrice:     req.GetStopPrice(),
		},
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBOrder(ord), nil
}

func (s *GRPCServer) CancelOrder(ctx context.Context, req *exchangepb.CancelOrderRequest) (*exchangepb.CancelOrderResponse, error) {
	ctx = attachCredentials(ctx)
	_, err := s.app.Commands.CancelOrder.Handle(ctx, command.CancelOrder{
		Platform: req.GetPlatform(),
		Request: adapter.CancelOrderRequest{
			Symbol:        req.GetSymbol(),
			OrderID:       req.GetOrderId(),
			ClientOrderID: req.GetClientOrderId(),
		},
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &exchangepb.CancelOrderResponse{}, nil
}

func (s *GRPCServer) GetOrder(ctx context.Context, req *exchangepb.GetOrderRequest) (*exchangepb.Order, error) {
	ctx = attachCredentials(ctx)
	ord, err := s.app.Queries.GetOrder.Handle(ctx, query.GetOrderQuery{
		Platform: req.GetPlatform(),
		Symbol:   req.GetSymbol(),
		OrderID:  req.GetOrderId(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toPBOrder(ord), nil
}

func attachCredentials(ctx context.Context) context.Context {
	if _, ok := contextx.CredentialsFromContext(ctx); ok {
		return ctx
	}
	cred, ok := contextx.CredentialsFromIncomingMetadata(ctx)
	if !ok {
		return ctx
	}
	return contextx.WithCredentials(ctx, cred)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unknown platform"),
		strings.Contains(msg, "required"),
		strings.Contains(msg, "invalid"),
		strings.Contains(msg, "unsupported"):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}
