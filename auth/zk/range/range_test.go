package rangepf

import (
	"math/big"
	"testing"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
	"github.com/Crows-Storm/Axis/auth/zk/pedersen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProveBit_ZeroAndOne(t *testing.T) {
	r, err := crypto.RandomScalar()
	require.NoError(t, err)

	c0 := pedersen.Commit(big.NewInt(0), r)
	p0, err := ProveBit(c0, big.NewInt(0), r)
	require.NoError(t, err)
	assert.True(t, VerifyBit(p0))

	c1 := pedersen.Commit(big.NewInt(1), r)
	p1, err := ProveBit(c1, big.NewInt(1), r)
	require.NoError(t, err)
	assert.True(t, VerifyBit(p1))

	assert.False(t, VerifyBit(p0) && p0.C.X.Cmp(c1.X) == 0)
}
