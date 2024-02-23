package core

import (
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
	"github.com/ryogrid/buzzoon/schema"
	"sync"
	"time"
)

const profReqInterval = 30 * time.Second

type DataManager struct {
	EvtListTimeKey    sortedlist.List // timestamp(int64) -> *schema.BuzzEvent
	EvtListTimeKeyMtx *sync.Mutex
	EvtMapIdKey       sync.Map // event id(uint64) -> *schema.BuzzEvent
	// latest profile only stored
	ProfMap sync.Map // pubkey lower 64bit (uint64) -> *schema.BuzzProfile
}

func NewDataManager() *DataManager {
	return &DataManager{
		EvtListTimeKey:    sortedlist.NewTree(),
		EvtListTimeKeyMtx: &sync.Mutex{},
		EvtMapIdKey:       sync.Map{},
		ProfMap:           sync.Map{},
	}
}

func (dman *DataManager) StoreEvent(evt *schema.BuzzEvent) {
	// TODO: current impl overwrites the same timestamp event on EvtListTimeKey (DataManager::StoreEvent)
	dman.EvtListTimeKeyMtx.Lock()
	evt.Sig = nil // set nil because already verified
	dman.EvtListTimeKey.Add(evt.Created_at, evt)
	dman.EvtListTimeKeyMtx.Unlock()
	dman.EvtMapIdKey.Store(evt.Id, evt)
}

func (dman *DataManager) StoreProfile(prof *schema.BuzzProfile) {
	dman.ProfMap.Store(prof.Pubkey64bit, prof)
}

func (dman *DataManager) GetProfileLocal(pubkey64bit uint64) *schema.BuzzProfile {
	if val, ok := dman.ProfMap.Load(pubkey64bit); ok {
		return val.(*schema.BuzzProfile)
	}
	return nil
}
