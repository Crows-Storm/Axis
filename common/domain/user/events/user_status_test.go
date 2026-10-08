package events_test

import (
	"testing"

	domain "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
	"github.com/stretchr/testify/assert"
)

func TestNewUserStatusMessage(t *testing.T) {
	msg := events.NewUserStatusMessage(9, domain.Disabled)
	assert.Equal(t, "UserStatusChange", msg.Name())
	assert.Equal(t, uint64(9), msg.AggregateID)
	assert.Equal(t, domain.Disabled, msg.Status)
}
