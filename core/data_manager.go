package core

import (
	"encoding/binary"
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"strconv"
	"sync"
)

type DataManager struct {
	EvtListTimeKey    sortedlist.List // timestamp(int64) -> *schema.Np2pEvent
	EvtListTimeKeyMtx *sync.Mutex
	EvtMapIdKey       sync.Map // [32]byte -> *schema.Np2pEvent
	// latest profile only stored
	ProfEvtMap             sync.Map // pubkey lower 64bit (uint64) -> *schema.Np2pEvent
	FollowListEvtMap       sync.Map // pubkey lower 64bit (uint64) -> *schema.Np2pEvent
	EvtLogger              *EventDataLogger
	reSendNeededEvtList    sortedlist.List // timestamp(int64) -> *schema.ResendEvent
	reSendNeededEvtListMtx *sync.Mutex
	//// Attention: in log file of reSendNeededEvtList, re-send finished events are not removed currently
	//reSendFinishedEvtList    sortedlist.List // timestamp(int64) -> created_at(int64)
	//reSendFinishedEvtListMtx *sync.Mutex
}

func NewDataManager() *DataManager {
	return &DataManager{
		EvtListTimeKey:         sortedlist.NewTree(),
		EvtListTimeKeyMtx:      &sync.Mutex{},
		EvtMapIdKey:            sync.Map{},
		ProfEvtMap:             sync.Map{}, // pubkey lower 64bit (uint64) -> *schema.Np2pEvent
		FollowListEvtMap:       sync.Map{}, // pubkey lower 64bit (uint64) -> *schema.Np2pEvent
		EvtLogger:              NewEventDataLogger("./" + strconv.FormatUint(glo_val.SelfPubkey64bit, 16)),
		reSendNeededEvtList:    sortedlist.NewTree(),
		reSendNeededEvtListMtx: &sync.Mutex{},
		//reSendFinishedEvtList:    sortedlist.NewTree(),
		//reSendFinishedEvtListMtx: &sync.Mutex{},
	}
}

func (dman *DataManager) StoreEvent(evt *schema.Np2pEvent) {
	// TODO: current impl overwrites the same timestamp event on EvtListTimeKey (DataManager::StoreEvent)
	dman.EvtListTimeKeyMtx.Lock()
	//evt.Sig = nil // set nil because already verified
	dman.EvtListTimeKey.Add(evt.Created_at, evt)
	dman.EvtListTimeKeyMtx.Unlock()
	if _, ok := dman.EvtMapIdKey.Load(evt.Id); ok {
		// do nothing when it is duplicated
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
		go dman.EvtLogger.WriteLog(dman.EvtLogger.eventLogFile, evt.Encode())
	}
}

func (dman *DataManager) GetEventById(evtId [32]byte) (*schema.Np2pEvent, bool) {
	if val, ok := dman.EvtMapIdKey.Load(evtId); ok {
		return val.(*schema.Np2pEvent), true
	}
	return nil, false
}

func (dman *DataManager) StoreProfile(evt *schema.Np2pEvent) {
	dman.ProfEvtMap.Store(np2p_util.GetLower64bitUint(evt.Pubkey), evt)
}

func (dman *DataManager) GetProfileLocal(pubkey64bit uint64) *schema.Np2pEvent {
	if val, ok := dman.ProfEvtMap.Load(pubkey64bit); ok {
		return val.(*schema.Np2pEvent)
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

func (dman *DataManager) StoreFollowList(evt *schema.Np2pEvent) {
	dman.FollowListEvtMap.Store(np2p_util.GetLower64bitUint(evt.Pubkey), evt)
}

func (dman *DataManager) GetFollowListLocal(pubkey64bit uint64) *schema.Np2pEvent {
	if val, ok := dman.FollowListEvtMap.Load(pubkey64bit); ok {
		return val.(*schema.Np2pEvent)
	}
	return nil
}

func (dman *DataManager) AddReSendNeededEvent(destIds []uint64, evt *schema.Np2pEvent, isLogging bool) {
	dman.reSendNeededEvtListMtx.Lock()
	resendEvent := schema.NewResendEvent(destIds, evt.Id, evt.Created_at)
	dman.reSendNeededEvtList.Add(evt.Created_at, resendEvent)
	dman.reSendNeededEvtListMtx.Unlock()
	if isLogging {
		buf := resendEvent.Encode()
		dman.EvtLogger.WriteLog(dman.EvtLogger.reSendNeededEvtLogFile, buf)
	}
}

// NOTE: add removed event info to reSendFinishedEvtList additionally
func (dman *DataManager) RemoveReSendNeededEvent(evt *schema.Np2pEvent) {
	dman.reSendNeededEvtListMtx.Lock()
	dman.reSendNeededEvtList.Remove(evt.Created_at)
	dman.reSendNeededEvtListMtx.Unlock()
	//dman.reSendFinishedEvtListMtx.Lock()
	//dman.reSendFinishedEvtList.Add(evt.Created_at, evt.Created_at)
	//dman.reSendFinishedEvtListMtx.Unlock()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(evt.Created_at))
	dman.EvtLogger.WriteLog(dman.EvtLogger.reSendFinishedEvtLogFile, buf)
}

func (dman *DataManager) GetReSendNeededEventItr() sortedlist.Iter {
	dman.reSendNeededEvtListMtx.Lock()
	defer dman.reSendNeededEvtListMtx.Unlock()
	return dman.reSendNeededEvtList.All()
}

//func (dman *DataManager) GetReSendFinishedEventItr() sortedlist.Iter {
//	dman.reSendFinishedEvtListMtx.Lock()
//	defer dman.reSendFinishedEvtListMtx.Unlock()
//	return dman.reSendNeededEvtList.All()
//}
