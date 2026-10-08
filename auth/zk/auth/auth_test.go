package auth

import (
	"math/big"
	"testing"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
	"github.com/Crows-Storm/Axis/auth/zk/schnorr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndLogin(t *testing.T) {
	store := NewUserStore()
	salt := []byte("salt-salt-salt-1")
	kp := schnorr.GenerateKeyPair(schnorr.DeriveKeyFromPassword([]byte("secret"), salt))
	require.NoError(t, store.Register("alice", salt, kp.PublicX.Bytes(), kp.PublicY.Bytes()))

	nonce, err := store.IssueChallenge("alice")
	require.NoError(t, err)
	proof, err := schnorr.Prove(kp, append([]byte("login:alice:"), nonce...))
	require.NoError(t, err)
	require.NoError(t, store.VerifyLogin("alice", proof))

	nonce2, err := store.IssueChallenge("alice")
	require.NoError(t, err)
	wrong := schnorr.GenerateKeyPair(schnorr.DeriveKeyFromPassword([]byte("nope"), salt))
	bad, err := schnorr.Prove(wrong, append([]byte("login:alice:"), nonce2...))
	require.NoError(t, err)
	assert.ErrorIs(t, store.VerifyLogin("alice", bad), ErrInvalidProof)
}

func TestPermissionAbove(t *testing.T) {
	r, err := crypto.RandomScalar()
	require.NoError(t, err)
	proof, err := ProvePermissionAbove(big.NewInt(850), big.NewInt(500), r)
	require.NoError(t, err)
	assert.True(t, VerifyPermissionAbove(proof, big.NewInt(500)))
	assert.False(t, VerifyPermissionAbove(proof, big.NewInt(900)))

	_, err = ProvePermissionAbove(big.NewInt(100), big.NewInt(500), r)
	assert.Error(t, err)
}
