package permission

import (
	"errors"
	"math/big"

	"github.com/Crows-Storm/Axis/auth/zk/crypto"
	"github.com/Crows-Storm/Axis/auth/zk/pedersen"
	rangepf "github.com/Crows-Storm/Axis/auth/zk/range"
)

const bitLen = 64

type PermissionProof struct {
	Commitment     *pedersen.Commitment
	Delta          *pedersen.Commitment
	RangeBits      []*rangepf.BitProof
	BitCommitments []*pedersen.Commitment
}

func ProvePermissionAbove(v, threshold *big.Int, r *big.Int) (*PermissionProof, error) {
	if v == nil || threshold == nil || r == nil {
		return nil, errors.New("missing arguments")
	}
	if v.Sign() < 0 || threshold.Sign() < 0 {
		return nil, errors.New("negative values are not supported")
	}
	if v.BitLen() > bitLen || threshold.BitLen() > bitLen {
		return nil, errors.New("value exceeds supported bit length")
	}
	if v.Cmp(threshold) < 0 {
		return nil, errors.New("permission below threshold")
	}

	C := pedersen.Commit(v, r)
	diff := new(big.Int).Sub(v, threshold)
	D := pedersen.Commit(diff, r)

	bits := toBits(diff, bitLen)
	blinds, err := splitBlind(r, bitLen)
	if err != nil {
		return nil, err
	}

	bitComms := make([]*pedersen.Commitment, bitLen)
	bitProofs := make([]*rangepf.BitProof, bitLen)
	for i, b := range bits {
		Ci := pedersen.Commit(big.NewInt(int64(b)), blinds[i])
		bp, err := rangepf.ProveBit(Ci, big.NewInt(int64(b)), blinds[i])
		if err != nil {
			return nil, err
		}
		bitComms[i] = Ci
		bitProofs[i] = bp
	}

	return &PermissionProof{
		Commitment:     C,
		Delta:          D,
		RangeBits:      bitProofs,
		BitCommitments: bitComms,
	}, nil
}

func VerifyPermissionAbove(proof *PermissionProof, threshold *big.Int) bool {
	if proof == nil || proof.Commitment == nil || proof.Delta == nil || threshold == nil {
		return false
	}
	if len(proof.RangeBits) != bitLen || len(proof.BitCommitments) != bitLen {
		return false
	}

	Tx, Ty := crypto.MulBase(threshold)
	dx, dy := crypto.Sub(proof.Commitment.X, proof.Commitment.Y, Tx, Ty)
	if dx.Cmp(proof.Delta.X) != 0 || dy.Cmp(proof.Delta.Y) != 0 {
		return false
	}

	var sumX, sumY *big.Int
	for i, Ci := range proof.BitCommitments {
		if Ci == nil {
			return false
		}
		bp := proof.RangeBits[i]
		if bp == nil || bp.C == nil || bp.C.X.Cmp(Ci.X) != 0 || bp.C.Y.Cmp(Ci.Y) != 0 {
			return false
		}
		if !rangepf.VerifyBit(bp) {
			return false
		}
		weight := new(big.Int).Lsh(big.NewInt(1), uint(i))
		wx, wy := crypto.Mul(weight, Ci.X, Ci.Y)
		if i == 0 {
			sumX, sumY = wx, wy
			continue
		}
		sumX, sumY = crypto.Add(sumX, sumY, wx, wy)
	}
	return sumX.Cmp(proof.Delta.X) == 0 && sumY.Cmp(proof.Delta.Y) == 0
}

func toBits(v *big.Int, n int) []int {
	bits := make([]int, n)
	for i := 0; i < n; i++ {
		bits[i] = int(v.Bit(i))
	}
	return bits
}

func splitBlind(r *big.Int, n int) ([]*big.Int, error) {
	blinds := make([]*big.Int, n)
	acc := big.NewInt(0)
	for i := 0; i < n-1; i++ {
		ri, err := crypto.RandomScalar()
		if err != nil {
			return nil, err
		}
		blinds[i] = ri
		weighted := crypto.MulMod(ri, new(big.Int).Lsh(big.NewInt(1), uint(i)))
		acc = crypto.AddMod(acc, weighted)
	}
	pow := new(big.Int).Lsh(big.NewInt(1), uint(n-1))
	inv, err := crypto.InvMod(pow)
	if err != nil {
		return nil, err
	}
	last := crypto.MulMod(crypto.SubMod(r, acc), inv)
	blinds[n-1] = last
	return blinds, nil
}
