package core

import (
	"errors"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"github.com/weaveworks/mesh"
	"log"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type Np2pPeer struct {
	//Send            *mesh.Gossip
	actions         chan<- func()
	quit            chan struct{}
	logger          *log.Logger
	dataMan         *DataManager
	MessageMan      *MessageManager
	SelfId          mesh.PeerName
	SelfPubkey      [np2p_const.PubkeySize]byte
	Router          *mesh.Router
	recvedEvtReqMap map[uint64]struct{}
}

// Np2pPeer implements mesh.Gossiper.
var _ mesh.Gossiper = &Np2pPeer{}

// Construct a Np2pPeer with empty state.
// Be sure to Register a channel, later,
// so we can make outbound communication.
func NewPeer(self mesh.PeerName, logger *log.Logger) *Np2pPeer {
	actions := make(chan func())
	dataMan := NewDataManager()
	msgMan := &MessageManager{DataMan: dataMan}

	p := &Np2pPeer{
		//Send:            nil, // must .Register() later
		actions:         actions,
		quit:            make(chan struct{}),
		logger:          logger,
		dataMan:         dataMan,
		MessageMan:      msgMan,
		SelfId:          self,
		recvedEvtReqMap: make(map[uint64]struct{}), // make(map[uint64]struct{}),
	}
	go p.loop(actions)
	return p
}

func (p *Np2pPeer) loop(actions <-chan func()) {
	for {
		select {
		case f := <-actions:
			f()
		case <-p.quit:
			return
		}
	}
}

// Register the result of a mesh.Router.NewGossip.
func (p *Np2pPeer) Register(send mesh.Gossip) {
	p.actions <- func() {
		//p.Send = &send
		p.MessageMan.send = send
	}
}

func (p *Np2pPeer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *Np2pPeer) Gossip() (complete mesh.GossipData) {
	np2p_util.Np2pDbgPrintln("Gossip called")
	//return &schema.Np2pPacket{}
	return nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *Np2pPeer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	np2p_util.Np2pDbgPrintln("OnGossip called")
	//return &schema.Np2pPacket{}, nil
	var retData schema.EncodableAndMergeable = &schema.Np2pPacket{}
	return retData.(mesh.GossipData), nil
}

func (p *Np2pPeer) OnRecvBroadcast(src uint64, buf []byte) (received schema.EncodableAndMergeable, err error) {
	//var pkt schema.Np2pPacket
	///if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
	pkt, err_ := schema.NewNp2pPacketFromBytes(buf)
	if err_ != nil {
		return nil, err_
	}
	if pkt.PktVer != np2p_const.PacketStructureVersion {
		return nil, errors.New("Invalid packet version")
	}
	if pkt.SrvVer != np2p_const.ServerImplVersion {
		fmt.Println("received packat from newer version of server")
	}

	tmpEvts := make([]*schema.Np2pEvent, 0)
	tmpReqs := make([]*schema.Np2pReq, 0)
	retPkt := schema.NewNp2pPacket(&tmpEvts, &tmpReqs)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if _, ok := p.recvedEvtReqMap[np2p_util.ExtractUint64FromBytes(evt.Id[:])]; !ok {
				err2 := p.MessageMan.handleRecvMsgBcastEvt(src, pkt, evt)
				if err2 != nil {
					panic(err2)
				}

				p.recvedEvtReqMap[np2p_util.ExtractUint64FromBytes(evt.Id[:])] = struct{}{}
				retPkt.Events = append(retPkt.Events, evt)
			} else {
				continue
			}
		}
	} else if pkt.Reqs != nil {
		for _, req := range pkt.Reqs {
			if _, ok := p.recvedEvtReqMap[req.Id]; !ok {
				err2 := p.MessageMan.handleRecvMsgBcastReq(src, pkt, req)
				if err2 != nil {
					panic(err2)
				}

				p.recvedEvtReqMap[req.Id] = struct{}{}
				retPkt.Reqs = append(retPkt.Reqs, req)
			} else {
				continue
			}
		}
	} else {
		return pkt, nil
	}

	if len(retPkt.Events) == 0 && len(retPkt.Reqs) == 0 {
		return nil, nil
	} else {
		return retPkt, nil
	}

	//return &pkt, nil
	//return &schema.Np2pPacket{}, nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *Np2pPeer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	np2p_util.Np2pDbgPrintln("OnGossipBroadcast called")
	recved, err_ := p.OnRecvBroadcast(uint64(src), buf)
	return recved.(mesh.GossipData), err_
}

func (p *Np2pPeer) OnRecvUnicast(src uint64, buf []byte) (err error) {
	//var pkt schema.Np2pPacket
	//if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
	//	return err_
	//}
	pkt, err := schema.NewNp2pPacketFromBytes(buf)
	if err != nil {
		return err
	}
	if pkt.PktVer != np2p_const.PacketStructureVersion {
		return errors.New("Invalid packet version")
	}
	if pkt.SrvVer != np2p_const.ServerImplVersion {
		fmt.Println("received packat from newer version of server")
	}

	err_ := p.MessageMan.handleRecvMsgUnicast(src, pkt)
	if err_ != nil {
		panic(err_)
	}

	return nil
}

// Merge the gossiped data represented by buf into our state.
func (p *Np2pPeer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	np2p_util.Np2pDbgPrintln("OnGossipUnicast called")
	return p.OnRecvUnicast(uint64(src), buf)
}

func (p *Np2pPeer) GetPeerList() []mesh.PeerName {
	tmpMap := p.Router.Routes.PeerNames()
	retArr := make([]mesh.PeerName, 0)
	for k, _ := range tmpMap {
		retArr = append(retArr, k)
	}
	return retArr
}
