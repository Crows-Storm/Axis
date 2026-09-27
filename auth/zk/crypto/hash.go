package crypto

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
)

func HashToScalar(label string, parts ...[]byte) *big.Int {
	h := sha256.New()
	writeChunk := func(b []byte) {
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(len(b)))
		h.Write(lenBuf[:])
		h.Write(b)
	}
	writeChunk([]byte(label))
	for _, p := range parts {
		writeChunk(p)
	}
	digest := h.Sum(nil)
	h2 := sha256.New()
	h2.Write(digest)
	h2.Write([]byte{0x01})
	extended := append(digest, h2.Sum(nil)...)
	x := new(big.Int).SetBytes(extended)
	return x.Mod(x, N)
}

func HashToField(label string, parts ...[]byte) *big.Int {
	h := sha256.New()
	writeChunk := func(b []byte) {
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(len(b)))
		h.Write(lenBuf[:])
		h.Write(b)
	}
	writeChunk([]byte(label))
	for _, p := range parts {
		writeChunk(p)
	}
	x := new(big.Int).SetBytes(h.Sum(nil))
	return x.Mod(x, Curve.Params().P)
}
