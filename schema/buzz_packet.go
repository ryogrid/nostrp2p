package schema

import (
	"bytes"
	"encoding/gob"
	"github.com/weaveworks/mesh"
)

// BuzzPacket is an implementation of GossipData
type BuzzPacket struct {
	Events []*BuzzEvent
	Req    *BuzzReq
	Resp   *BuzzResp
}

// BuzzPacket implements GossipData.
var _ mesh.GossipData = &BuzzPacket{}

// Construct an empty BuzzPacket object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func NewBuzzPacket(events *[]*BuzzEvent, req *BuzzReq, resp *BuzzResp) *BuzzPacket {
	return &BuzzPacket{
		Events: *events,
		Req:    req,
		Resp:   resp,
	}
}

func newBuzzPacketFromBytes(data []byte) (*BuzzPacket, error) {
	var bp BuzzPacket
	decBuf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(decBuf).Decode(&bp); err != nil {
		return nil, err
	}

	return &bp, nil
}

// Encode serializes BuzzPacket to a slice of byte-slices.
func (pkt *BuzzPacket) Encode() [][]byte {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(pkt); err != nil {
		panic(err)
	}

	return [][]byte{buf.Bytes()}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete BuzzPacket.
func (st *BuzzPacket) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	st.Events = append(st.Events, other.(*BuzzPacket).Events...)
	return st
}
