package core

import (
	"github.com/ryogrid/nostrp2p/schema"
)

type DataManager interface {
	StoreEvent(evt *schema.Np2pEvent)
	GetEventById(evtId [32]byte) (*schema.Np2pEvent, bool)
	StoreProfile(evt *schema.Np2pEvent)
	GetProfileLocal(pubkey64bit uint64) *schema.Np2pEvent
	GetLatestEvents(since int64, until int64, limit int64) *[]*schema.Np2pEvent
	StoreFollowList(evt *schema.Np2pEvent)
	GetFollowListLocal(pubkey64bit uint64) *schema.Np2pEvent
	AddReSendNeededEvent(destIds []uint64, evt *schema.Np2pEvent, isLogging bool)
	RemoveReSendNeededEvent(resendEvt *schema.ResendEvent, evt *schema.Np2pEvent)
	GetReSendNeededEventItr() Np2pItr
	RemoveStoreAmountLimitOveredEvents()
}

type Np2pItr interface {
	Next() bool
	Value() interface{}
}
