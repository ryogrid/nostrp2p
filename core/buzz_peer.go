package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_const"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/glo_val"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"log"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type BuzzPeer struct {
	Send            *mesh.Gossip
	actions         chan<- func()
	quit            chan struct{}
	logger          *log.Logger
	dataMan         *DataManager
	MessageMan      *MessageManager
	SelfId          mesh.PeerName
	SelfPubkey      [buzz_const.PubkeySize]byte
	Router          *mesh.Router
	recvedEvtReqMap map[uint64]struct{}
}

// BuzzPeer implements mesh.Gossiper.
var _ mesh.Gossiper = &BuzzPeer{}

// Construct a BuzzPeer with empty state.
// Be sure to Register a channel, later,
// so we can make outbound communication.
func NewPeer(self mesh.PeerName, logger *log.Logger) *BuzzPeer {
	buf := make([]byte, binary.MaxVarintLen64)
	// TODO: need to set correct pubkey
	binary.PutUvarint(buf, uint64(self))
	var pubkeyBytes [buzz_const.PubkeySize]byte
	copy(pubkeyBytes[:], buf)
	glo_val.SelfPubkey = &pubkeyBytes
	glo_val.SelfPubkey64bit = uint64(self)

	actions := make(chan func())
	dataMan := NewDataManager()
	msgMan := &MessageManager{DataMan: dataMan}

	p := &BuzzPeer{
		Send:            nil, // must .Register() later
		actions:         actions,
		quit:            make(chan struct{}),
		logger:          logger,
		dataMan:         dataMan,
		MessageMan:      msgMan,
		SelfId:          self,
		recvedEvtReqMap: make(map[uint64]struct{}),
	}
	go p.loop(actions)
	return p
}

func (p *BuzzPeer) loop(actions <-chan func()) {
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
func (p *BuzzPeer) Register(send mesh.Gossip) {
	p.actions <- func() {
		p.Send = &send
		p.MessageMan.send = send
	}
}

func (p *BuzzPeer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *BuzzPeer) Gossip() (complete mesh.GossipData) {
	buzz_util.BuzzDbgPrintln("Gossip called")
	//return &schema.BuzzPacket{}
	return nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *BuzzPeer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	buzz_util.BuzzDbgPrintln("OnGossip called")
	return &schema.BuzzPacket{}, nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *BuzzPeer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	buzz_util.BuzzDbgPrintln("OnGossipBroadcast called")
	//var pkt schema.BuzzPacket
	///if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
	pkt, err_ := schema.NewBuzzPacketFromBytes(buf)
	if err_ != nil {
		return nil, err_
	}
	if pkt.PktVer != buzz_const.PacketStructureVersion {
		return nil, errors.New("Invalid packet version")
	}
	if pkt.SrvVer != buzz_const.ServerImplVersion {
		fmt.Println("received packat from newer version of server")
	}

	tmpEvts := make([]*schema.BuzzEvent, 0)
	tmpReqs := make([]*schema.BuzzReq, 0)
	retPkt := schema.NewBuzzPacket(&tmpEvts, &tmpReqs)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if _, ok := p.recvedEvtReqMap[evt.Id]; !ok {
				err2 := p.MessageMan.handleRecvMsgBcastEvt(src, pkt, evt)
				if err2 != nil {
					panic(err2)
				}

				p.recvedEvtReqMap[evt.Id] = struct{}{}
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
	//return &schema.BuzzPacket{}, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *BuzzPeer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	buzz_util.BuzzDbgPrintln("OnGossipUnicast called")
	//var pkt schema.BuzzPacket
	//if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
	//	return err_
	//}
	pkt, err := schema.NewBuzzPacketFromBytes(buf)
	if err != nil {
		return err
	}
	if pkt.PktVer != buzz_const.PacketStructureVersion {
		return errors.New("Invalid packet version")
	}
	if pkt.SrvVer != buzz_const.ServerImplVersion {
		fmt.Println("received packat from newer version of server")
	}

	err_ := p.MessageMan.handleRecvMsgUnicast(src, pkt)
	if err_ != nil {
		panic(err_)
	}

	return nil
}

func (p *BuzzPeer) GetPeerList() []mesh.PeerName {
	tmpMap := p.Router.Routes.PeerNames()
	retArr := make([]mesh.PeerName, 0)
	for k, _ := range tmpMap {
		retArr = append(retArr, k)
	}
	return retArr
}
