package core

import (
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
)

type MessageManager struct {
	dataManager *DataManager
	send        mesh.Gossip // set by BuzzPeer.Register
}

func (mm *MessageManager) handleRecvMsgBcast(pkt *schema.BuzzPacket) error {
	mm.dataManager.storeReceived(pkt)

	return nil
}

func (mm *MessageManager) handleRecvMsgUnicast(src mesh.PeerName, pkt *schema.BuzzPacket) error {
	mm.dataManager.storeReceived(pkt)

	return nil
}

func (mm *MessageManager) SendMsgUnicast(dst mesh.PeerName, pkt *schema.BuzzPacket) {
	mm.send.GossipUnicast(dst, pkt.Encode()[0])
}

func (mm *MessageManager) SendMsgBroadcast(pkt *schema.BuzzPacket) {
	mm.send.GossipBroadcast(pkt)
}
