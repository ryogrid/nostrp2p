package core

import (
	"encoding/binary"
	"fmt"
	"github.com/nutsdb/nutsdb"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
)

type NutsDBItr struct {
	curIdx int
	elems  []interface{}
}

func NewNutsDBItr(elems []interface{}) *NutsDBItr {
	return &NutsDBItr{
		curIdx: -1,
		elems:  elems,
	}
}

func (n *NutsDBItr) Next() bool {
	if n.curIdx < len(n.elems) {
		n.curIdx++
		return true
	}
	return false
}

var _ Np2pItr = &NutsDBItr{}

func (n *NutsDBItr) Value() interface{} {
	return n.elems[n.curIdx]
}

type NutsDBDataManager struct {
	dbFilePath string
	db         *nutsdb.DB
}

var _ DataManager = &NutsDBDataManager{}

const EventListTimeKey = "EvtListTimeKey"
const EventIdxMapIdKey = "EvtIdxMapIdKey"
const ProfEvtIdxMap = "ProfEvtIdxMap"
const FollowListEvtIdxMap = "FollowListEvtIdxMap"
const ReSendNeededEvtList = "ReSendNeededEvtList"

func NewNutsDBDataManager() DataManager {
	dbFilePath := "./" + strconv.FormatUint(glo_val.SelfPubkey64bit, 16)
	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		err2 := os.Mkdir(dbFilePath, os.ModePerm)
		if err2 != nil {
			panic(err2)
		}
	}
	opt := nutsdb.DefaultOptions
	opt.EntryIdxMode = nutsdb.HintKeyValAndRAMIdxMode
	opt.HintKeyAndRAMIdxCacheSize = 300 * 1024 * 1024 // 300MB
	db, err := nutsdb.Open(
		opt,
		nutsdb.WithDir(dbFilePath),
	)
	if err != nil {
		log.Fatal(err)
	}

	// key: "timestamp"
	// score: timestamp(float64) -> value: serialized schema.Np2pEvent
	if err2 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureSortedSet, EventListTimeKey)
	}); err2 != nil {
		fmt.Println(err2)
	}

	// serialized event ID [32]byte -> serialized timestamp(int64)
	if err2 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, EventIdxMapIdKey)
	}); err2 != nil {
		fmt.Println(err2)
	}

	// serialized pubkey lower 64bit (uint64) -> serialized timestamp(int64)
	if err3 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, ProfEvtIdxMap)
	}); err3 != nil {
		fmt.Println(err3)
	}

	// serialized pubkey lower 64bit (uint64) -> serialized timestamp(int64)
	if err4 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, FollowListEvtIdxMap)
	}); err4 != nil {
		fmt.Println(err4)
	}

	// serialized pubkey lower 64bit (uint64) -> timestamp(int64)
	if err5 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureSortedSet, ReSendNeededEvtList)
	}); err5 != nil {
		fmt.Println(err5)
	}

	return &NutsDBDataManager{
		dbFilePath: dbFilePath,
		db:         db,
	}
}

func (n *NutsDBDataManager) StoreEvent(evt *schema.Np2pEvent) {
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		return tx.ZAdd(EventListTimeKey, []byte("time"), float64(evt.Created_at), evt.Encode())
	}); err != nil {
		fmt.Println(err)
	}
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(EventIdxMapIdKey, evt.Id[:], np2p_util.ConvInt64ToBytes(evt.Created_at), nutsdb.Persistent)
	}); err != nil {
		fmt.Println(err)
	}
}

func (n *NutsDBDataManager) getEventByTimestampBytes(tsBytes []byte) *schema.Np2pEvent {
	var ret *schema.Np2pEvent
	ts := float64(binary.LittleEndian.Uint64(tsBytes))
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if entries, err2 := tx.ZRangeByScore(EventListTimeKey, []byte("time"), ts, ts, nil); err2 != nil {
			return err2
		} else {
			if len(entries) == 0 {
				return nil
			}
			ret, _ = schema.NewNp2pEventFromBytes(entries[0].Value)
			return nil
		}
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	return ret
}

func (n *NutsDBDataManager) GetEventById(evtId [32]byte) (*schema.Np2pEvent, bool) {
	var ret *schema.Np2pEvent
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if val, err2 := tx.Get(EventIdxMapIdKey, evtId[:]); err2 != nil {
			return err2
		} else {
			ret = n.getEventByTimestampBytes(val)
			return nil
		}
	}); err != nil {
		fmt.Println(err)
		return nil, false
	}
	return ret, true
}

