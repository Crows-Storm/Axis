package role

import "context"

type Repository interface {
	Get(ctx context.Context, id uint64) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*Role, error)
	ListAssignedByUserID(ctx context.Context, userID uint64) ([]*Role, error)
	Create(ctx context.Context, role *Role) error
	AssignUser(ctx context.Context, userID uint64, role *Role) error
	UnassignUser(ctx context.Context, userID uint64, role *Role) error
	ReplaceUserRoles(ctx context.Context, userID uint64, roles []*Role) error
	Update(
		ctx context.Context,
		role *Role,
		updateFun func(context.Context, *Role) (*Role, error),
	) error
}
