package principal

import (
	"context"
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
