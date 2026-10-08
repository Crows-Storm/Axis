package schnorr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProveVerify_AcceptsMatchingKey(t *testing.T) {
	kp := GenerateKeyPair(DeriveKeyFromPassword([]byte("pw"), []byte("salt")))
	proof, err := Prove(kp, []byte("ctx"))
	require.NoError(t, err)
	assert.True(t, Verify(kp.PublicX, kp.PublicY, proof, []byte("ctx")))
}

func TestProveVerify_RejectsWrongKeyAndContext(t *testing.T) {
	kp := GenerateKeyPair(DeriveKeyFromPassword([]byte("pw"), []byte("salt")))
	wrong := GenerateKeyPair(DeriveKeyFromPassword([]byte("no"), []byte("salt")))
	proof, err := Prove(kp, []byte("ctx"))
	require.NoError(t, err)
	assert.False(t, Verify(wrong.PublicX, wrong.PublicY, proof, []byte("ctx")))
	assert.False(t, Verify(kp.PublicX, kp.PublicY, proof, []byte("other")))
}
