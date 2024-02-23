package core

import (
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/glo_val"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"time"
)

type MessageManager struct {
	DataMan *DataManager
	send    mesh.Gossip // set by BuzzPeer.Register
}

func (mm *MessageManager) handleRecvMsgBcast(pkt *schema.BuzzPacket) error {
	mm.DataMan.handleReceived(pkt)

	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src mesh.PeerName, pkt *schema.BuzzPacket) error {
	mm.DataMan.handleReceived(pkt)

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
	events := []*schema.BuzzEvent{&event}
	mm.SendMsgBroadcast(schema.NewBuzzPacket(&events, nil))
	mm.DataMan.StoreEvent(&event)

	storeProf := GenProfileFromEvent(&event)
	mm.DataMan.StoreProfile(storeProf)

	return storeProf
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
