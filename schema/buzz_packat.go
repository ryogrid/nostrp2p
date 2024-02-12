package schema

import (
	"github.com/weaveworks/mesh"
)

// BuzzPacket is an implementation of a G-counter.
type BuzzPacket struct {
	Event *BuzzEvent
	Req   *BuzzReq
	Resp  *BuzzResp
}

// BuzzPacket implements GossipData.
var _ mesh.GossipData = &BuzzPacket{}

// Construct an empty BuzzPacket object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func newPacket() *BuzzPacket {
	// TODO: need to implement (newPacket func in packet.go)
	panic("not implemented yet")
}

// Encode serializes BuzzPacket to a slice of byte-slices.
func (st *BuzzPacket) Encode() [][]byte {
	// TODO: need to implement (BuzzPacket::Encode)
	panic("not implemented yet")
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete BuzzPacket.
func (st *BuzzPacket) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	// TODO: need to implement (BuzzPacket::Merge)
	panic("not implemented yet")
}
