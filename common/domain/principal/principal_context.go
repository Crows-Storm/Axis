package principal

import (
	"context"
	"fmt"
)

type ContextKey struct{}

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, ContextKey{}, p)
}

func FromContext(ctx context.Context) *Principal {
	if p, ok := ctx.Value(ContextKey{}).(*Principal); ok && p != nil {
		return p
	}
	return nil
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	p := FromContext(ctx)
	if p == nil {
		return 0, fmt.Errorf("principal not found in context")
	}
	return p.UserId, nil
}
