package adapter

import (
	"testing"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
	"github.com/Crows-Storm/Axis/exchange/domain"
	"github.com/stretchr/testify/require"
)

func TestSyncMarketDataRequestValidate(t *testing.T) {
	err := (SyncMarketDataRequest{}).Validate()
	require.Error(t, err)

	err = (SyncMarketDataRequest{Symbols: []string{"BTCUSDT"}}).Validate()
	require.Error(t, err)

	err = (SyncMarketDataRequest{
		Symbols:  []string{"BTCUSDT"},
		Interval: commexchange.OneHour,
	}).Validate()
	require.NoError(t, err)

	err = (SyncMarketDataRequest{
		Symbols:  []string{"BTCUSDT"},
		Channels: []domain.MarketChannel{domain.ChannelTicker},
	}).Validate()
	require.NoError(t, err)
}
