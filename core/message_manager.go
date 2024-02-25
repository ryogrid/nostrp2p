package core

import (
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_const"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/glo_val"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"math"
	"strconv"
	"time"
)

const (
	KIND_EVT_PROFILE        = 0
	KIND_EVT_POST           = 1
	KIND_REQ_PROFILE        = KIND_EVT_PROFILE
	KIND_REQ_SHARE_EVT_DATA = 40000
)

type MessageManager struct {
	DataMan *DataManager
	send    mesh.Gossip // set by BuzzPeer.Register
}

// when recovery, src is math.MaxUint64
func (mm *MessageManager) handleRecvMsgBcast(src mesh.PeerName, pkt *schema.BuzzPacket) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)

	if pkt.Events != nil {
		// handle with new goroutine
		go func() {
			for _, evt := range pkt.Events {
				if evt.Pubkey != *glo_val.SelfPubkey || src == math.MaxUint64 {
					switch evt.Kind {
					case KIND_EVT_PROFILE: // profile
						// store received profile data
						prof := GenProfileFromEvent(evt)
						mm.DataMan.StoreProfile(prof)
						if prof.Pubkey64bit == glo_val.SelfPubkey64bit && glo_val.ProfileMyOwn.UpdatedAt < evt.Created_at {
							// this route works only when recovery
							glo_val.ProfileMyOwn = prof
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
							shortId := buzz_util.GetLower64bitUint(evt.Pubkey)
							if mm.DataMan.GetProfileLocal(shortId) == nil ||
								recvdTime > mm.DataMan.GetProfileLocal(buzz_util.GetLower64bitUint(evt.Pubkey)).UpdatedAt {
								// TODO: need to implement limitation of request times (MessageManager::handleRecvMsgBcast)
								// profile is updated. request latest profile asynchronous.
								go mm.UnicastProfileReq(shortId)
							}
						}
						// display (TEMPORAL IMPL)
						mm.DispPostAtStdout(evt)
					default:
						fmt.Println("received unknown kind event: " + strconv.Itoa(int(evt.Kind)))
					}

					// store received event data (on memory)
					tmpEvt := *evt
					mm.DataMan.StoreEvent(&tmpEvt)
				}
			}
		}()
	}

	if pkt.Reqs != nil {
		for _, req := range pkt.Reqs {
			if src != mesh.PeerName(glo_val.SelfPubkey64bit) {
				switch req.Kind {
				case KIND_REQ_SHARE_EVT_DATA: // need to having event datas
					go mm.UnicastHavingEvtData(src)
				default:
					fmt.Println("received unknown kind req: " + strconv.Itoa(int(req.Kind)))
				}
				buzz_util.BuzzDbgPrintln("received request: " + strconv.Itoa(int(req.Kind)))
			}
		}
	}
	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src mesh.PeerName, pkt *schema.BuzzPacket) error {
	if pkt.Events != nil {
		// handle with new goroutine
		go func() {
			// handle response of request

			// store received event data (on memory)
			mm.DataMan.StoreEvent(pkt.Events[0])

			switch pkt.Events[0].Kind {
			case KIND_EVT_PROFILE: // profile
				// store received profile data
				mm.DataMan.StoreProfile(GenProfileFromEvent(pkt.Events[0]))
			default:
				fmt.Println("received unknown kind event: " + strconv.Itoa(int(pkt.Events[0].Kind)))
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

func (mm *MessageManager) SendMsgUnicast(dst mesh.PeerName, pkt *schema.BuzzPacket) {
	mm.send.GossipUnicast(dst, pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt *schema.BuzzPacket) {
	mm.send.GossipBroadcast(pkt)
}

func (mm *MessageManager) BcastOwnPost(content string) *schema.BuzzEvent {
	pubSlice := glo_val.SelfPubkey[:]
	var sigBytes [buzz_const.SignatureSize]byte
	copy(sigBytes[:], pubSlice)
	tagsMap := make(map[string][]interface{})
	tagsMap["nickname"] = []interface{}{*glo_val.Nickname}
	if buzz_util.IsHit(buzz_const.AttachProfileUpdateProb) && glo_val.ProfileMyOwn.UpdatedAt > 0 {
		tagsMap["u"] = []interface{}{glo_val.ProfileMyOwn.UpdatedAt}
	}
	event := schema.BuzzEvent{
		Id:         buzz_util.GetRandUint64(),
		Pubkey:     *glo_val.SelfPubkey,
		Created_at: time.Now().Unix(),
		Kind:       KIND_EVT_POST,
		Tags:       tagsMap,
		Content:    content,
		Sig:        &sigBytes,
	}
	events := []*schema.BuzzEvent{&event}
	//for _, peerId := range s.buzzPeer.GetPeerList() {
	//	s.buzzPeer.MessageMan.SendMsgUnicast(peerId, schema.NewBuzzPacket(&events, nil, nil))
	//}
	mm.SendMsgBroadcast(schema.NewBuzzPacket(&events, nil))
	// store own issued event
	mm.DataMan.StoreEvent(&event)

	return &event
	//fmt.Println(event.Tags["nickname"][0] + "> " + event.Content)
}

func (mm *MessageManager) BcastOwnProfile(name *string, about *string, picture *string) *schema.BuzzProfile {
	event := mm.constructProfileEvt(name, about, picture)
	events := []*schema.BuzzEvent{event}
	mm.SendMsgBroadcast(schema.NewBuzzPacket(&events, nil))
	mm.DataMan.StoreEvent(event)

	storeProf := GenProfileFromEvent(event)
	mm.DataMan.StoreProfile(storeProf)

	return storeProf
}

func (mm *MessageManager) constructProfileEvt(name *string, about *string, picture *string) *schema.BuzzEvent {
	pubSlice := glo_val.SelfPubkey[:]
	var sigBytes [buzz_const.SignatureSize]byte
	copy(sigBytes[:], pubSlice)
	tagsMap := make(map[string][]interface{})
	tagsMap["name"] = []interface{}{*name}
	tagsMap["about"] = []interface{}{*about}
	tagsMap["picture"] = []interface{}{*picture}
	event := schema.BuzzEvent{
		Id:         buzz_util.GetRandUint64(),
		Pubkey:     *glo_val.SelfPubkey,
		Created_at: time.Now().Unix(),
		Kind:       KIND_EVT_PROFILE,
		Tags:       tagsMap,
		Content:    "",
		Sig:        &sigBytes,
	}
	return &event
}

func (mm *MessageManager) UnicastProfileReq(pubkey64bit uint64) {
	reqs := []*schema.BuzzReq{schema.NewBuzzReq(KIND_REQ_SHARE_EVT_DATA, nil)}
	pkt := schema.NewBuzzPacket(nil, &reqs)
	mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), pkt)
}

// used for response of profile request
func (mm *MessageManager) UnicastOwnProfile(pubkey64bit uint64) {
	evt := mm.constructProfileEvt(&glo_val.ProfileMyOwn.Name, &glo_val.ProfileMyOwn.About, &glo_val.ProfileMyOwn.Picture)
	events := []*schema.BuzzEvent{evt}
	mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), schema.NewBuzzPacket(&events, nil))
}

// TODO: TEMPORAL IMPL
func (mm *MessageManager) DispPostAtStdout(evt *schema.BuzzEvent) {
	fmt.Println(evt.Tags["nickname"][0].(string) + "> " + evt.Content)
}

func GenProfileFromEvent(evt *schema.BuzzEvent) *schema.BuzzProfile {
	return &schema.BuzzProfile{
		Pubkey64bit: buzz_util.GetLower64bitUint(evt.Pubkey),
		Name:        evt.Tags["name"][0].(string),
		About:       evt.Tags["about"][0].(string),
		Picture:     evt.Tags["picture"][0].(string),
		UpdatedAt:   evt.Created_at,
	}
}

// TODO: TEMPORAL IMPL
func (mm *MessageManager) BcastShareEvtDataReq() {
	reqs := []*schema.BuzzReq{schema.NewBuzzReq(KIND_REQ_SHARE_EVT_DATA, nil)}

	pkt := schema.NewBuzzPacket(nil, &reqs)
	mm.SendMsgBroadcast(pkt)
}

// TODO: TEMPORAL IMPL
// send latest 3days events
func (mm *MessageManager) UnicastHavingEvtData(dest mesh.PeerName) {
	events := mm.DataMan.GetLatestEvents(time.Now().Unix() - 3*26*3600)
	pkt := schema.NewBuzzPacket(events, nil)
	mm.SendMsgUnicast(dest, pkt)
}
