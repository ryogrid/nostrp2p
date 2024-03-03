package core

import (
	"fmt"
	"github.com/ryogrid/nostrp2p/schema"
	"math"
)

type RecoveryManager struct {
	messageMan *MessageManager
}

func NewRecoveryManager(messageMan *MessageManager) *RecoveryManager {
	return &RecoveryManager{messageMan: messageMan}
}

func (rm *RecoveryManager) Recover() {
	if rm.messageMan.DataMan.EvtLogger.GetLogfileSize() == 0 {
		return
	}

	fmt.Println("Recovering from log file...")
	// do recovery
	_, buf, err := rm.messageMan.DataMan.EvtLogger.ReadLog()
	for err == nil {
		evt, err_ := schema.NewNp2pEventFromBytes(buf)
		if evt.Tags != nil {
			evt.Tags["recovering"] = []interface{}{true}
		} else {
			evt.Tags = make(map[string][]interface{})
			evt.Tags["recovering"] = []interface{}{true}
		}
		if err_ != nil {
			// EOF
			break
		}
		pkt := schema.NewNp2pPacket(&[]*schema.Np2pEvent{evt}, nil)
		rm.messageMan.handleRecvMsgBcastEvt(math.MaxUint64, pkt, evt)
		_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog()
	}
}
