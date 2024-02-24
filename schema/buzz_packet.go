package schema

import (
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weaveworks/mesh"
)

const PacketStructureVersion uint16 = 2

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

func NewBuzzPacketFromBytes(data []byte) (*BuzzPacket, error) {
	var bp BuzzPacket
	//copiedData := make([]byte, len(data))
	//copy(copiedData, data)
	//copiedData = buzz_util.GzipDecompless(copiedData)
	//decBuf := bytes.NewBuffer(copiedData)
	//decBuf := bytes.NewBuffer(data)
	//if err := gob.NewDecoder(decBuf).Decode(&bp); err != nil {
	//	return nil, err
	//}
	err := msgpack.Unmarshal(data, &bp)
	if err != nil {
		return nil, err
	}

	return &bp, nil
}

// Encode serializes BuzzPacket to a slice of byte-slices.
func (pkt *BuzzPacket) Encode() [][]byte {
	//buf := bytes.NewBuffer(nil)
	//if err := gob.NewEncoder(buf).Encode(pkt); err != nil {
	//	panic(err)
	//}
	//
	////return [][]byte{buzz_util.GzipCompless(buf.Bytes())}
	//return [][]byte{buf.Bytes()}
	b, err := msgpack.Marshal(pkt)
	if err != nil {
		panic(err)
	}
	return [][]byte{b}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete BuzzPacket.
func (st *BuzzPacket) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	st.Events = append(st.Events, other.(*BuzzPacket).Events...)
	return st
}

//func (st *BuzzPacket) GobEncode() ([]byte, error) {
//	buf := make([]byte, 0)
//	buf = binary.LittleEndian.AppendUint16(buf, st.SrvVer)
//	buf = binary.LittleEndian.AppendUint16(buf, st.PktVer)
//
//	if st.Events != nil {
//		eventsBuf := bytes.NewBuffer(nil)
//		_ = gob.NewEncoder(eventsBuf).Encode(st.Events)
//		eventsBytes := eventsBuf.Bytes()
//		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(eventsBytes)))
//		buf = append(buf, eventsBytes...)
//	} else {
//		buf = binary.LittleEndian.AppendUint16(buf, 0)
//	}
//
//	if st.Req != nil {
//		reqBuf := bytes.NewBuffer(nil)
//		_ = gob.NewEncoder(reqBuf).Encode(st.Req)
//		reqBytes := reqBuf.Bytes()
//		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(reqBytes)))
//		buf = append(buf, reqBytes...)
//	} else {
//		buf = binary.LittleEndian.AppendUint16(buf, 0)
//	}
//
//	return buzz_util.GzipCompless(buf), nil
//}
//
//func (st *BuzzPacket) GobDecode(data []byte) error {
//	decomped := buzz_util.GzipDecompless(data)
//	st.SrvVer = binary.LittleEndian.Uint16(decomped[:2])
//	st.PktVer = binary.LittleEndian.Uint16(decomped[2:4])
//	eventsLen := binary.LittleEndian.Uint16(decomped[4:6])
//	var offset = 6
//	if eventsLen != 0 {
//		offset = 6 + int(eventsLen)
//		decoder := gob.NewDecoder(bytes.NewBuffer(decomped[6:offset]))
//		err := decoder.Decode(&st.Events)
//		if err != nil {
//			panic(err)
//		}
//	} else {
//		st.Events = nil
//	}
//	reqLen := binary.LittleEndian.Uint16(decomped[offset : offset+2])
//	if reqLen != 0 {
//		offset = offset + 2
//		offset2 := offset + int(reqLen)
//		gob.NewDecoder(bytes.NewBuffer(decomped[offset:offset2])).Decode(st.Req)
//	} else {
//		st.Req = nil
//	}
//
//	return nil
//}
