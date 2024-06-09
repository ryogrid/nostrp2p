package core

import (
	"errors"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type Np2pPeer struct {
	//send            *mesh.Gossip
	Actions         chan<- func()
	quit            chan struct{}
	logger          *log.Logger
	dataMan         DataManager
	MessageMan      *MessageManager
	SelfId          uint64 //mesh.PeerName
	SelfPubkey      [np2p_const.PubkeySize]byte
	recvedEvtReqMap map[uint64]struct{}
}

// Construct a Np2pPeer with empty state.
// Be sure to Register a channel, later,
// so we can make outbound communication.
func NewPeer(self uint64, logger *log.Logger) *Np2pPeer {
	actions := make(chan func())
	//dataMan := NewOnMemoryDataManager()
	dataMan := NewNutsDBDataManager()

	msgMan := NewMessageManager(dataMan)

	p := &Np2pPeer{
		Actions:         actions,
		quit:            make(chan struct{}),
		logger:          logger,
		dataMan:         dataMan,
		MessageMan:      msgMan,
		SelfId:          self,
		recvedEvtReqMap: make(map[uint64]struct{}), // make(map[uint64]struct{}),
	}
	go p.loop(actions)

	// start event resender
	msgMan.evtReSender.Start()

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

func (p *Np2pPeer) stop() {
	close(p.quit)
}

func (p *Np2pPeer) OnRecvBroadcast(src uint64, buf []byte) (received schema.EncodableAndMergeable, err error) {
	tmpEvts := make([]*schema.Np2pEvent, 0)
	tmpReqs := make([]*schema.Np2pReq, 0)
	retPkt := schema.NewNp2pPacket(&tmpEvts, &tmpReqs)

	pkt, err_ := schema.NewNp2pPacketFromBytes(buf)
	if err_ != nil {
		// returns NP2pPacket having zero length fields
		// this means received data is already known to mesh library...
		fmt.Println("received strange packet. decoding failed. err = ", err_)
		return retPkt, nil
	}
	if pkt.PktVer != np2p_const.PacketStructureVersion {
		return nil, errors.New("Invalid packet version")
	}
	if pkt.SrvVer != np2p_const.ServerImplVersion {
		fmt.Println("received packat from newer version of server")
	}

	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if _, ok := p.recvedEvtReqMap[np2p_util.ExtractUint64FromBytes(evt.Id[:])]; !ok {
				if evt.Verify() == false {
					// invalid signiture
					fmt.Println("invalid signiture")
					continue
				}

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
		fmt.Println("received empty packet")
		return pkt, nil
	}

	if len(retPkt.Events) == 0 && len(retPkt.Reqs) == 0 {
		fmt.Println("received strange packet")
		// returns NP2pPacket having zero length fields
		// this means received data is already known to mesh library...
		return pkt, nil
	} else {
		return retPkt, nil
	}
}

func (p *Np2pPeer) OnRecvUnicast(src uint64, buf []byte) (err error) {
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

/*
func (p *Np2pPeer) GetPeerList() []mesh.PeerName {
	tmpMap := p.Router.Routes.PeerNames()
	retArr := make([]mesh.PeerName, 0)
	for k, _ := range tmpMap {
		retArr = append(retArr, k)
	}
	return retArr
}
*/
