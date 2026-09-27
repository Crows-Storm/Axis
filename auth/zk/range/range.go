package rangepf

import (
	"errors"
	"math/big"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
	"github.com/Crows-Storm/Axis/auth/zk/pedersen"
)

type Point struct {
	X, Y *big.Int
}

type BitProof struct {
	C  *Point
	T0 *Point
	T1 *Point
	E0 *big.Int
	E1 *big.Int
	Z0 *big.Int
	Z1 *big.Int
}

func ProveBit(C *pedersen.Commitment, b *big.Int, r *big.Int) (*BitProof, error) {
	if C == nil || r == nil {
		return nil, errors.New("missing commitment")
	}
	if b.Sign() != 0 && b.Cmp(big.NewInt(1)) != 0 {
		return nil, errors.New("bit must be 0 or 1")
	}

	k, err := crypto.RandomScalar()
	if err != nil {
		return nil, err
	}
	eFake, err := crypto.RandomScalar()
	if err != nil {
		return nil, err
	}
	zFake, err := crypto.RandomScalar()
	if err != nil {
		return nil, err
	}

	cPt := &Point{X: C.X, Y: C.Y}
	cMinusG := subPoints(cPt, basePoint())

	var t0, t1 *Point
	var e0, e1, z0, z1 *big.Int

	if b.Sign() == 0 {
		t0 = mulH(k)
		t1 = subPoints(mulH(zFake), scalePoint(eFake, cMinusG))
		e := challenge(t0, t1, cPt)
		e1 = eFake
		e0 = crypto.SubMod(e, e1)
		z0 = crypto.AddMod(k, crypto.MulMod(e0, r))
		z1 = zFake
	} else {
		t1 = mulH(k)
		t0 = subPoints(mulH(zFake), scalePoint(eFake, cPt))
		e := challenge(t0, t1, cPt)
		e0 = eFake
		e1 = crypto.SubMod(e, e0)
		z1 = crypto.AddMod(k, crypto.MulMod(e1, r))
		z0 = zFake
	}

	return &BitProof{C: cPt, T0: t0, T1: t1, E0: e0, E1: e1, Z0: z0, Z1: z1}, nil
}

func VerifyBit(p *BitProof) bool {
	if p == nil || p.C == nil || p.T0 == nil || p.T1 == nil || p.E0 == nil || p.E1 == nil || p.Z0 == nil || p.Z1 == nil {
		return false
	}
	if !onCurve(p.C) || !onCurve(p.T0) || !onCurve(p.T1) {
		return false
	}
	e := challenge(p.T0, p.T1, p.C)
	if crypto.AddMod(p.E0, p.E1).Cmp(e) != 0 {
		return false
	}
	left0 := subPoints(mulH(p.Z0), scalePoint(p.E0, p.C))
	if !pointsEqual(left0, p.T0) {
		return false
	}
	cMinusG := subPoints(p.C, basePoint())
	left1 := subPoints(mulH(p.Z1), scalePoint(p.E1, cMinusG))
	return pointsEqual(left1, p.T1)
}

func challenge(t0, t1, c *Point) *big.Int {
	return crypto.HashToScalar("RangeBit-OR-v1",
		(&crypto.Point{X: t0.X, Y: t0.Y}).Bytes(),
		(&crypto.Point{X: t1.X, Y: t1.Y}).Bytes(),
		(&crypto.Point{X: c.X, Y: c.Y}).Bytes(),
	)
}

func basePoint() *Point {
	gx, gy := crypto.MulBase(big.NewInt(1))
	return &Point{X: gx, Y: gy}
}

func mulH(k *big.Int) *Point {
	x, y := crypto.Mul(k, pedersen.Hx, pedersen.Hy)
	return &Point{X: x, Y: y}
}

func scalePoint(k *big.Int, p *Point) *Point {
	x, y := crypto.Mul(k, p.X, p.Y)
	return &Point{X: x, Y: y}
}

func subPoints(a, b *Point) *Point {
	x, y := crypto.Sub(a.X, a.Y, b.X, b.Y)
	return &Point{X: x, Y: y}
}

func onCurve(p *Point) bool {
	return p != nil && p.X != nil && p.Y != nil && crypto.Curve.IsOnCurve(p.X, p.Y)
}

func pointsEqual(a, b *Point) bool {
	return a != nil && b != nil && a.X.Cmp(b.X) == 0 && a.Y.Cmp(b.Y) == 0
}
