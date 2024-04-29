package core

import (
	"encoding/binary"
	"fmt"
	"github.com/ryogrid/nostrp2p/schema"
	"math"
	"slices"
)

// this variable is set at initialization of OnMemoryDataManager
var _edlogger *EventDataLogger

type RecoveryManager struct {
	messageMan *MessageManager
}

func NewRecoveryManager(messageMan *MessageManager) *RecoveryManager {
	return &RecoveryManager{messageMan: messageMan}
}

func (rm *RecoveryManager) Recover() {
	if _edlogger.GetLogfileSize(_edlogger.eventLogFile) == 0 {
		return
	}

	fmt.Println("Recovering from log file...")
	// do recovery (event log file)
	_, buf, err := _edlogger.ReadLog(_edlogger.eventLogFile)
	for err == nil {
		evt, err_ := schema.NewNp2pEventFromBytes(buf)
		if evt.Tags != nil {
			evt.Tags = append(evt.Tags, []schema.TagElem{schema.TagElem("recovering")})
		} else {
			evt.Tags = make([][]schema.TagElem, 0)
			evt.Tags = append(evt.Tags, []schema.TagElem{schema.TagElem("recovering")})
		}
		if err_ != nil {
			// EOF
			break
		}
		pkt := schema.NewNp2pPacket(&[]*schema.Np2pEvent{evt}, nil)
		rm.messageMan.handleRecvMsgBcastEvt(math.MaxUint64, pkt, evt)
		_, buf, err = _edlogger.ReadLog(_edlogger.eventLogFile)
	}

	// do recovery (resend finished events log file)
	tmpFinishedMap := make(map[int64]struct{})
	_, buf, err = _edlogger.ReadLog(_edlogger.reSendFinishedEvtLogFile)
	for err == nil {
		if len(buf) != 8 {
			// EOF
			break
		}

		createdAtUint := binary.LittleEndian.Uint64(buf)
		createdAt := int64(createdAtUint)
		tmpFinishedMap[createdAt] = struct{}{}
		_, buf, err = _edlogger.ReadLog(_edlogger.reSendFinishedEvtLogFile)
	}

	// do recovery (resend needed events log file)
	tmpReSendNeededEvtList := make([]*schema.ResendEvent, 0)
	_, buf, err = _edlogger.ReadLog(_edlogger.reSendNeededEvtLogFile)
	for err == nil {
		resendEvt, err_ := schema.NewResendEventFromBytes(buf)
		if err_ != nil {
			// EOF
			break
		}

		if _, ok := tmpFinishedMap[resendEvt.CreatedAt]; !ok {
			// resend needed event
			tmpReSendNeededEvtList = append(tmpReSendNeededEvtList, resendEvt)
		}
		_, buf, err = _edlogger.ReadLog(_edlogger.reSendNeededEvtLogFile)
	}
	// store read data reverse order
	slices.Reverse(tmpReSendNeededEvtList)
	for ii := 0; ii < len(tmpReSendNeededEvtList); ii++ {
		resendEvt := tmpReSendNeededEvtList[ii]
		rm.messageMan.DataMan.AddReSendNeededEvent(resendEvt.DestIds, &schema.Np2pEvent{Id: resendEvt.EvtId}, false)
	}

}
