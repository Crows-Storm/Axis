package domain

import (
	"fmt"

	"github.com/Crows-Storm/Axis/common/domain/event"
)

type Aggregate interface {
	AggregateName() string

	ApplyEvent(event event.DomainEvent) error
	raiseEvent(...event.DomainEvent) error
}

type AggregateRoot struct {
	Aggregate

	name    string
	version int64
	events  []event.DomainEvent
}

func (a *AggregateRoot) SetAggName(n string) {
	a.name = n
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

func (a *AggregateRoot) RaiseEvent(events ...event.DomainEvent) error {
	return a.raiseEvent(events...)
}

func (a *AggregateRoot) Events() []event.DomainEvent {
	return a.events
}

func (a *AggregateRoot) ClearEvents() {
	a.events = nil
}

func (a *AggregateRoot) HasEvents() bool {
	return len(a.events) > 0
}

func (a *AggregateRoot) FlushEvents() []event.DomainEvent {
	flushed := a.Events()
	a.events = nil
	return flushed
}

// The raiseEvent is private method
// - It performs a series of atomic operations to raise events: call the domain `ApplyEvent`, and append events to the root events array
// - Each valid event increments the version number; each event represents an operation (an operation signifies a change in the domain's state)
func (a *AggregateRoot) raiseEvent(events ...event.DomainEvent) error {
	for _, e := range events {
		if err := a.ApplyEvent(e); err != nil {
			return fmt.Errorf("aggregateRoot.BaseRoot: failed to record event, %w", err)
		}
		a.events = append(a.events, e)

		a.IncrementVersion()
	}
	return nil
}
