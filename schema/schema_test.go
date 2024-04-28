package schema

import (
	"crypto/sha256"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
	"strconv"
	"testing"
	"time"
)

//func TestNp2pPacketEncodeMessagePack(t *testing.T) {
//	tagList := make([][]TagElem, 0)
//	nickNameTag := make([]TagElem, 0)
//	nickNameTag = append(nickNameTag, TagElem("nickname"))
//	nickNameTag = append(nickNameTag, TagElem("ryogrid"))
//	tagList = append(tagList, nickNameTag)
//	uTag := make([]TagElem, 0)
//	uTag = append(uTag, TagElem("u"))
//	uTag = append(uTag, TagElem(strconv.FormatInt(time.Now().Unix(), 10)))
//	tagList = append(tagList, uTag)
//
//	np2p_util.InitializeRandGen(int64(777))
//
//	event := &Np2pEvent{
//		Id:         [np2p_const.EventIdSize]byte{},
//		Created_at: time.Now().Unix(),
//		Kind:       1,
//		Tags:       tagList,
//		Content:    "こんにちは世界",
//		Sig:        &[np2p_const.SignatureSize]byte{},
//	}
//
//	// set value to SelfPubkey and Sig field
//	hf := sha256.New()
//	hf.Write([]byte("test"))
//	pubkey := hf.Sum(nil)[:np2p_const.PubkeySize]
//	for idx, val := range pubkey {
//		event.Pubkey[idx] = val
//	}
//
//	hf2 := sha256.New()
//	hf2.Write([]byte("test22222"))
//	sig := hf2.Sum(nil)[:32]
//	for idx, val := range sig {
//		event.Sig[idx] = val
//	}
//	hf3 := sha256.New()
//	hf3.Write([]byte("test33333"))
//	sig2 := hf3.Sum(nil)[:32]
//	for idx, val := range sig2 {
//		event.Sig[32+idx] = val
//	}
//
//	argsMap := make(map[string][]byte)
//	//tmpStr := "ryogrid"
//	argsMap["hoge"] = []byte{250}
//	pkt := Np2pPacket{
//		Events: []*Np2pEvent{event},
//		Reqs:   []*Np2pReq{NewNp2pReq(1, argsMap)},
//	}
//
//	//encodedPkt := pkt.Encode()[0]
//	encodedPkt, err := msgpack.Marshal(pkt)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("marshaled size:" + strconv.Itoa(len(encodedPkt)))
//
//	//fmt.Println("no compressed:" + strconv.Itoa(len(encodedPkt)))
//	//
//	//var buf bytes.Buffer
//	//zw := gzip.NewWriter(&buf)
//	//_, err := zw.Write(encodedPkt)
//	//if err != nil {
//	//	t.Error(err)
//	//}
//	//
//	//if err := zw.Close(); err != nil {
//	//	t.Error(err)
//	//}
//	//
//	//fmt.Println("compressed:" + strconv.Itoa(len(buf.Bytes())))
//
//	var decodedPkt Np2pPacket
//	err = msgpack.Unmarshal(encodedPkt, &decodedPkt)
//	if err != nil {
//		panic(err)
//	}
//	//decodedPkt, err := NewNp2pPacketFromBytes(encodedPkt)
//	np2p_util.Assert(t, err == nil, "decode error")
//
//	fmt.Println(*decodedPkt.Events[0])
//	fmt.Println(*decodedPkt.Reqs[0])
//	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"])
//	fmt.Println((*decodedPkt.Reqs[0]).Args["hoge"][0])
//	fmt.Println(*(*decodedPkt.Events[0]).Sig)
//}

func TestNp2pEventEncodeMesagePack(t *testing.T) {
	tagList := make([][]TagElem, 0)
	nickNameTag := make([]TagElem, 0)
	nickNameTag = append(nickNameTag, TagElem("nickname"))
	nickNameTag = append(nickNameTag, TagElem("ryogrid"))
	tagList = append(tagList, nickNameTag)
	uTag := make([]TagElem, 0)
	uTag = append(uTag, TagElem("u"))
	uTag = append(uTag, TagElem(strconv.FormatInt(time.Now().Unix(), 10)))
	tagList = append(tagList, uTag)

	np2p_util.InitializeRandGen(int64(777))

	event := &Np2pEvent{
		Id:         [np2p_const.EventIdSize]byte{},
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagList,
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

	encodedEvt, err := msgpack.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Println("marshaled size:" + strconv.Itoa(len(encodedEvt)))

	var decodedEvt Np2pEvent
	err = msgpack.Unmarshal(encodedEvt, &decodedEvt)
	if err != nil {
		panic(err)
	}
	np2p_util.Assert(t, err == nil, "decode error")

	fmt.Println(decodedEvt)
	fmt.Println(*decodedEvt.Sig)
	for _, tagElem := range decodedEvt.Tags {
		for _, tag := range tagElem {
			fmt.Println(string(tag))
		}
	}
}

func TestNp2pEventPBEncodeProtbuf(t *testing.T) {
	tagList := make([]*Tag, 0)
	nickNameTag := make([][]byte, 0)
	nickNameTag = append(nickNameTag, TagElem("nickname"))
	nickNameTag = append(nickNameTag, TagElem("ryogrid"))
	tagList = append(tagList, &Tag{Tag: nickNameTag})
	uTag := make([][]byte, 0)
	uTag = append(uTag, TagElem("u"))
	uTag = append(uTag, TagElem(strconv.FormatInt(time.Now().Unix(), 10)))
	tagList = append(tagList, &Tag{Tag: uTag})

	np2p_util.InitializeRandGen(int64(777))

	event := &Np2PEventPB{
		Id:        make([]byte, np2p_const.EventIdSize),
		Pubkey:    make([]byte, np2p_const.PubkeySize),
		CreatedAt: uint64(time.Now().Unix()),
		Kind:      1,
		Tags:      tagList,
		Content:   "こんにちは世界",
		Sig:       make([]byte, np2p_const.SignatureSize),
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

	encodedData, err := proto.Marshal(event)
	if err != nil {
		panic(err)
	}

	fmt.Println("marshaled size:" + strconv.Itoa(len(encodedData)))

	decodedData := new(Np2PEventPB)
	err = proto.Unmarshal(encodedData, decodedData)
	if err != nil {
		panic(err)
	}
	np2p_util.Assert(t, err == nil, "decode error")

	fmt.Println(*decodedData)
	fmt.Println(decodedData.Sig)
	for _, tagElem := range decodedData.Tags {
		for _, tag := range tagElem.Tag {
			fmt.Println(string(tag))
		}
	}
}
