package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/rbac"
)

func afterOK[C any, R any](inner decorator.CommandHandler[C, R], fn func(context.Context, C)) decorator.CommandHandler[C, R] {
	return afterHandler[C, R]{inner: inner, after: fn}
}

type afterHandler[C any, R any] struct {
	inner decorator.CommandHandler[C, R]
	after func(context.Context, C)
}

func (h afterHandler[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	out, err := h.inner.Handle(ctx, cmd)
	if err == nil && h.after != nil {
		h.after(ctx, cmd)
	}
	return out, err
}

type loginAndWarm struct {
	inner command.LoginCommandHandler
	authz *rbac.Authorizer
}

func (h loginAndWarm) Handle(ctx context.Context, cmd command.LoginCommand) (command.LoginResult, error) {
	out, err := h.inner.Handle(ctx, cmd)
	if err == nil && h.authz != nil {
		_ = h.authz.WarmUser(ctx, out.UserID)
	}
	return out, err
}
