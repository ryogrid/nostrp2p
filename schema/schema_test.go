package schema

import (
	"crypto/sha256"
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_util"
	"strconv"
	"testing"
	"time"
)

func TestBuzzPacket_Encode(t *testing.T) {
	tagMap := make(map[string][]string)
	tagMap["nickname"] = []string{"ryogrid"}
	tagMap["u"] = []string{strconv.FormatInt(time.Now().Unix(), 10)}

	event := &BuzzEvent{
		Id:         11111,
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagMap,
		Content:    "こんにちは世界",
		Sig:        &[64]byte{},
	}

	// set value to SelfPubkey and Sig field
	hf := sha256.New()
	hf.Write([]byte("test"))
	pubkey := hf.Sum(nil)[:32]
	for idx, val := range pubkey {
		event.Pubkey[idx] = val
	}

	hf2 := sha256.New()
	hf2.Write([]byte("test22222"))
	sig := hf2.Sum(nil)[:32]
	for idx, val := range sig {
		event.Sig[idx] = val
	}
	hf3 := sha256.New()
	hf3.Write([]byte("test33333"))
	sig2 := hf3.Sum(nil)[:32]
	for idx, val := range sig2 {
		event.Sig[32+idx] = val
	}

	pkt := BuzzPacket{
		Events: []*BuzzEvent{event},
		Req:    nil,
	}

	encodedPkt := pkt.Encode()[0]
	//fmt.Println("no compressed:" + strconv.Itoa(len(encodedPkt)))
	//
	//var buf bytes.Buffer
	//zw := gzip.NewWriter(&buf)
	//_, err := zw.Write(encodedPkt)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//if err := zw.Close(); err != nil {
	//	t.Error(err)
	//}
	//
	//fmt.Println("compressed:" + strconv.Itoa(len(buf.Bytes())))

	decodedPkt, err := newBuzzPacketFromBytes(encodedPkt)
	buzz_util.Assert(t, err == nil, "decode error")

	fmt.Println(*decodedPkt.Events[0])
}