func (n *NutsDBDataManager) StoreProfile(evt *schema.Np2pEvent) {
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		tmpPubKey := evt.Pubkey
		key := tmpPubKey[len(tmpPubKey)-8:]
		// little endian
		slices.Reverse(key)
		return tx.Put(ProfEvtIdxMap, key, np2p_util.ConvInt64ToBytes(evt.Created_at), nutsdb.Persistent)
	}); err != nil {
		fmt.Println(err)
	}
}

func (n *NutsDBDataManager) GetProfileLocal(pubkey64bit uint64) *schema.Np2pEvent {
	var ret *schema.Np2pEvent
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if val, err2 := tx.Get(ProfEvtIdxMap, np2p_util.ConvUint64ToBytes(pubkey64bit)); err2 != nil {
			return err2
		} else {
			ret = n.getEventByTimestampBytes(val)
			return nil
		}
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	return ret
}

func (n *NutsDBDataManager) GetLatestEvents(since int64, until int64) *[]*schema.Np2pEvent {
	var ret []*schema.Np2pEvent
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if entries, err2 := tx.ZRangeByScore(EventListTimeKey, []byte("time"), float64(since), float64(until), nil); err2 != nil {
			return err2
		} else {
			if entries != nil {
				ret = make([]*schema.Np2pEvent, len(entries))
				for idx, entry := range entries {
					ret[idx], _ = schema.NewNp2pEventFromBytes(entry.Value)
				}
				return nil
			} else {
				ret = make([]*schema.Np2pEvent, 0)
				return nil
			}
		}
	}); err != nil {
		fmt.Println(err)
		ret = make([]*schema.Np2pEvent, 0)
		return &ret
	}
	return &ret
}

func (n *NutsDBDataManager) StoreFollowList(evt *schema.Np2pEvent) {
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		tmpPubKey := evt.Pubkey
		key := tmpPubKey[len(tmpPubKey)-8:]
		// little endian
		slices.Reverse(key)
		return tx.Put(FollowListEvtIdxMap, key, np2p_util.ConvInt64ToBytes(evt.Created_at), nutsdb.Persistent)
	}); err != nil {
		fmt.Println(err)
	}
}

func (n *NutsDBDataManager) GetFollowListLocal(pubkey64bit uint64) *schema.Np2pEvent {
	var ret *schema.Np2pEvent
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if val, err2 := tx.Get(FollowListEvtIdxMap, np2p_util.ConvUint64ToBytes(pubkey64bit)); err2 != nil {
			return err2
		} else {
			ret = n.getEventByTimestampBytes(val)
			return nil
		}
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	return ret
}

func (n *NutsDBDataManager) AddReSendNeededEvent(destIds []uint64, evt *schema.Np2pEvent, _isLogging bool) {
	resendEvent := schema.NewResendEvent(destIds, evt.Id, evt.Created_at)
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		return tx.ZAdd(ReSendNeededEvtList, []byte("time"), float64(evt.Created_at), resendEvent.Encode())
	}); err != nil {
		fmt.Println(err)
	}
}

func (n *NutsDBDataManager) RemoveReSendNeededEvent(resendEvt *schema.ResendEvent, evt *schema.Np2pEvent) {
	if err := n.db.Update(func(tx *nutsdb.Tx) error {
		return tx.ZRem(ReSendNeededEvtList, []byte("time"), resendEvt.Encode())
	}); err != nil {
		fmt.Println(err)
	}
}

func (n *NutsDBDataManager) GetReSendNeededEventItr() Np2pItr {
	var ret []interface{} // *schema.ResendEvent
	var entries_ []*nutsdb.SortedSetMember
	if err := n.db.View(func(tx *nutsdb.Tx) error {
		if entries, err2 := tx.ZRangeByScore(ReSendNeededEvtList, []byte("time"), 0, math.MaxFloat64, nil); err2 != nil {
			return err2
		} else {
			entries_ = entries
			return nil
		}
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	for _, entry := range entries_ {
		decoded, _ := schema.NewResendEventFromBytes(entry.Value)
		ret = append(ret, decoded)
	}
	return NewNutsDBItr(ret)
}
