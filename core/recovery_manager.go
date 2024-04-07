package core

import (
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/schema"
	"math"
	"slices"
)

type RecoveryManager struct {
	messageMan *MessageManager
}

func NewRecoveryManager(messageMan *MessageManager) *RecoveryManager {
	return &RecoveryManager{messageMan: messageMan}
}

func (rm *RecoveryManager) Recover() {
	if rm.messageMan.DataMan.EvtLogger.GetLogfileSize(rm.messageMan.DataMan.EvtLogger.eventLogFile) == 0 {
		return
	}

	fmt.Println("Recovering from log file...")
	// do recovery (event log file)
	_, buf, err := rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.eventLogFile)
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
		_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.eventLogFile)
	}

	// do recovery (resend finished events log file)
	tmpFinishedMap := make(map[[np2p_const.EventIdSize]byte]struct{})
	_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.reSendFinishedEvtLogFile)
	for err == nil {
		if len(buf) != np2p_const.EventIdSize {
			// EOF
			break
		}

		var evtId [np2p_const.EventIdSize]byte
		copy(evtId[:], buf)
		tmpFinishedMap[evtId] = struct{}{}
		_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.eventLogFile)
	}

	// do recovery (resend needed events log file)
	tmpReSendNeededEvtList := make([][np2p_const.EventIdSize]byte, 0)
	_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.reSendNeededEvtLogFile)
	for err == nil {
		if len(buf) != np2p_const.EventIdSize {
			// EOF
			break
		}

		var evtId [np2p_const.EventIdSize]byte
		copy(evtId[:], buf)

		if _, ok := tmpFinishedMap[evtId]; !ok {
			// resend needed event
			tmpReSendNeededEvtList = append(tmpReSendNeededEvtList, evtId)
		}
		_, buf, err = rm.messageMan.DataMan.EvtLogger.ReadLog(rm.messageMan.DataMan.EvtLogger.reSendNeededEvtLogFile)
	}
	// store read data reverse order
	slices.Reverse(tmpReSendNeededEvtList)
	for ii := 0; ii < len(tmpReSendNeededEvtList); ii++ {
		evtId := tmpReSendNeededEvtList[ii]
		rm.messageMan.DataMan.AddReSendNeededEvent(&schema.Np2pEvent{Id: evtId}, false)

	}

}
