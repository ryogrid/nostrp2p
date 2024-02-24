package core

import (
	"fmt"
	"github.com/ryogrid/buzzoon/schema"
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

	rm.messageMan.DataMan.EvtLogger.IsLoggingActive = false
	// do recovery
	for _, buf, err := rm.messageMan.DataMan.EvtLogger.ReadLog(); err == nil; {
		evt, err_ := schema.NewBuzzEventFromBytes(buf)
		if err_ != nil {
			// EOF
			break
		}
		pkt := schema.NewBuzzPacket(&[]*schema.BuzzEvent{evt}, nil)
		rm.messageMan.handleRecvMsgBcast(pkt)
		fmt.Println("Recovered events...")
	}
	rm.messageMan.DataMan.EvtLogger.IsLoggingActive = true
}
