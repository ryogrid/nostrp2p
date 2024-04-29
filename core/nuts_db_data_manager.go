package core

import (
	"fmt"
	"github.com/nutsdb/nutsdb"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
	"strconv"
)

type NutsDBDataManager struct {
	dbFilePath string
	db         *nutsdb.DB
}

var _ DataManager = &NutsDBDataManager{}

func NewNutsDBDataManager() DataManager {
	dbFilePath := "./" + strconv.FormatUint(glo_val.SelfPubkey64bit, 16)
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
		return tx.NewSortSetBucket("EvtListTimeKey")
	}); err2 != nil {
		fmt.Println(err2)
	}

	// serialized event ID [32]byte -> serialized timestamp(int64)
	if err2 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewKVBucket("EvtIdxMapIdKey")
	}); err2 != nil {
		fmt.Println(err2)
	}

	// serialized pubkey lower 64bit (uint64) -> serialized timestamp(int64)
	if err3 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewKVBucket("ProfEvtIdxMap")
	}); err3 != nil {
		fmt.Println(err3)
	}

	// serialized pubkey lower 64bit (uint64) -> serialized timestamp(int64)
	if err4 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewKVBucket("FollowListEvtIdxMap")
	}); err4 != nil {
		fmt.Println(err4)
	}

	// serialized pubkey lower 64bit (uint64) -> timestamp(int64)
	if err5 := db.Update(func(tx *nutsdb.Tx) error {
		return tx.NewKVBucket("FollowListEvtIdxMap")
	}); err5 != nil {
		fmt.Println(err5)
	}

	return &NutsDBDataManager{
		dbFilePath: dbFilePath,
		db:         db,
	}
}

func (n NutsDBDataManager) StoreEvent(evt *schema.Np2pEvent) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) GetEventById(evtId [32]byte) (*schema.Np2pEvent, bool) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) StoreProfile(evt *schema.Np2pEvent) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) GetProfileLocal(pubkey64bit uint64) *schema.Np2pEvent {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) GetLatestEvents(since int64, until int64) *[]*schema.Np2pEvent {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) StoreFollowList(evt *schema.Np2pEvent) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) GetFollowListLocal(pubkey64bit uint64) *schema.Np2pEvent {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) AddReSendNeededEvent(destIds []uint64, evt *schema.Np2pEvent, isLogging bool) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) RemoveReSendNeededEvent(evt *schema.Np2pEvent) {
	//TODO implement me
	panic("implement me")
}

func (n NutsDBDataManager) GetReSendNeededEventItr() Np2pItr {
	//TODO implement me
	panic("implement me")
}
