package decorator

import (
	"context"

	"github.com/Crows-Storm/Axis/common/domain/event"
)

// EventBus publishes domain messages raised by an aggregate.
type EventBus interface {
	Publish(ctx context.Context, events ...event.Envelope[event.Message]) error
	Close() error
}
