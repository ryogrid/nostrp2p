package core

import (
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/glo_val"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"strconv"
	"time"
)

type MessageManager struct {
	DataMan *DataManager
	send    mesh.Gossip // set by BuzzPeer.Register
}

func (mm *MessageManager) handleRecvMsgBcast(pkt *schema.BuzzPacket) error {
	// TODO: need to use on-disk DB (DataManager::mergeReceived)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if evt.Pubkey != *glo_val.SelfPubkey {
				// store received event data (on memory)
				tmpEvt := *evt
				mm.DataMan.StoreEvent(&tmpEvt)

				switch evt.Kind {
				case 0: // profile
					// store received profile data
					mm.DataMan.StoreProfile(GenProfileFromEvent(evt))
				case 1: // post
					// display (TEMPORAL IMPL)
					if val, ok := evt.Tags["u"]; ok {
						// profile update time is attached
						//(periodically attached the tag for avoiding old profile is kept)
						recvdTime, err := strconv.ParseInt("17"+val[0], 10, 64)
						if err != nil {
							fmt.Println("strconv.Atoi error: " + err.Error())
						}
						shortId := buzz_util.GetLower64bitUint(evt.Pubkey)
						if mm.DataMan.GetProfileLocal(shortId) == nil ||
							recvdTime > mm.DataMan.GetProfileLocal(buzz_util.GetLower64bitUint(evt.Pubkey)).UpdatedAt {
							// TODO: need to implement limitation of request times (MessageManager::handleRecvMsgBcast)
							// profile is updated. request latest profile asynchronous.
							go mm.RequestProfile(shortId)
						}
					}
					mm.DispPostAtStdout(evt)
				default:
					fmt.Println("received unknown kind event: " + strconv.Itoa(int(evt.Kind)))
				}
			}
		}
	} else {
		fmt.Println("pkt.Events is nil")
	}
	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src mesh.PeerName, pkt *schema.BuzzPacket) error {
	if pkt.Events != nil {
		// handle response of request

		// store received event data (on memory)
		mm.DataMan.StoreEvent(pkt.Events[0])

		switch pkt.Events[0].Kind {
		case 0: // profile
			// store received profile data
			mm.DataMan.StoreProfile(GenProfileFromEvent(pkt.Events[0]))
		default:
			fmt.Println("received unknown kind event: " + strconv.Itoa(int(pkt.Events[0].Kind)))
		}
		return nil
	}

	if pkt.Req != nil {
		// handle request
		switch pkt.Req.Kind {
		case 0: // profile request
		// send profile data

		default:
			fmt.Println("received unknown kind request: " + strconv.Itoa(int(pkt.Req.Kind)))
		}
		return nil
	}

	fmt.Println("MessageManager::handleRecvMsgUnicast: received unknown kind message")
	return nil
}

func (mm *MessageManager) SendMsgUnicast(dst mesh.PeerName, pkt *schema.BuzzPacket) {
	mm.send.GossipUnicast(dst, pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt *schema.BuzzPacket) {
	mm.send.GossipBroadcast(pkt)
}

func (mm *MessageManager) BrodcastOwnPost(content string) *schema.BuzzEvent {
	pubSlice := glo_val.SelfPubkey[:]
	var sigBytes [64]byte
	copy(sigBytes[:], pubSlice)
	tagsMap := make(map[string][]string)
	tagsMap["nickname"] = []string{*glo_val.Nickname}
	if buzz_util.IsHit(buzz_util.AttachProfileUpdateProb) {
		// remove head 17 for data size reduction
		updatedAt := strconv.FormatInt(glo_val.ProfileMyOwn.UpdatedAt, 10)[2:]
		tagsMap["u"] = []string{updatedAt}
	}
	event := schema.BuzzEvent{
		Id:         buzz_util.GetRandUint64(),
		Pubkey:     *glo_val.SelfPubkey,
		Created_at: time.Now().Unix(),
		Kind:       1,
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

func (mm *MessageManager) BrodcastOwnProfile(name *string, about *string, picture *string) *schema.BuzzProfile {
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
	var sigBytes [64]byte
	copy(sigBytes[:], pubSlice)
	tagsMap := make(map[string][]string)
	tagsMap["name"] = []string{*name}
	tagsMap["about"] = []string{*about}
	tagsMap["picture"] = []string{*picture}
	event := schema.BuzzEvent{
		Id:         buzz_util.GetRandUint64(),
		Pubkey:     *glo_val.SelfPubkey,
		Created_at: time.Now().Unix(),
		Kind:       0,
		Tags:       tagsMap,
		Content:    "",
		Sig:        &sigBytes,
	}
	return &event
}

func (mm *MessageManager) RequestProfile(pubkey64bit uint64) {
	req := &schema.BuzzReq{
		Kind: 0,
	}
	pkt := schema.NewBuzzPacket(nil, req)
	mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), pkt)
}

// used for response of profile request
func (mm *MessageManager) SendProfileUnicast(pubkey64bit uint64) {
	evt := mm.constructProfileEvt(&glo_val.ProfileMyOwn.Name, &glo_val.ProfileMyOwn.About, &glo_val.ProfileMyOwn.Picture)
	events := []*schema.BuzzEvent{evt}
	mm.SendMsgUnicast(mesh.PeerName(pubkey64bit), schema.NewBuzzPacket(&events, nil))
}

// TODO: TEMPORAL IMPL
func (mm *MessageManager) DispPostAtStdout(evt *schema.BuzzEvent) {
	fmt.Println(evt.Tags["nickname"][0] + "> " + evt.Content)
}

func GenProfileFromEvent(evt *schema.BuzzEvent) *schema.BuzzProfile {
	return &schema.BuzzProfile{
		Pubkey64bit: buzz_util.GetLower64bitUint(evt.Pubkey),
		Name:        evt.Tags["name"][0],
		About:       evt.Tags["about"][0],
		Picture:     evt.Tags["picture"][0],
		UpdatedAt:   evt.Created_at,
	}
}
