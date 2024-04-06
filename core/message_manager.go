package core

import (
	"fmt"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"github.com/weaveworks/mesh"
	"math"
	"strconv"
	"time"
)

const (
	KIND_EVT_PROFILE        = 0
	KIND_EVT_POST           = 1
	KIND_EVT_FOLLOW_LIST    = 3
	KIND_EVT_REACTION       = 7
	KIND_REQ_PROFILE        = KIND_EVT_PROFILE
	KIND_REQ_POST           = KIND_EVT_POST
	KIND_REQ_FOLLOW_LIST    = KIND_EVT_FOLLOW_LIST
	KIND_REQ_SHARE_EVT_DATA = 40000
)

type MessageManager struct {
	DataMan *DataManager
	send    mesh.Gossip // set by Np2pPeer.Register
}

// when recovery, src is math.MaxUint64
func (mm *MessageManager) handleRecvMsgBcastEvt(src uint64, pkt *schema.Np2pPacket, evt *schema.Np2pEvent) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: received from " + strconv.Itoa(int(src)))
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: ", pkt)

	// handle with new goroutine
	go func() {
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
			case KIND_EVT_POST: // post
				if val, ok := evt.Tags["u"]; ok {
					// delete "u" tag because it is not necessary for storing
					delete(evt.Tags, "u")
					if len(evt.Tags) == 0 {
						evt.Tags = nil
					}

					// profile update time is attached
					//(periodically attached the tag for avoiding old profile is kept)
					recvdTime := val[0].(int64)
					shortId := np2p_util.GetLower64bitUint(evt.Pubkey)
					if mm.DataMan.GetProfileLocal(shortId) == nil ||
						recvdTime > mm.DataMan.GetProfileLocal(shortId).Created_at {
						// TODO: need to implement limitation of request times (MessageManager::handleRecvMsgBcastEvt)
						// profile is updated. request latest profile asynchronous.
						go mm.UnicastProfileReq(shortId & 0x0000ffffffffffff)
					}
				}
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
	}()

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
				go mm.UnicastOwnProfile(uint64(src))
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
	return mm.send.GossipUnicast(mesh.PeerName(dst), pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt schema.EncodableAndMergeable) {
	mm.send.GossipBroadcast(pkt.(mesh.GossipData))
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
	mm.SendMsgUnicast(pubkey64bit&0x0000ffffffffffff, pkt)
}

func (mm *MessageManager) UnicastFollowListReq(pubkey64bit uint64) {
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_FOLLOW_LIST, nil)}
	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgUnicast(pubkey64bit&0x0000ffffffffffff, pkt)
}

// used for response of profile request
func (mm *MessageManager) UnicastOwnProfile(dest uint64) {
	if glo_val.CurrentProfileEvt != nil {
		// send latest profile data
		events := []*schema.Np2pEvent{glo_val.CurrentProfileEvt}
		mm.SendMsgUnicast(dest&0x0000ffffffffffff, schema.NewNp2pPacket(&events, nil))
	}
}

// TODO: need to implent MessageManager::GenProfileFromEvent
func GenProfileFromEvent(evt *schema.Np2pEvent) *schema.Np2pProfile {
	return &schema.Np2pProfile{
		Pubkey64bit: np2p_util.GetLower64bitUint(evt.Pubkey),
		//Name:        evt.Tags["name"][0].(string),
		//About:       evt.Tags["about"][0].(string),
		//Picture:     evt.Tags["picture"][0].(string),
		Name:      "",
		About:     "",
		Picture:   "",
		UpdatedAt: evt.Created_at,
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
	events := mm.DataMan.GetLatestEvents(time.Now().Unix()-3*24*3600, math.MaxInt64)
	pkt := schema.NewNp2pPacket(events, nil)
	mm.SendMsgUnicast(dest, pkt)
	np2p_util.Np2pDbgPrintln("UnicastHavingEvtData: sent " + strconv.Itoa(len(*events)) + " events")
}

func (mm *MessageManager) UnicastEventData(destPubHexStr string, evt *schema.Np2pEvent) error {
	events := []*schema.Np2pEvent{evt}
	pkt := schema.NewNp2pPacket(&events, nil)
	return mm.SendMsgUnicast(np2p_util.Get6ByteUint64FromHexPubKeyStr(destPubHexStr), pkt)
}
