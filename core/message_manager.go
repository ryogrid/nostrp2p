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
	KIND_EVT_REACTION       = 7
	KIND_REQ_PROFILE        = KIND_EVT_PROFILE
	KIND_REQ_POST           = KIND_EVT_POST
	KIND_REQ_SHARE_EVT_DATA = 40000
)

type MessageManager struct {
	DataMan *DataManager
	send    mesh.Gossip // set by Np2pPeer.Register
}

// when recovery, src is math.MaxUint64
func (mm *MessageManager) handleRecvMsgBcastEvt(src mesh.PeerName, pkt *schema.Np2pPacket, evt *schema.Np2pEvent) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: received from " + strconv.Itoa(int(src)))
	np2p_util.Np2pDbgPrintln("handleRecvMsgBcastEvt: ", pkt)

	// handle with new goroutine
	go func() {
		if evt.Pubkey != *glo_val.SelfPubkey || src == math.MaxUint64 {
			switch evt.Kind {
			case KIND_EVT_PROFILE: // profile
				// store received profile data
				prof := GenProfileFromEvent(evt)
				mm.DataMan.StoreProfile(prof)
				if prof.Pubkey64bit == glo_val.SelfPubkey64bit && glo_val.ProfileMyOwn.UpdatedAt < evt.Created_at {
					// this route works only when recovery
					glo_val.ProfileMyOwn = prof
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
						recvdTime > mm.DataMan.GetProfileLocal(np2p_util.GetLower64bitUint(evt.Pubkey)).UpdatedAt {
						// TODO: need to implement limitation of request times (MessageManager::handleRecvMsgBcastEvt)
						// profile is updated. request latest profile asynchronous.
						go mm.UnicastProfileReq(shortId)
					}
				}
				//// display (TEMPORAL IMPL)
				//mm.DispPostAtStdout(evt)
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
func (mm *MessageManager) handleRecvMsgBcastReq(src mesh.PeerName, pkt *schema.Np2pPacket, req *schema.Np2pReq) error {
	go func() {
		if src != mesh.PeerName(glo_val.SelfPubkey64bit) {
			switch req.Kind {
			case KIND_REQ_SHARE_EVT_DATA: // need to having event datas
				go mm.UnicastHavingEvtData(src)
			default:
				fmt.Println("received unknown kind req: " + strconv.Itoa(int(req.Kind)))
			}
			np2p_util.Np2pDbgPrintln("received request: " + strconv.Itoa(int(req.Kind)))
		}

	}()

	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src mesh.PeerName, pkt *schema.Np2pPacket) error {
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
					mm.DataMan.StoreProfile(GenProfileFromEvent(evt))
				case KIND_EVT_POST: // response of KIND_REQ_SHARE_EVT_DATA
					//// display (TEMPORAL IMPL)
					//mm.DispPostAtStdout(evt)
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

func (mm *MessageManager) SendMsgUnicast(dst mesh.PeerName, pkt *schema.Np2pPacket) {
	mm.send.GossipUnicast(dst, pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt *schema.Np2pPacket) {
	mm.send.GossipBroadcast(pkt)
}

func (mm *MessageManager) BcastOwnPost(evt *schema.Np2pEvent) {
	//pubSlice := glo_val.SelfPubkey[:]
	//var sigBytes [np2p_const.SignatureSize]byte
	//copy(sigBytes[:], pubSlice)
	//tagsMap := make(map[string][]interface{})
	//tagsMap["nickname"] = []interface{}{*glo_val.Nickname}
	//if np2p_util.IsHit(np2p_const.AttachProfileUpdateProb) && glo_val.ProfileMyOwn.UpdatedAt > 0 {
	//	tagsMap["u"] = []interface{}{glo_val.ProfileMyOwn.UpdatedAt}
	//}
	//event := schema.Np2pEvent{
	//	Id:         np2p_util.GetRandUint64(),
	//	Pubkey:     *glo_val.SelfPubkey,
	//	Created_at: time.Now().Unix(),
	//	Kind:       KIND_EVT_POST,
	//	Tags:       tagsMap,
	//	Content:    content,
	//	Sig:        &sigBytes,
	//}
	//events := []*schema.Np2pEvent{&event}

	//for _, peerId := range s.buzzPeer.GetPeerList() {
	//	s.buzzPeer.MessageMan.SendMsgUnicast(peerId, schema.NewNp2pPacket(&events, nil, nil))
	//}

	events := []*schema.Np2pEvent{evt}
	mm.SendMsgBroadcast(schema.NewNp2pPacket(&events, nil))
	// store own issued event
	mm.DataMan.StoreEvent(evt)

	//return &event
	//fmt.Println(event.Tags["nickname"][0] + "> " + event.Content)
}

// func (mm *MessageManager) BcastOwnProfile(name *string, about *string, picture *string) *schema.Np2pProfile {
func (mm *MessageManager) BcastOwnProfile(evt *schema.Np2pEvent) *schema.Np2pProfile {
	//event := mm.constructProfileEvt(name, about, picture)
	events := []*schema.Np2pEvent{evt}
	mm.SendMsgBroadcast(schema.NewNp2pPacket(&events, nil))
	mm.DataMan.StoreEvent(evt)

	storeProf := GenProfileFromEvent(evt)
	mm.DataMan.StoreProfile(storeProf)

	return storeProf
}

//func (mm *MessageManager) constructProfileEvt(name *string, about *string, picture *string) *schema.Np2pEvent {
//	pubSlice := glo_val.SelfPubkey[:]
//	var sigBytes [np2p_const.SignatureSize]byte
//	copy(sigBytes[:], pubSlice)
//	tagsMap := make(map[string][]interface{})
//	tagsMap["name"] = []interface{}{*name}
//	tagsMap["about"] = []interface{}{*about}
//	tagsMap["picture"] = []interface{}{*picture}
//	event := schema.Np2pEvent{
//		Id:         np2p_util.GetRandUint64(),
//		Pubkey:     *glo_val.SelfPubkey,
//		Created_at: time.Now().Unix(),
//		Kind:       KIND_EVT_PROFILE,
//		Tags:       tagsMap,
//		Content:    "",
//		Sig:        &sigBytes,
//	}
//	return &event
//}

func (mm *MessageManager) UnicastProfileReq(pubkey64bit uint64) {
	reqs := []*schema.Np2pReq{schema.NewNp2pReq(KIND_REQ_SHARE_EVT_DATA, nil)}
	pkt := schema.NewNp2pPacket(nil, &reqs)
	mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), pkt)
}

// used for response of profile request
func (mm *MessageManager) UnicastOwnProfile(pubkey64bit uint64) {
	if glo_val.CurrentProfileEvt != nil {
		// send latest profile data
		events := []*schema.Np2pEvent{glo_val.CurrentProfileEvt}
		mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), schema.NewNp2pPacket(&events, nil))
	}
}

//// TODO: TEMPORAL IMPL
//func (mm *MessageManager) DispPostAtStdout(evt *schema.Np2pEvent) {
//	fmt.Println(evt.Tags["nickname"][0].(string) + "> " + evt.Content)
//}

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
func (mm *MessageManager) UnicastHavingEvtData(dest mesh.PeerName) {
	events := mm.DataMan.GetLatestEvents(time.Now().Unix()-3*24*3600, math.MaxInt64)
	pkt := schema.NewNp2pPacket(events, nil)
	mm.SendMsgUnicast(dest, pkt)
	np2p_util.Np2pDbgPrintln("UnicastHavingEvtData: sent " + strconv.Itoa(len(*events)) + " events")
}
