package core

import (
	"fmt"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"math"
	"strconv"
	"time"
)

type Np2pTransport interface {
	SendMsgBroadcast(schema.EncodableAndMergeable)
	SendMsgUnicast(uint64, []byte) error
}

const (
	KIND_EVT_PROFILE        = 0
	KIND_EVT_POST           = 1
	KIND_EVT_REPOST         = 6
	KIND_EVT_FOLLOW_LIST    = 3
	KIND_EVT_REACTION       = 7
	KIND_REQ_PROFILE        = KIND_EVT_PROFILE
	KIND_REQ_POST           = KIND_EVT_POST
	KIND_REQ_FOLLOW_LIST    = KIND_EVT_FOLLOW_LIST
	KIND_REQ_SHARE_EVT_DATA = 40000
)

type MessageManager struct {
	DataMan     DataManager
	send        Np2pTransport // set with Register later
	evtReSender *EventResender
}

func NewMessageManager(dman DataManager) *MessageManager {
	ret := &MessageManager{DataMan: dman}
	ret.evtReSender = NewEventResender(dman, ret)
	return ret
}

// when recovery, src is math.MaxUint64
func (mm *MessageManager) handleRecvMsgBcastEvt(src uint64, pkt *schema.Np2pPacket, evt *schema.Np2pEvent) error {
	// TODO: need to use on-disk DB (OnMemoryDataManager::mergeReceived)
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: received from " + strconv.Itoa(int(src)))
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: ", pkt)

	// handle with new goroutine
	handleFunc := func() {
		if evt.Pubkey != *glo_val.SelfPubkey || src == math.MaxUint64 {
			switch evt.Kind {
			case KIND_EVT_PROFILE: // profile
				// store received profile data

				mm.DataMan.StoreProfile(evt)
				if evt.Pubkey == *glo_val.SelfPubkey && (glo_val.CurrentProfileEvt == nil || glo_val.CurrentProfileEvt.Created_at < evt.Created_at) {
					// this route works only when recovery
					//glo_val.ProfileMyOwn = prof
					glo_val.CurrentProfileEvt = evt
				}
			case KIND_EVT_POST:
				// do nothing
			case KIND_EVT_REPOST:
				// do nothing
			case KIND_EVT_FOLLOW_LIST:
				mm.DataMan.StoreFollowList(evt)
				if evt.Pubkey == *glo_val.SelfPubkey && (glo_val.CurrentFollowListEvt == nil || glo_val.CurrentFollowListEvt.Created_at < evt.Created_at) {
					// this route works only when recovery
					//glo_val.ProfileMyOwn = prof
					glo_val.CurrentProfileEvt = evt
				}
			case KIND_EVT_REACTION:
				// do nothing
			default:
				fmt.Println("received unknown kind event: " + strconv.Itoa(int(evt.Kind)))
			}

			// store received event data (on memory)
			tmpEvt := *evt
			mm.DataMan.StoreEvent(&tmpEvt)
		}
	}

	if src == math.MaxUint64 {
		// when recovery, execute handleFunc() directly (sequentially)
		handleFunc()
	} else {
		// when normal, execute handleFunc() asynchronously
		go handleFunc()
	}

	return nil
}

// when recovery, src is math.MaxUint64
func (mm *MessageManager) handleRecvMsgBcastReq(src uint64, pkt *schema.Np2pPacket, req *schema.Np2pReq) error {
	go func() {
		if src != glo_val.SelfPubkey64bit {
			switch req.Kind {
			case KIND_REQ_SHARE_EVT_DATA: // need to having event datas
				go mm.UnicastHavingEvtData(src)
			case KIND_EVT_REACTION:
				// do nothing
			default:
				fmt.Println("received unknown kind req: " + strconv.Itoa(int(req.Kind)))
			}
			np2p_util.Np2pDbgPrintln("received request: " + strconv.Itoa(int(req.Kind)))
		}

	}()

	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src uint64, pkt *schema.Np2pPacket) error {
	if pkt.Events != nil && len(pkt.Events) > 0 {
		// handle with new goroutine
		go func() {
			for _, evt := range pkt.Events {
				// handle response of request

				if evt.Verify() == false {
					// invalid signiture
					fmt.Println("invalid signiture")
					continue
				}

				// store received event data (on memory)
				mm.DataMan.StoreEvent(evt)

				switch evt.Kind {
				case KIND_EVT_PROFILE: // response of KIND_REQ_PROFILE or KIND_REQ_SHARE_EVT_DATA
					// store received profile data
					mm.DataMan.StoreProfile(evt)
				case KIND_EVT_POST: // response of KIND_REQ_SHARE_EVT_DATA
					// do nothing
				case KIND_EVT_FOLLOW_LIST: // response of KIND_REQ_FOLLOW_LIST
					mm.DataMan.StoreFollowList(evt) // store received follow list data
				case KIND_EVT_REACTION:
					// do nothing
				default:
					fmt.Println("received unknown kind event: " + strconv.Itoa(int(pkt.Events[0].Kind)))
				}
			}
		}()
		return nil
	}

	if pkt.Reqs != nil {
		// handle with new goroutine
		go func() {
			// handle request
			switch pkt.Reqs[0].Kind {
			case KIND_REQ_PROFILE: // profile request
				// send profile data asynchronous
				go mm.UnicastOwnProfile(src)
			case KIND_REQ_FOLLOW_LIST: // follow list request
				// send follow list data asynchronous
				go mm.UnicastOwnFollowList(src)
			case KIND_REQ_POST:
				// send post data asynchronous
				if tgtEvtId, ok := pkt.Reqs[0].Args["evtId"][0].([32]byte); ok {
					if tgtEvt, ok2 := mm.DataMan.GetEventById(tgtEvtId); ok2 {
						events := []*schema.Np2pEvent{tgtEvt}
						go mm.SendMsgUnicast(src, schema.NewNp2pPacket(&events, nil))
					}
				}
			default:
				fmt.Println("received unknown kind request: " + strconv.Itoa(int(pkt.Reqs[0].Kind)))
			}
		}()
		return nil
	}

	fmt.Println("MessageManager::handleRecvMsgUnicast: received package which does not include both Events and Reqs")
	return nil
}

