package core

import (
	"fmt"
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
	"github.com/ryogrid/buzzoon/schema"
	"sync"
)

type DataManager struct {
	SelfPubkey   [32]byte
	EventList    sortedlist.List // timestamp(int64) -> *schema.BuzzEvent
	EventListMtx *sync.Mutex
}

func NewDataManager(pubkey [32]byte) *DataManager {
	return &DataManager{
		SelfPubkey:   pubkey,
		EventList:    sortedlist.NewTree(),
		EventListMtx: &sync.Mutex{},
	}
}

func (dman *DataManager) handleReceived(pkt *schema.BuzzPacket) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if evt.Pubkey != dman.SelfPubkey {
				// store received event data (on memory)
				tmpEvt := *evt
				dman.StoreEvent(&tmpEvt)

				switch evt.Kind {
				case 1: // post
					// display (TEMPORAL IMPL)
					dman.DispPostAtStdout(evt)
				}
			}
		}
	} else {
		fmt.Println("pkt.Events is nil")
	}
	return nil
}

func (dman *DataManager) StoreEvent(evt *schema.BuzzEvent) {
	// TODO: current impl overwrites the same timestamp event (DataManager::StoreEvent)
	dman.EventListMtx.Lock()
	dman.EventList.Add(evt.Created_at, evt)
	dman.EventListMtx.Unlock()
}

// TODO: TEMPORAL IMPL
func (dman *DataManager) DispPostAtStdout(evt *schema.BuzzEvent) {
	fmt.Println(evt.Tags["nickname"][0] + "> " + evt.Content)
}
