package schema

import (
	"github.com/holiman/uint256"
	"math/big"
)

type BuzzEvent struct {
	Id         uint64              // random 64bit uint value to identify the event (Not sha256 32bytes hash)
	Pubkey     [32]byte            // encoded 256bit uint (holiman/uint256)
	Created_at int64               // unix timestamp in seconds
	Kind       uint16              // integer between 0 and 65535
	Tags       map[string][]string // Key: tag string, Value: other strings
	Content    string
	Sig        [64]byte // 64-bytes integr of the signature of the sha256 hash of the serialized event data
}

func (e *BuzzEvent) GetPubkey() *big.Int {
	fixed256 := uint256.NewInt(0)
	return fixed256.SetBytes(e.Pubkey[:]).ToBig()
}

func (e *BuzzEvent) SetPubkey(pubkey *big.Int) {
	fixed256, isOverflow := uint256.FromBig(pubkey)
	if isOverflow {
		panic("overflow")
	}
	fixed256.WriteToArray32(&e.Pubkey)
}
