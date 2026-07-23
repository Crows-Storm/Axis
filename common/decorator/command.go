package decorator

import "context"

type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (result R, err error)
}
