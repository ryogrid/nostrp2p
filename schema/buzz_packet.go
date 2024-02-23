package schema

import (
	"bytes"
	"encoding/gob"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/weaveworks/mesh"
)

const PacketStructureVersion uint16 = 1

// BuzzPacket is an implementation of GossipData
type BuzzPacket struct {
	SrvVer uint16 // version of buzzoon server implementation
	PktVer uint16 // BuzzPacket data structure version for compatibility
	Events []*BuzzEvent
	Req    *BuzzReq
}

// BuzzPacket implements GossipData.
var _ mesh.GossipData = &BuzzPacket{}

// Construct an empty BuzzPacket object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func NewBuzzPacket(events *[]*BuzzEvent, req *BuzzReq) *BuzzPacket {
	return &BuzzPacket{
		SrvVer: buzz_util.ServerImplVersion,
		PktVer: PacketStructureVersion,
		Events: *events,
		Req:    req,
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
