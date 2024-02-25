package schema

import (
	"github.com/ryogrid/buzzoon/buzz_const"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weaveworks/mesh"
)

// BuzzPacket is an implementation of GossipData
type BuzzPacket struct {
	SrvVer uint16 // version of buzzoon server implementation
	PktVer uint16 // BuzzPacket data structure version for compatibility
	Events []*BuzzEvent
	Reqs   []*BuzzReq
}

// BuzzPacket implements GossipData.
var _ mesh.GossipData = &BuzzPacket{}

// Construct an empty BuzzPacket object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func NewBuzzPacket(events *[]*BuzzEvent, req *[]*BuzzReq) *BuzzPacket {
	var events_ []*BuzzEvent = nil
	if events != nil {
		events_ = *events
	}
	var req_ []*BuzzReq = nil
	if req != nil {
		req_ = *req
	}

	return &BuzzPacket{
		SrvVer: buzz_const.ServerImplVersion,
		PktVer: buzz_const.PacketStructureVersion,
		Events: events_,
		Reqs:   req_,
	}
}

func NewBuzzPacketFromBytes(data []byte) (*BuzzPacket, error) {
	var bp BuzzPacket
	err := msgpack.Unmarshal(data, &bp)
	if err != nil {
		return nil, err
	}

	return &bp, nil
}

// Encode serializes BuzzPacket to a slice of byte-slices.
func (pkt *BuzzPacket) Encode() [][]byte {
	b, err := msgpack.Marshal(pkt)
	if err != nil {
		panic(err)
	}
	return [][]byte{b}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete BuzzPacket.
func (st *BuzzPacket) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	if st.Events != nil && other.(*BuzzPacket).Events != nil {
		st.Events = append(st.Events, other.(*BuzzPacket).Events...)
	} else if st.Events == nil && other.(*BuzzPacket).Events != nil {
		st.Events = other.(*BuzzPacket).Events
	}
	if st.Reqs != nil && other.(*BuzzPacket).Reqs != nil {
		st.Reqs = append(st.Reqs, other.(*BuzzPacket).Reqs...)
	} else if st.Reqs == nil && other.(*BuzzPacket).Reqs != nil {
		st.Reqs = other.(*BuzzPacket).Reqs
	}
	return st
}
