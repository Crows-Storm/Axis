package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

func publishUserEvents(ctx context.Context, bus decorator.EventBus, u *domain.User) error {
	if bus == nil || u == nil || !u.HasEvents() {
		return nil
	}
	if err := bus.Publish(ctx, u.Events()...); err != nil {
		return decorator.CommandExecutedError{Msg: fmt.Sprintf("publish domain events: %v", err)}
	}
	u.ClearEvents()
	return nil
}
