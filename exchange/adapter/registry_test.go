package adapter

import (
	"testing"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/stretchr/testify/require"
)

type stubPlatform struct {
	ExchangeClient
	platform commexchange.Platform
}

func (s stubPlatform) Platform() commexchange.Platform { return s.platform }

func TestRegistryClient(t *testing.T) {
	r := NewRegistry(stubPlatform{platform: commexchange.Binance})
	c, err := r.Client("binance")
	require.NoError(t, err)
	require.Equal(t, commexchange.Binance, c.Platform())
	_, err = r.Client("okx")
	require.Error(t, err)
	_, err = r.Client("")
	require.Error(t, err)
}
