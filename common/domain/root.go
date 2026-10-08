package domain

import (
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/domain/event"
)

type Aggregate interface {
	RaiseEvent(self Applier, e ...event.Message) error
	HasEvents() bool
}

type Applier interface {
	ApplyEvent(e event.Message) error
}

type AggregateRoot struct {
	version int64
	events  []event.Envelope[event.Message]
}

func (a *AggregateRoot) Version() int64 {
	return a.version
}

func (a *AggregateRoot) SetVersion(version int64) {
	a.version = version
}

func (a *AggregateRoot) IncrementVersion() {
	a.version++
}

// The RaiseEvent apply and append events and increment version
//
// Possible variations:
// However, RaiseEvent requires "self" as an input
// The AggregateRoot can adds {self Applier}
// When constructing the user, Bind(u) to AggregateRoot.self
func (a *AggregateRoot) RaiseEvent(self Applier, events ...event.Message) error {
	for _, message := range events {
		if message == nil {
			return fmt.Errorf("aggregate root: nil event")
		}
		topic, err := event.Topic(message.Domain(), message.Name())
		if err != nil {
			return fmt.Errorf("aggregate root: %w", err)
		}
		if err := self.ApplyEvent(message); err != nil {
			return fmt.Errorf("aggregateRoot.BaseRoot: failed to audit event, %w", err)
		}
		a.events = append(a.events, event.Envelope[event.Message]{
			Topic:     topic,
			Message:   message,
			Timestamp: time.Now(),
		})

		a.IncrementVersion()
	}
	return nil
}

func (a *AggregateRoot) Events() []event.Envelope[event.Message] {
	return a.events
}

func (a *AggregateRoot) ClearEvents() {
	a.events = nil
}

func (a *AggregateRoot) HasEvents() bool {
	return len(a.events) > 0
}

func (a *AggregateRoot) FlushEvents() []event.Envelope[event.Message] {
	flushed := a.Events()
	a.events = nil
	return flushed
}
