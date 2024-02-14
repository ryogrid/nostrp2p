package core

import (
	"fmt"
	"github.com/ryogrid/buzzoon/schema"
)

type DataManager struct {
	// TODO: need to implement (DataManager)
}

func (dman *DataManager) storeReceived(pkt *schema.BuzzPacket) error {
	// TODO: need to implement (DataManager::mergeReceived)
	if pkt.Events != nil {
		fmt.Println(pkt.Events[0].Tags["nickname"][0] + "> " + pkt.Events[0].Content)
	} else {
		fmt.Println("pkt.Events is nil")
	}
	return nil
}
