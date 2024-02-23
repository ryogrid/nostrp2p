package core

import (
	"fmt"
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
	"github.com/ryogrid/buzzoon/glo_val"
	"github.com/ryogrid/buzzoon/schema"
	"sync"
)

type DataManager struct {
	EvtListTimeKey    sortedlist.List // timestamp(int64) -> *schema.BuzzEvent
	EvtListTimeKeyMtx *sync.Mutex
	EvtMapIdKey       sync.Map // event id(uint64) -> *schema.BuzzEvent
}

func NewDataManager() *DataManager {
	return &DataManager{
		EvtListTimeKey:    sortedlist.NewTree(),
		EvtListTimeKeyMtx: &sync.Mutex{},
		EvtMapIdKey:       sync.Map{},
	}
}

func (dman *DataManager) handleReceived(pkt *schema.BuzzPacket) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if evt.Pubkey != *glo_val.SelfPubkey {
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
	// TODO: current impl overwrites the same timestamp event on EvtListTimeKey (DataManager::StoreEvent)
	dman.EvtListTimeKeyMtx.Lock()
	evt.Sig = nil // set nil because already verified
	dman.EvtListTimeKey.Add(evt.Created_at, evt)
	dman.EvtListTimeKeyMtx.Unlock()
	dman.EvtMapIdKey.Store(evt.Id, evt)
}

// TODO: TEMPORAL IMPL
func (dman *DataManager) DispPostAtStdout(evt *schema.BuzzEvent) {
	fmt.Println(evt.Tags["nickname"][0] + "> " + evt.Content)
}
