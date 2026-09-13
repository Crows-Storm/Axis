package event

import (
	"context"
	"fmt"
	"sync"
)

type EventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (d *EventDispatcher) Register(handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	name := handler.EventName()
	d.handlers[name] = append(d.handlers[name], handler)
}

func (d *EventDispatcher) Dispatch(ctx context.Context, envelope Envelope) error {
	d.mu.RLock()
	handlers, ok := d.handlers[envelope.EventName]
	d.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		return fmt.Errorf("no handler registered for event: %s", envelope.EventName)
	}

	for _, h := range handlers {
		if err := h.Handle(ctx, envelope); err != nil {
			return fmt.Errorf("handler %T failed for event %s: %w", h, envelope.EventName, err)
		}
	}
	return nil
}
