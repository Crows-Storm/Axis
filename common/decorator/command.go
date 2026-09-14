package decorator

import (
	"context"
)

type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (result R, err error)
}

func ApplyCommandDecorators[C, R any](handler CommandHandler[C, R], metricsClient MetricsClient) CommandHandler[C, R] {
	return queryLoggingDecorator[C, R]{
		base: queryMetricsDecorator[C, R]{
			base: commandValidateDecorator[C, R]{ // command validate decorator
				base: handler,
			},
			client: metricsClient,
		},
	}
}

type CommandExecutedError struct {
	Msg string
}

// CommandExecutedError implement Error interface
func (c CommandExecutedError) Error() string {
	return c.Msg
}

type CommandValidator interface {
	Validate() error
}

type commandValidateDecorator[C, R any] struct {
	base CommandHandler[C, R]
}

func (d commandValidateDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var zero R
	if v, ok := any(cmd).(CommandValidator); ok {
		// Check the validity of the command in advance
		if err := v.Validate(); err != nil {
			return zero, InvalidCommand
		}
	}
	return d.base.Handle(ctx, cmd)
}
