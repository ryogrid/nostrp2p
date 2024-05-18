package core

import (
	"fmt"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"strconv"
	"testing"
)

func TestGzipCompressLateSurveyMsgPack(t *testing.T) {
	hexPubKeyStr := "09f7437e5ad50770222a9d158fb5a0e947ca4089ef4b07e8ede374d7302d8daf"
	glo_val.SelfPubkey64bit = np2p_util.GetUint64FromHexPubKeyStr(hexPubKeyStr)
	dman := NewNutsDBDataManager()
	allEvents := dman.GetLatestEvents(0, np2p_util.GetCurUnixTimeInSec(), -1)
	fmt.Println("allEvents size:" + strconv.Itoa(len(*allEvents)))
	allEventsBytes := make([]byte, 0)
	for _, evt := range *allEvents {
		allEventsBytes = append(allEventsBytes, evt.Encode()...)
	}
	fmt.Println("allEventsBytes size:" + strconv.Itoa(len(allEventsBytes)))
	compressedBytes := np2p_util.GzipCompless(allEventsBytes)
	fmt.Println("compressedBytes size:" + strconv.Itoa(len(compressedBytes)))
}
