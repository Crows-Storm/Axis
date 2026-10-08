package command

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
)

func publishAccountEvents(ctx context.Context, bus decorator.EventBus, acc *domain.Account) error {
	if bus == nil || acc == nil || !acc.HasEvents() {
		return nil
	}
	if err := bus.Publish(ctx, acc.Events()...); err != nil {
		return decorator.CommandExecutedError{Msg: fmt.Sprintf("publish domain events: %v", err)}
	}
	acc.ClearEvents()
	return nil
}
