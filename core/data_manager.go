package core

import (
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/schema"
	"strconv"
	"sync"
	"time"
)

const profReqInterval = 30 * time.Second

type DataManager struct {
	EvtListTimeKey    sortedlist.List // timestamp(int64) -> *schema.Np2pEvent
	EvtListTimeKeyMtx *sync.Mutex
	EvtMapIdKey       sync.Map // [32]byte -> *schema.Np2pEvent
	// latest profile only stored
	ProfMap   sync.Map // pubkey lower 64bit (uint64) -> *schema.Np2pProfile
	EvtLogger *EventDataLogger
}

func NewDataManager() *DataManager {
	return &DataManager{
		EvtListTimeKey:    sortedlist.NewTree(),
		EvtListTimeKeyMtx: &sync.Mutex{},
		EvtMapIdKey:       sync.Map{},
		ProfMap:           sync.Map{},
		EvtLogger:         NewEventDataLogger("./" + strconv.FormatUint(glo_val.SelfPubkey64bit, 16) + ".evtlog"),
	}
}

func (dman *DataManager) StoreEvent(evt *schema.Np2pEvent) {
	// TODO: current impl overwrites the same timestamp event on EvtListTimeKey (DataManager::StoreEvent)
	dman.EvtListTimeKeyMtx.Lock()
	//evt.Sig = nil // set nil because already verified
	dman.EvtListTimeKey.Add(evt.Created_at, evt)
	dman.EvtListTimeKeyMtx.Unlock()
	if _, ok := dman.EvtMapIdKey.Load(evt.Id); ok {
		dman.EvtMapIdKey.Store(evt.Id, evt)
	} else {
		dman.EvtMapIdKey.Store(evt.Id, evt)
		// log event data when it is not duplicated
		// write asynchrounously
		if evt.Tags != nil {
			// consider when func is called from recovery
			if _, ok2 := evt.Tags["recovering"]; ok2 {
				delete(evt.Tags, "recovering")
				if len(evt.Tags) == 0 {
					evt.Tags = nil
				}
				return
			}
		}
		go dman.EvtLogger.WriteLog(evt.Encode())
	}
}

func (dman *DataManager) StoreProfile(prof *schema.Np2pProfile) {
	dman.ProfMap.Store(prof.Pubkey64bit, prof)
}

func (dman *DataManager) GetProfileLocal(pubkey64bit uint64) *schema.Np2pProfile {
	if val, ok := dman.ProfMap.Load(pubkey64bit); ok {
		return val.(*schema.Np2pProfile)
	}
	return nil
}

func (dman *DataManager) GetLatestEvents(since int64, until int64) *[]*schema.Np2pEvent {
	dman.EvtListTimeKeyMtx.Lock()
	defer dman.EvtListTimeKeyMtx.Unlock()
	itr := dman.EvtListTimeKey.Range(since, until)

	ret := make([]*schema.Np2pEvent, 0)
	for itr.Next() {
		val := itr.Value()
		ret = append(ret, val.(*schema.Np2pEvent))
	}
	return &ret
}
