package domain

import "context"

type Repository interface {
	GetPrincipal(ctx context.Context, token string) (*Principal, error)
}
