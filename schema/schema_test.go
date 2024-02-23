package schema

import (
	"crypto/sha256"
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_util"
	"testing"
	"time"
)

func TestBuzzPacket_Encode(t *testing.T) {
	tagMap := make(map[string][]string)
	tagMap["#e"] = []string{"a", "b", "c"}
	tagMap["#p"] = []string{"aa", "bb", "cc"}

	event := &BuzzEvent{
		Id:         11111,
		Created_at: time.Now().Unix(),
		Kind:       777,
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
		event.Sig[idx] = val
		event.Sig[32+idx] = val
	}

	pkt := BuzzPacket{
		Events: []*BuzzEvent{event},
		Req:    nil,
	}

	encodedPkt := pkt.Encode()[0]

	decodedPkt, err := newBuzzPacketFromBytes(encodedPkt)
	buzz_util.Assert(t, err == nil, "decode error")

	fmt.Println(*decodedPkt.Events[0])
}
