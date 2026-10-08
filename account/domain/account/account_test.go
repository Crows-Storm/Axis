package account_test

import (
	"testing"

	"github.com/Crows-Storm/Axis/account/domain/account"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
	"github.com/stretchr/testify/require"
)

func TestOpenAccountAndWallet(t *testing.T) {
	acc, err := account.Open(42, commwallet.Spot)
	require.NoError(t, err)
	require.Equal(t, uint64(42), acc.UserID)
	require.Equal(t, commwallet.Spot, acc.Type)
	require.Equal(t, "axis.account.AccountOpened", acc.Events()[0].Topic)

	opened, err := acc.OpenWallet(commaccount.Currency{ID: 10, Symbol: "USDT", Decimals: 6, Class: commaccount.AssetClassCrypto, Name: "Tether USD", SmallestUnit: "micro"})
	require.NoError(t, err)
	require.Equal(t, "USDT", opened.Currency.Symbol)
	require.Equal(t, "0", opened.Balance.String())
	require.Len(t, acc.Wallet, 1)

	_, err = acc.OpenWallet(commaccount.Currency{ID: 10, Symbol: "USDT"})
	require.ErrorIs(t, err, commaccount.ErrInvalid)
}

func TestOpenRejectsInvalid(t *testing.T) {
	_, err := account.Open(0, commwallet.Spot)
	require.ErrorIs(t, err, commaccount.ErrInvalid)
	_, err = account.Open(1, "margin")
	require.ErrorIs(t, err, commaccount.ErrInvalid)
}
