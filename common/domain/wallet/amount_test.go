package wallet_test

import (
	"math/big"
	"testing"

	"github.com/Crows-Storm/Axis/common/domain/wallet"
	"github.com/stretchr/testify/require"
)

func TestFormatDecimal(t *testing.T) {
	require.Equal(t, "0", wallet.FormatDecimal(0))
	require.Equal(t, "1.5", wallet.FormatDecimal(150_000_000))
	require.Equal(t, "0.05", wallet.FormatDecimal(5_000_000))
	require.Equal(t, "-1.5", wallet.FormatDecimal(-150_000_000))
}

func TestFormatMinor(t *testing.T) {
	require.Equal(t, "0", wallet.FormatMinor(nil, 2))
	require.Equal(t, "1.5", wallet.FormatMinor(big.NewInt(150), 2))
	require.Equal(t, "1", wallet.FormatMinor(big.NewInt(1_000_000_000_000_000_000), 18))
	require.Equal(t, "100", wallet.FormatMinor(big.NewInt(100), 0))
}

func TestParseCurrencyAndStatus(t *testing.T) {
	currency, ok := wallet.ParseCurrency("usd")
	require.True(t, ok)
	require.Equal(t, wallet.USD, currency)
	_, ok = wallet.ParseCurrency("DOGE")
	require.False(t, ok)

	status, ok := wallet.ParseStatus("frozen")
	require.True(t, ok)
	require.Equal(t, wallet.Frozen, status)
	_, ok = wallet.ParseType("margin")
	require.False(t, ok)
}
