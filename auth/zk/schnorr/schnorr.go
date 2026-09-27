package schnorr

import (
	"errors"
	"math/big"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
)

type KeyPair struct {
	X       *big.Int
	PublicX *big.Int
	PublicY *big.Int
}

func DeriveKeyFromPassword(password, salt []byte) *big.Int {
	x := crypto.HashToScalar("zk-LOGIN-KDF-v1", salt, password)
	if x.Sign() == 0 {
		x = big.NewInt(1)
	}
	return x
}

func GenerateKeyPair(x *big.Int) *KeyPair {
	px, py := crypto.MulBase(x)
	return &KeyPair{X: crypto.Reduce(x), PublicX: px, PublicY: py}
}

type Proof struct {
	R *crypto.Point
	S *big.Int
}

func Prove(kp *KeyPair, ctx []byte) (*Proof, error) {
	if kp == nil || kp.X == nil {
		return nil, errors.New("missing key pair")
	}
	r, err := crypto.RandomScalar()
	if err != nil {
		return nil, err
	}
	rx, ry := crypto.MulBase(r)
	R := &crypto.Point{X: rx, Y: ry}

	X := (&crypto.Point{X: kp.PublicX, Y: kp.PublicY}).Bytes()
	e := crypto.HashToScalar("Schnorr-NIZK-v1", X, R.Bytes(), ctx)
	s := crypto.AddMod(r, crypto.MulMod(e, kp.X))

	return &Proof{R: R, S: s}, nil
}

func Verify(pubX, pubY *big.Int, proof *Proof, ctx []byte) bool {
	if proof == nil || proof.R == nil || proof.S == nil || pubX == nil || pubY == nil {
		return false
	}
	if !proof.R.OnCurve() || !crypto.Curve.IsOnCurve(pubX, pubY) {
		return false
	}
	X := (&crypto.Point{X: pubX, Y: pubY}).Bytes()
	e := crypto.HashToScalar("Schnorr-NIZK-v1", X, proof.R.Bytes(), ctx)

	sGx, sGy := crypto.MulBase(proof.S)
	eXx, eXy := crypto.Mul(e, pubX, pubY)
	rx, ry := crypto.Add(proof.R.X, proof.R.Y, eXx, eXy)
	return sGx.Cmp(rx) == 0 && sGy.Cmp(ry) == 0
}

var ErrInvalidProof = errors.New("invalid zk proof")
