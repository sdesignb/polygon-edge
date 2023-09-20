package types

import (
	"math/big"
	"sync/atomic"

	"github.com/sdesignb/polygon-edge/helper/common"
	"github.com/sdesignb/polygon-edge/helper/keccak"
)

type Transaction struct {
	Nonce     uint64
	GasPrice  *big.Int
	GasTipCap *big.Int
	GasFeeCap *big.Int
	Gas       uint64
	To        *Address
	Value     *big.Int
	Input     []byte
	V         *big.Int
	R         *big.Int
	S         *big.Int
	Hash      Hash
	From      Address

	// Cache
	size atomic.Value
}

func (t *Transaction) IsContractCreation() bool {
	return t.To == nil
}

// ComputeHash computes the hash of the transaction
func (t *Transaction) ComputeHash() *Transaction {
	ar := marshalArenaPool.Get()
	hash := keccak.DefaultKeccakPool.Get()

	v := t.MarshalRLPWith(ar)
	hash.WriteRlp(t.Hash[:0], v)

	marshalArenaPool.Put(ar)
	keccak.DefaultKeccakPool.Put(hash)

	return t
}

func (t *Transaction) Copy() *Transaction {
	tt := new(Transaction)
	*tt = *t

	tt.GasPrice = new(big.Int)
	if t.GasPrice != nil {
		tt.GasPrice.Set(t.GasPrice)
	}

	tt.Value = new(big.Int)
	if t.Value != nil {
		tt.Value.Set(t.Value)
	}

	if t.R != nil {
		tt.R = new(big.Int)
		tt.R = big.NewInt(0).SetBits(t.R.Bits())
	}

	if t.S != nil {
		tt.S = new(big.Int)
		tt.S = big.NewInt(0).SetBits(t.S.Bits())
	}

	tt.Input = make([]byte, len(t.Input))
	copy(tt.Input[:], t.Input[:])

	return tt
}

// Cost returns gas * gasPrice + value
func (t *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(t.GasPrice, new(big.Int).SetUint64(t.Gas))
	total.Add(total, t.Value)

	return total
}

// GetGasPrice returns gas price if not empty, or calculates one based on
// the given EIP-1559 fields if exist
//
// Here is the logic:
//   - use existing gas price if exists
//   - or calculate a value with formula: min(gasFeeCap, gasTipCap + baseFee);
func (t *Transaction) GetGasPrice(baseFee uint64) *big.Int {
	if t.GasPrice != nil && t.GasPrice.BitLen() > 0 {
		return new(big.Int).Set(t.GasPrice)
	} else if baseFee == 0 {
		return big.NewInt(0)
	}

	gasFeeCap := new(big.Int)
	if t.GasFeeCap != nil {
		gasFeeCap = gasFeeCap.Set(t.GasFeeCap)
	}

	gasTipCap := new(big.Int)
	if t.GasTipCap != nil {
		gasTipCap = gasTipCap.Set(t.GasTipCap)
	}

	if gasFeeCap.BitLen() > 0 || gasTipCap.BitLen() > 0 {
		return common.BigMin(
			gasTipCap.Add(
				gasTipCap,
				new(big.Int).SetUint64(baseFee),
			),
			gasFeeCap,
		)
	}

	return big.NewInt(0)
}

func (t *Transaction) Size() uint64 {
	if size := t.size.Load(); size != nil {
		sizeVal, ok := size.(uint64)
		if !ok {
			return 0
		}

		return sizeVal
	}

	size := uint64(len(t.MarshalRLP()))
	t.size.Store(size)

	return size
}

func (t *Transaction) ExceedsBlockGasLimit(blockGasLimit uint64) bool {
	return t.Gas > blockGasLimit
}

func (t *Transaction) IsUnderpriced(priceLimit uint64) bool {
	return t.GasPrice.Cmp(big.NewInt(0).SetUint64(priceLimit)) < 0
}
