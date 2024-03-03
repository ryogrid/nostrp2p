package schema

import (
	"github.com/holiman/uint256"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/vmihailenco/msgpack/v5"
	"math/big"
)

type Np2pEvent struct {
	Id         uint64                      // random 64bit uint value to identify the event (Not sha256 32bytes hash)
	Pubkey     [np2p_const.PubkeySize]byte // encoded 256bit uint (holiman/uint256)
	Created_at int64                       // unix timestamp in seconds
	Kind       uint16                      // integer between 0 and 65535
	Tags       map[string][]interface{}    // Key: tag string, Value: any type
	Content    string
	Sig        *[np2p_const.SignatureSize]byte // 64-bytes integr of the signature of the sha256 hash of the serialized event data
}

func (e *Np2pEvent) GetPubkey() *big.Int {
	fixed256 := uint256.NewInt(0)
	return fixed256.SetBytes(e.Pubkey[:]).ToBig()
}

func (e *Np2pEvent) SetPubkey(pubkey *big.Int) {
	fixed256, isOverflow := uint256.FromBig(pubkey)
	if isOverflow {
		panic("overflow")
	}
	fixed256.WriteToArray32(&e.Pubkey)
}

func (e *Np2pEvent) Encode() []byte {
	b, err := msgpack.Marshal(e)
	if err != nil {
		panic(err)
	}
	return b
}

func NewNp2pEventFromBytes(b []byte) (*Np2pEvent, error) {
	var e Np2pEvent
	if err := msgpack.Unmarshal(b, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
