package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

var Curve = elliptic.P256()
var N = Curve.Params().N

type Scalar = big.Int

func RandomScalar() (*big.Int, error) {
	return RandomScalarFrom(rand.Reader)
}

func RandomScalarFrom(r io.Reader) (*big.Int, error) {
	for {
		k, err := rand.Int(r, N)
		if err != nil {
			return nil, err
		}
		if k.Sign() != 0 {
			return k, nil
		}
	}
}

func Reduce(k *big.Int) *big.Int {
	if k == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Mod(k, N)
}

func MulBase(k *big.Int) (x, y *big.Int) {
	return Curve.ScalarBaseMult(ScalarBytes(Reduce(k)))
}

func Mul(k *big.Int, px, py *big.Int) (x, y *big.Int) {
	return Curve.ScalarMult(px, py, ScalarBytes(Reduce(k)))
}

func Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	return Curve.Add(x1, y1, x2, y2)
}

func Neg(px, py *big.Int) (*big.Int, *big.Int) {
	p := Curve.Params().P
	ny := new(big.Int).Neg(py)
	ny.Mod(ny, p)
	return new(big.Int).Set(px), ny
}

func Sub(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	nx, ny := Neg(x2, y2)
	return Add(x1, y1, nx, ny)
}

type Point struct {
	X, Y *big.Int
}

func (p *Point) Bytes() []byte {
	if p == nil || p.X == nil || p.Y == nil {
		return []byte{0}
	}
	buf := make([]byte, 1+2*32)
	buf[0] = 0x04
	p.X.FillBytes(buf[1:33])
	p.Y.FillBytes(buf[33:65])
	return buf
}

func (p *Point) OnCurve() bool {
	if p == nil || p.X == nil || p.Y == nil {
		return false
	}
	return Curve.IsOnCurve(p.X, p.Y)
}

func ScalarBytes(s *big.Int) []byte {
	b := make([]byte, 32)
	Reduce(s).FillBytes(b)
	return b
}

func AddMod(a, b *big.Int) *big.Int {
	c := new(big.Int).Add(Reduce(a), Reduce(b))
	return c.Mod(c, N)
}

func SubMod(a, b *big.Int) *big.Int {
	c := new(big.Int).Sub(Reduce(a), Reduce(b))
	return c.Mod(c, N)
}

func MulMod(a, b *big.Int) *big.Int {
	c := new(big.Int).Mul(Reduce(a), Reduce(b))
	return c.Mod(c, N)
}

func InvMod(a *big.Int) (*big.Int, error) {
	if Reduce(a).Sign() == 0 {
		return nil, errors.New("division by zero")
	}
	inv := new(big.Int).ModInverse(Reduce(a), N)
	if inv == nil {
		return nil, errors.New("no inverse")
	}
	return inv, nil
}

func NegMod(a *big.Int) *big.Int {
	if Reduce(a).Sign() == 0 {
		return big.NewInt(0)
	}
	return new(big.Int).Sub(N, Reduce(a))
}
