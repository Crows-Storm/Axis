package crypto

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashToScalar_NotAlwaysZero(t *testing.T) {
	a := HashToScalar("lab", []byte("a"))
	b := HashToScalar("lab", []byte("b"))
	assert.NotEqual(t, 0, a.Sign())
	assert.NotEqual(t, 0, b.Sign())
	assert.NotEqual(t, 0, a.Cmp(b))
	assert.True(t, a.Cmp(N) < 0)
}

func TestPointSubAddRoundTrip(t *testing.T) {
	k, err := RandomScalar()
	require.NoError(t, err)
	x1, y1 := MulBase(k)
	x2, y2 := MulBase(big.NewInt(3))
	sx, sy := Sub(x1, y1, x2, y2)
	ax, ay := Add(sx, sy, x2, y2)
	assert.Equal(t, 0, ax.Cmp(x1))
	assert.Equal(t, 0, ay.Cmp(y1))
}
