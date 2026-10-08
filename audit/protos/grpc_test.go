package protos

import (
	"errors"
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToProtoMapsRecord(t *testing.T) {
	occurred := time.Date(2026, 9, 30, 8, 0, 0, 0, time.UTC)
	got := toProto(&audit.AuditEvent{
		ID:         "evt-1",
		EventID:    "evt-1",
		EventName:  "UserCreated",
		OperatorID: 9,
		Category:   "CREATED",
		Timestamp:  occurred,
		Metadata: audit.Metadata{
			Service:       "user",
			AggregateID:   "7",
			CorrelationId: "corr",
			CausationId:   "cause",
			Payload:       []byte(`{"n":1}`),
		},
	})
	require.NotNil(t, got)
	assert.Equal(t, "evt-1", got.EventId)
	assert.Equal(t, uint64(9), got.OperatorId)
	assert.Equal(t, occurred.Format(time.RFC3339Nano), got.Timestamp)
	assert.Equal(t, "corr", got.Metadata.CorrelationId)
	assert.Equal(t, []byte(`{"n":1}`), got.Metadata.Payload)
}

func TestStatusFromMapsDomainErrors(t *testing.T) {
	err := statusFrom(decorator.CommandExecutedError{Msg: "event_id is required"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	err = statusFrom(pagination.ErrInvalidCursor)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	err = statusFrom(audit.ErrNotFound)
	assert.Equal(t, codes.NotFound, status.Code(err))

	err = statusFrom(errors.New("mongo down"))
	assert.Equal(t, codes.Internal, status.Code(err))
}
