package schema

import (
	"crypto/sha256"
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_const"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/vmihailenco/msgpack/v5"
	"strconv"
	"testing"
	"time"
)

func TestBuzzPacketEncodeMessagePack(t *testing.T) {
	tagMap := make(map[string][]interface{})
	tagMap["nickname"] = []interface{}{"ryogrid"}
	tagMap["u"] = []interface{}{strconv.FormatInt(time.Now().Unix(), 10)}

	event := &BuzzEvent{
		Id:         11111,
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagMap,
		Content:    "こんにちは世界",
		Sig:        &[buzz_const.SignatureSize]byte{},
	}

	// set value to SelfPubkey and Sig field
	hf := sha256.New()
	hf.Write([]byte("test"))
	pubkey := hf.Sum(nil)[:buzz_const.PubkeySize]
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

	argsMap := make(map[string][]interface{})
	//tmpStr := "ryogrid"
	argsMap["hoge"] = []interface{}{2024}
	pkt := BuzzPacket{
		Events: []*BuzzEvent{event},
		Reqs:   []*BuzzReq{NewBuzzReq(1, argsMap)},
	}

	//encodedPkt := pkt.Encode()[0]
	encodedPkt, err := msgpack.Marshal(pkt)
	if err != nil {
		panic(err)
	}
	fmt.Println("marshaled size:" + strconv.Itoa(len(encodedPkt)))

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

	var decodedPkt BuzzPacket
	err = msgpack.Unmarshal(encodedPkt, &decodedPkt)
	if err != nil {
		panic(err)
	}
	//decodedPkt, err := NewBuzzPacketFromBytes(encodedPkt)
	buzz_util.Assert(t, err == nil, "decode error")

	fmt.Println(*decodedPkt.Events[0])
	fmt.Println(*decodedPkt.Reqs[0])
	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"])
	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"][0])
	fmt.Println(*(*decodedPkt.Events[0]).Sig)
}
