package menu

import "context"

type Repository interface {
	ListActive(ctx context.Context) ([]*Menu, error)
	ListAll(ctx context.Context) ([]*Menu, error)
	Get(ctx context.Context, id uint64) (*Menu, error)
	GetByCode(ctx context.Context, code string) (*Menu, error)
	Create(ctx context.Context, menu *Menu) error
	Update(ctx context.Context, menu *Menu) error
}
