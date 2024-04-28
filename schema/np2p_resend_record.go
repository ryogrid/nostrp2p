package schema

import (
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/vmihailenco/msgpack/v5"
)

// represents a record of the event to be resent
type ResendRecord struct {
	DestIds   []uint64
	EvtId     [np2p_const.EventIdSize]byte
	CreatedAt int64
}

func NewResendRecord(destIds []uint64, evtId [np2p_const.EventIdSize]byte, createdAt int64) *ResendRecord {
	return &ResendRecord{DestIds: destIds, EvtId: evtId, CreatedAt: createdAt}
}

func (rr *ResendRecord) Encode() []byte {
	b, err := msgpack.Marshal(rr)
	if err != nil {
		panic(err)
	}
	return b
}

func NewResendEventFromBytes(b []byte) (*ResendRecord, error) {
	var re ResendRecord
	if err := msgpack.Unmarshal(b, &re); err != nil {
		return nil, err
	}

	return &re, nil
}
