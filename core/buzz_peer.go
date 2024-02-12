package core

import (
	"bytes"
	"encoding/gob"
	"github.com/ryogrid/buzzoon/schema"
	"github.com/weaveworks/mesh"
	"log"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type BuzzPeer struct {
	send        mesh.Gossip
	actions     chan<- func()
	quit        chan struct{}
	logger      *log.Logger
	dataManager *DataManager
}

// BuzzPeer implements mesh.Gossiper.
var _ mesh.Gossiper = &BuzzPeer{}

// Construct a BuzzPeer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func newPeer(self mesh.PeerName, logger *log.Logger) *BuzzPeer {
	actions := make(chan func())
	p := &BuzzPeer{
		send:        nil, // must .register() later
		actions:     actions,
		quit:        make(chan struct{}),
		logger:      logger,
		dataManager: &DataManager{},
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

// register the result of a mesh.Router.NewGossip.
func (p *BuzzPeer) register(send mesh.Gossip) {
	p.actions <- func() { p.send = send }
}

func (p *BuzzPeer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *BuzzPeer) Gossip() (complete mesh.GossipData) {
	//TODO: need to return empty data or nil (BuzzPeer::Gossip)
	panic("not implemented yet")
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *BuzzPeer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	//TODO: need to return empty data or nil (BuzzPeer::OnGossip)
	panic("not implemented yet")
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *BuzzPeer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	var pkt schema.BuzzPacket
	if err_ := gob.NewDecoder(bytes.NewReader(buf)).Decode(&pkt); err_ != nil {
		return nil, err_
	}

	err_ := p.dataManager.mergeReceived(&pkt)
	if err_ != nil {
		panic(err_)
	}
	//if received == nil {
	//	p.logger.Printf("OnGossipBroadcast %s %v => delta %v", src, set, received)
	//} else {
	//	p.logger.Printf("OnGossipBroadcast %s %v => delta %v", src, set, received.(*state).set)
	//}
	return received, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *BuzzPeer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	var set map[mesh.PeerName]int
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return err
	}

	//TODO: need to implement (BuzzPeer::OnGossipUnicast)
	panic("not implemented yet")
}
