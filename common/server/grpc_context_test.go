package server

import (
	"context"
	"errors"
	"testing"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

type stubParser struct {
	principal *principal.Principal
	err       error
}

func (s stubParser) Parse(ctx context.Context, accessToken string) (*principal.Principal, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.principal, nil
}

func (s stubParser) Preprocessing(ctx context.Context, token string) (string, error) {
	return token, nil
}

func TestRestoreInboundContext_GeneratesSpanAndLocalInstance(t *testing.T) {
	md := metadata.New(map[string]string{
		contextx.MetadataRequestID:     "req-grpc",
		contextx.MetadataCorrelationID: "corr-grpc",
		contextx.MetadataCausationId:   "forged-cause",
		contextx.MetadataParentSpanID:  "parent-span",
		contextx.MetadataAuthorization: "Bearer tok",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	got, err := restoreInboundContext(ctx, stubParser{principal: &principal.Principal{UserID: 7}}, "svc-b")
	require.NoError(t, err)
	reader := contextx.FromContext(got)

	assert.Equal(t, "req-grpc", reader.RequestID())
	assert.Equal(t, "corr-grpc", reader.CorrelationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.NotEqual(t, "forged-cause", reader.CausationID())
	assert.Equal(t, "parent-span", reader.ParentSpanID())
	assert.NotEqual(t, "parent-span", reader.SpanID())
	assert.Equal(t, "svc-b", reader.InstanceID())
	require.NotNil(t, reader.Principal())
	assert.Equal(t, uint64(7), reader.UserID())
}

func TestRestoreInboundContext_InvalidTokenDoesNotFail(t *testing.T) {
	md := metadata.New(map[string]string{
		contextx.MetadataRequestID:     "req-grpc",
		contextx.MetadataAuthorization: "Bearer bad",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	got, err := restoreInboundContext(ctx, stubParser{err: errors.New("invalid")}, "svc-b")
	require.NoError(t, err)
	assert.Nil(t, contextx.FromContext(got).Principal())
	assert.Equal(t, "req-grpc", contextx.FromContext(got).RequestID())
}

func TestRestoreInboundContext_MissingRequestIDRejected(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	_, err := restoreInboundContext(ctx, nil, "svc-b")
	require.Error(t, err)
}
