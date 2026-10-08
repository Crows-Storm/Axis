package domain_test

import (
	"testing"

	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
	"github.com/stretchr/testify/require"
)

type recordingRoot struct {
	domain.AggregateRoot
	applied int
}

func (r *recordingRoot) ApplyEvent(event.Message) error {
	r.applied++
	return nil
}

func TestRaiseEventSetsTopicBeforeApply(t *testing.T) {
	root := &recordingRoot{}
	msg := events.NewUserCreatedMessage(42, "login", "a@b.c", "alice")
	require.NoError(t, root.RaiseEvent(root, msg))
	require.Equal(t, 1, root.applied)
	require.Equal(t, "axis.user.UserCreated", root.Events()[0].Topic)
}

type badMessage struct{}

func (badMessage) Name() string   { return "Has Space" }
func (badMessage) Domain() string { return "user" }

func TestRaiseEventRejectsInvalidTopicWithoutApply(t *testing.T) {
	root := &recordingRoot{}
	err := root.RaiseEvent(root, badMessage{})
	require.Error(t, err)
	require.Zero(t, root.applied)
	require.Empty(t, root.Events())
}
