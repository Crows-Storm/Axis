package event_test

import (
	"testing"

	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/stretchr/testify/require"
)

func TestTopic(t *testing.T) {
	got, err := event.Topic("user", "UserCreated")
	require.NoError(t, err)
	require.Equal(t, "axis.user.UserCreated", got)
	require.Equal(t, "axis.audit", event.AuditTopic)
}

func TestTopicRejectsEmptyOrDottedSegments(t *testing.T) {
	_, err := event.Topic("", "UserCreated")
	require.Error(t, err)
	_, err = event.Topic("user", "User Created")
	require.Error(t, err)
	_, err = event.Topic("axis.user", "UserCreated")
	require.Error(t, err)
}
