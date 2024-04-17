package transport

import (
	"github.com/ryogrid/nostrp2p/core"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"github.com/weaveworks/mesh"
)

type MeshTransport struct {
	peer   *core.Np2pPeer
	send   mesh.Gossip
	router *mesh.Router
}

func NewMeshTransport(peer *core.Np2pPeer) *MeshTransport {
	return &MeshTransport{peer: peer}
}

// MeshTransport implements mesh.Gossiper.
var _ mesh.Gossiper = &MeshTransport{}

// Return a copy of our complete state.
func (mt *MeshTransport) Gossip() (complete mesh.GossipData) {
	np2p_util.Np2pDbgPrintln("Gossip called")
	//return &schema.Np2pPacket{}
	return nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (mt *MeshTransport) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	np2p_util.Np2pDbgPrintln("OnGossip called")
	//return &schema.Np2pPacket{}, nil
	var retData schema.EncodableAndMergeable = &schema.Np2pPacket{}
	return retData.(mesh.GossipData), nil
}

// Merge the gossiped data represented by buf into our state.
func (mt *MeshTransport) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	np2p_util.Np2pDbgPrintln("OnGossipUnicast called")
	return mt.peer.OnRecvUnicast(uint64(src), buf)
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (mt *MeshTransport) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	np2p_util.Np2pDbgPrintln("OnGossipBroadcast called")
	recved, err_ := mt.peer.OnRecvBroadcast(uint64(src), buf)
	return recved.(mesh.GossipData), err_
}

// Register the result of a mesh.Router.NewGossip.
func (mt *MeshTransport) Register(send mesh.Gossip) {
	//p.peer.Actions <- func() {
	//	//p.send = &send
	//	p.peer.MessageMan.send = send
	//}
	mt.send = send
}

func (mt *MeshTransport) SendMsgUnicast(dst uint64, buf []byte) error {
	return mt.send.GossipUnicast(mesh.PeerName(dst&0x0000ffffffffffff), buf)
}

func (mt *MeshTransport) SendMsgBroadcast(pkt schema.EncodableAndMergeable) {
	mt.send.GossipBroadcast(pkt.(mesh.GossipData))
}

func (mt *MeshTransport) SetRouter(router *mesh.Router) {
	mt.router = router
}
