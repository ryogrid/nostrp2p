package core

import (
	"fmt"
	"github.com/ryogrid/buzzoon/schema"
)

type DataManager struct {
	SelfPubkey [32]byte
	// TODO: need to implement (DataManager)
}

func (dman *DataManager) storeReceived(pkt *schema.BuzzPacket) error {
	// TODO: need to implement (DataManager::mergeReceived)
	if pkt.Events != nil {
		for _, evt := range pkt.Events {
			if evt.Pubkey != dman.SelfPubkey {
				fmt.Println(evt.Tags["nickname"][0] + "> " + evt.Content)
			}
		}
	} else {
		fmt.Println("pkt.Events is nil")
	}
	return nil
}
