package command

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

func publishRole(ctx context.Context, bus decorator.EventBus, current *role.Role) error {
	if bus == nil || current == nil || !current.HasEvents() {
		return nil
	}
	pending := current.Events()
	if err := bus.Publish(ctx, pending...); err != nil {
		return decorator.CommandExecutedError{Msg: fmt.Sprintf("publish domain events: %v", err)}
	}
	current.ClearEvents()
	return nil
}

type menuEvent struct {
	AggregateID uint64
	Code        string
	MenuName    string
	Status      int8
	eventName   string
}

func (m menuEvent) Name() string   { return m.eventName }
func (m menuEvent) Domain() string { return "menu" }

func publishMenu(ctx context.Context, bus decorator.EventBus, eventName string, current *menu.Menu) error {
	if bus == nil || current == nil {
		return nil
	}
	msg := menuEvent{
		AggregateID: current.ID,
		Code:        current.Code,
		MenuName:    current.Name,
		Status:      current.Status,
		eventName:   eventName,
	}
	topic, err := event.Topic(msg.Domain(), msg.Name())
	if err != nil {
		return decorator.CommandExecutedError{Msg: fmt.Sprintf("publish domain events: %v", err)}
	}
	envs := []event.Envelope[event.Message]{{
		Topic:     topic,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}}
	if err := bus.Publish(ctx, envs...); err != nil {
		return decorator.CommandExecutedError{Msg: fmt.Sprintf("publish domain events: %v", err)}
	}
	return nil
}
