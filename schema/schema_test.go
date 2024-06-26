package schema

import (
	"crypto/sha256"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/vmihailenco/msgpack/v5"
	"strconv"
	"testing"
	"time"
)

func TestNp2pPacketEncodeMessagePack(t *testing.T) {
	tagMap := make([][]TagElem, 0)
	tagMap = append(tagMap, []TagElem{[]byte("nickname"), []byte("ryogrid")})
	tagMap = append(tagMap, []TagElem{[]byte("u"), []byte(strconv.FormatInt(time.Now().Unix(), 10))})

	np2p_util.InitializeRandGen(int64(777))

	event := &Np2pEvent{
		Id:         [np2p_const.EventIdSize]byte{},
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagMap,
		Content:    "こんにちは世界",
		Sig:        &[np2p_const.SignatureSize]byte{},
	}

	// set value to SelfPubkey and Sig field
	hf := sha256.New()
	hf.Write([]byte("test"))
	pubkey := hf.Sum(nil)[:np2p_const.PubkeySize]
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
	pkt := Np2pPacket{
		Events: []*Np2pEvent{event},
		Reqs:   []*Np2pReq{NewNp2pReq(1, argsMap)},
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

	var decodedPkt Np2pPacket
	err = msgpack.Unmarshal(encodedPkt, &decodedPkt)
	if err != nil {
		panic(err)
	}
	//decodedPkt, err := NewNp2pPacketFromBytes(encodedPkt)
	np2p_util.Assert(t, err == nil, "decode error")

	fmt.Println(*decodedPkt.Events[0])
	fmt.Println(*decodedPkt.Reqs[0])
	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"])
	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"][0])
	fmt.Println(*(*decodedPkt.Events[0]).Sig)
}
