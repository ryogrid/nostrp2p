package core

import (
	"context"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"time"
)

// EventResender is a struct that manages the resending of events
// and the removal of store amount limit overed events
type EventResender struct {
	dman   DataManager
	msgMan *MessageManager
	cancel *context.CancelFunc
}

func NewEventResender(dman DataManager, msgMan *MessageManager) *EventResender {
	return &EventResender{dman: dman, msgMan: msgMan}
}

func (er *EventResender) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	er.cancel = &cancel
	go er.ResendEvents(ctx, np2p_const.ResendCcheckInterval)
}

func (er *EventResender) Stop() {
	(*er.cancel)()
}

func (er *EventResender) ResendEvents(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// resend events
			np2p_util.Np2pDbgPrintln("ResendEvents: start")
			itr := er.dman.GetReSendNeededEventItr()
			if itr == nil {
				return
			}
			unixtime := time.Now().Unix()
			for itr.Next() {
				val := itr.Value()
				if val == nil {
					fmt.Println("on re-send iterating, strange nil value found")
					continue
				}
				resendEvt := val.(*schema.ResendEvent)
				elapsedMin := (unixtime - resendEvt.CreatedAt) / 60
				if evt, ok := er.dman.GetEventById(resendEvt.EvtId); ok {
					for n := 1; n <= np2p_const.ResendMaxTimes; n++ {
						diff := elapsedMin - int64(np2p_const.ResendTimeBaseMin*2^n)
						// if elapsed min is match with resend time, resend
						if diff == 0 {
							for _, destId := range resendEvt.DestIds {
								// resend
								evtArr := []*schema.Np2pEvent{evt}
								err := er.msgMan.SendMsgUnicast(destId, schema.NewNp2pPacket(&evtArr, nil))
								if err == nil {
									// remove from resend needed list
									er.dman.RemoveReSendNeededEvent(resendEvt, evt)
								}
							}
							break
						}
					}
				}
				// if elapsed min is over resend max time, remove from resend needed list
				if elapsedMin > int64(np2p_const.ResendTimeBaseMin*(2^np2p_const.ResendMaxTimes)) {
					er.dman.RemoveReSendNeededEvent(resendEvt, nil)
				}
			}

			// remove store amount limit overed events from DB
			er.dman.RemoveStoreAmountLimitOveredEvents()
		}
	}
}
