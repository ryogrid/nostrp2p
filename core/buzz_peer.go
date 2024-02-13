package core

import (
	"bytes"
	"encoding/gob"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"log"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type BuzzPeer struct {
	send       *mesh.Gossip
	actions    chan<- func()
	quit       chan struct{}
	logger     *log.Logger
	dataMan    *DataManager
	MessageMan *MessageManager
	selfId     mesh.PeerName
}

// BuzzPeer implements mesh.Gossiper.
var _ mesh.Gossiper = &BuzzPeer{}

// Construct a BuzzPeer with empty state.
// Be sure to Register a channel, later,
// so we can make outbound communication.
func NewPeer(self mesh.PeerName, logger *log.Logger) *BuzzPeer {
	actions := make(chan func())
	dataMan := &DataManager{}
	msgMan := &MessageManager{dataManager: dataMan}
	p := &BuzzPeer{
		send:       nil, // must .Register() later
		actions:    actions,
		quit:       make(chan struct{}),
		logger:     logger,
		dataMan:    dataMan,
		MessageMan: msgMan,
		selfId:     self,
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
		p.send = &send
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
	var pkt schema.BuzzPacket
	if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
		return nil, err_
	}

	err_ := p.MessageMan.handleRecvMsgBcast(&pkt)
	if err_ != nil {
		panic(err_)
	}

	return &pkt, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *BuzzPeer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	buzz_util.BuzzDbgPrintln("OnGossipUnicast called")
	var pkt schema.BuzzPacket
	if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
		return err_
	}

	err_ := p.MessageMan.handleRecvMsgUnicast(src, &pkt)
	if err_ != nil {
		panic(err_)
	}

	return nil
}
