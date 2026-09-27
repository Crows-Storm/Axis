package pedersen

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
)

var Hx, Hy = hashToCurve("Pedersen-H-v1")

func hashToCurve(seed string) (*big.Int, *big.Int) {
	for ctr := uint32(1); ctr < 10000; ctr++ {
		var ctrBuf [4]byte
		binary.BigEndian.PutUint32(ctrBuf[:], ctr)
		x := crypto.HashToField("H2C-x", []byte(seed), ctrBuf[:])
		px, py := decompressPoint(x)
		if px != nil && crypto.Curve.IsOnCurve(px, py) {
			return px, py
		}
	}
	panic("hash-to-curve failed")
}

func decompressPoint(x *big.Int) (*big.Int, *big.Int) {
	p := crypto.Curve.Params().P
	x3 := new(big.Int).Exp(x, big.NewInt(3), p)
	threeX := new(big.Int).Mul(big.NewInt(3), x)
	threeX.Mod(threeX, p)
	rhs := new(big.Int).Sub(x3, threeX)
	rhs.Add(rhs, crypto.Curve.Params().B)
	rhs.Mod(rhs, p)

	exp := new(big.Int).Add(p, big.NewInt(1))
	exp.Rsh(exp, 2)
	y := new(big.Int).Exp(rhs, exp, p)
	check := new(big.Int).Mul(y, y)
	check.Mod(check, p)
	if check.Cmp(rhs) != 0 {
		return nil, nil
	}
	return new(big.Int).Set(x), y
}

type Commitment struct {
	X, Y *big.Int
	R    *big.Int
}

func Commit(v, r *big.Int) *Commitment {
	vGx, vGy := crypto.MulBase(v)
	rHx, rHy := crypto.Mul(r, Hx, Hy)
	cx, cy := crypto.Add(vGx, vGy, rHx, rHy)
	return &Commitment{X: cx, Y: cy, R: new(big.Int).Set(crypto.Reduce(r))}
}

func Verify(c *Commitment, v, r *big.Int) bool {
	if c == nil {
		return false
	}
	expected := Commit(v, r)
	return expected.X.Cmp(c.X) == 0 && expected.Y.Cmp(c.Y) == 0
}

func Add(c1, c2 *Commitment) *Commitment {
	x, y := crypto.Add(c1.X, c1.Y, c2.X, c2.Y)
	var r *big.Int
	if c1.R != nil && c2.R != nil {
		r = crypto.AddMod(c1.R, c2.R)
	}
	return &Commitment{X: x, Y: y, R: r}
}

func Scale(c *Commitment, k *big.Int) *Commitment {
	x, y := crypto.Mul(k, c.X, c.Y)
	var r *big.Int
	if c.R != nil {
		r = crypto.MulMod(c.R, k)
	}
	return &Commitment{X: x, Y: y, R: r}
}

var ErrMismatch = errors.New("commitment mismatch")
