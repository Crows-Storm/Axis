package decorator

import (
	"context"
)

type CommandValidator interface {
	Validate() error
}

type commandValidateDecorator[C, R any] struct {
	base CommandHandler[C, R]
}

func (d commandValidateDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var zero R
	if v, ok := any(cmd).(CommandValidator); ok {
		if err := v.Validate(); err != nil {
			return zero, err
		}
	}
	return d.base.Handle(ctx, cmd)
}