func (mm *MessageManager) SendMsgUnicast(dst uint64, pkt *schema.Np2pPacket) error {
	//return mm.send.GossipUnicast(mesh.PeerName(dst), pkt.Encode()[0])
	return mm.send.SendMsgUnicast(dst, pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt schema.EncodableAndMergeable) {
	//mm.send.GossipBroadcast(pkt.(mesh.GossipData))
	mm.send.SendMsgBroadcast(pkt)
}

func (mm *MessageManager) BcastOwnPost(evt *schema.Np2pEvent) {
	events := []*schema.Np2pEvent{evt}
	mm.SendMsgBroadcast(schema.NewNp2pPacket(&events, nil))
	// store own issued event
	mm.DataMan.StoreEvent(evt)
}

// func (mm *MessageManager) BcastProfile(name *string, about *string, picture *string) *schema.Np2pProfile {
func (mm *MessageManager) BcastProfile(evt *schema.Np2pEvent) {
	//event := mm.constructProfileEvt(name, about, picture)
	events := []*schema.Np2pEvent{evt}
	mm.SendMsgBroadcast(schema.NewNp2pPacket(&events, nil))
	mm.DataMan.StoreEvent(evt)

	//storeProf := GenProfileFromEvent(evt)
	mm.DataMan.StoreProfile(evt)
}

func (mm *MessageManager) UnicastProfileReq(pubkey64bit uint64) {
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_PROFILE, nil)}
	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgUnicast(pubkey64bit, pkt)
}

func (mm *MessageManager) UnicastPostReq(pubkey64bit uint64, evtId [32]byte) {
	arg := make(map[string][]interface{})
	arg["evtId"] = []interface{}{evtId}
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_POST, arg)}
	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgUnicast(pubkey64bit, pkt)
}

func (mm *MessageManager) UnicastFollowListReq(pubkey64bit uint64) {
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_FOLLOW_LIST, nil)}
	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgUnicast(pubkey64bit, pkt)
}

// used for response of follow list request
func (mm *MessageManager) UnicastOwnFollowList(dest uint64) {
	if flistEvt := mm.DataMan.GetFollowListLocal(glo_val.SelfPubkey64bit); flistEvt != nil {
		// send latest follow list data
		events := []*schema.Np2pEvent{flistEvt}
		mm.SendMsgUnicast(dest, schema.NewNp2pPacket(&events, nil))
	}
	// when no follow list data, do nothing
}

// used for response of profile request
func (mm *MessageManager) UnicastOwnProfile(dest uint64) {
	if glo_val.CurrentProfileEvt != nil {
		// send latest profile data
		events := []*schema.Np2pEvent{glo_val.CurrentProfileEvt}
		mm.SendMsgUnicast(dest, schema.NewNp2pPacket(&events, nil))
	}
}

// TODO: TEMPORAL IMPL
func (mm *MessageManager) BcastShareEvtDataReq() {
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_SHARE_EVT_DATA, nil)}

	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgBroadcast(pkt)
}

// TODO: TEMPORAL IMPL
// send latest 3days events
func (mm *MessageManager) UnicastHavingEvtData(dest uint64) {
	events := mm.DataMan.GetLatestEvents(time.Now().Unix()-3*24*3600, math.MaxInt64, -1)
	pkt := schema.NewNp2pPacket(events, nil)
	mm.SendMsgUnicast(dest, pkt)
	np2p_util.Np2pDbgPrintln("UnicastHavingEvtData: sent " + strconv.Itoa(len(*events)) + " events")
}

func (mm *MessageManager) UnicastEventData(destPubHexStr string, evt *schema.Np2pEvent) error {
	var events []*schema.Np2pEvent
	if evt != nil {
		events = []*schema.Np2pEvent{evt}
	} else {
		events = []*schema.Np2pEvent{}
	}

	pkt := schema.NewNp2pPacket(&events, nil)
	return mm.SendMsgUnicast(np2p_util.Get6ByteUint64FromHexPubKeyStr(destPubHexStr), pkt)
}

func (mm *MessageManager) SetTransport(tport Np2pTransport) {
	mm.send = tport
}
