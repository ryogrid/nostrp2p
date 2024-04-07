package schema

import (
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/vmihailenco/msgpack/v5"
)

type ResendEvent struct {
	DestIds   []uint64
	EvtId     [np2p_const.EventIdSize]byte
	CreatedAt int64
}

func NewResendEvent(destIds []uint64, evtId [np2p_const.EventIdSize]byte, createdAt int64) *ResendEvent {
	return &ResendEvent{DestIds: destIds, EvtId: evtId, CreatedAt: createdAt}
}

func (re *ResendEvent) Encode() []byte {
	b, err := msgpack.Marshal(re)
	if err != nil {
		panic(err)
	}
	return b
}

func NewResendEventFromBytes(b []byte) (*ResendEvent, error) {
	var re ResendEvent
	if err := msgpack.Unmarshal(b, &re); err != nil {
		return nil, err
	}

	return &re, nil
}
