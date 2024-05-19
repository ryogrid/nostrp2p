package schema

import (
	"encoding/hex"
	"github.com/nbd-wtf/go-nostr"
	"github.com/ryogrid/nostrp2p/np2p_util"
)

type Np2pEventForREST struct {
	Id         string     `json:"id"`         // string of ID (32bytes) in hex
	Pubkey     string     `json:"pubkey"`     // string of Pubkey(encoded 256bit uint (holiman/uint256)) in hex
	Created_at int64      `json:"created_at"` // unix timestamp in seconds
	Kind       uint16     `json:"kind"`       // integer between 0 and 65535
	Tags       [][]string `json:"tags"`       // Key: tag string, Value: string
	Content    string     `json:"content"`
	Sig        string     `json:"sig"` // string of Sig(64-bytes integr of the signature) in hex
}

func NewNp2pEventForREST(evt *Np2pEvent) *Np2pEventForREST {
	idStr := hex.EncodeToString(evt.Id[:])
	pubkeyStr := hex.EncodeToString(evt.Pubkey[:])
	sigStr := ""
	if evt.Sig != nil {
		sigStr = hex.EncodeToString(evt.Sig[:])
	}

	tagsArr := make([][]string, 0)
	for _, v := range evt.Tags {
		tmpArr := make([]string, 0)
		for _, vv := range v {
			tmpArr = append(tmpArr, string(vv))
		}
		tagsArr = append(tagsArr, tmpArr)
	}

	return &Np2pEventForREST{
		Id:         idStr,     // remove leading zeros
		Pubkey:     pubkeyStr, //fmt.Sprintf("%x", evt.Pubkey[:]),
		Created_at: int64(evt.Created_at),
		Kind:       evt.Kind,
		Tags:       tagsArr,
		Content:    evt.Content,
		Sig:        sigStr,
	}
}

func NewNp2pEventFromREST(evt *Np2pEventForREST) *Np2pEvent {
	tagsList := make([][]TagElem, 0)
	if evt.Tags != nil {
		for _, tagList := range evt.Tags {
			elems := make([]TagElem, 0)
			for _, elem := range tagList {
				elems = append(elems, TagElem(elem))
			}
			tagsList = append(tagsList, elems)
		}
	}

	pkey, err := hex.DecodeString(evt.Pubkey)
	if err != nil {
		panic(err)
	}
	pkey32 := [32]byte{}
	copy(pkey32[:], pkey)
	evtId, err := hex.DecodeString(evt.Id)
	if err != nil {
		panic(err)
	}
	evtId32 := [32]byte{}
	copy(evtId32[:], evtId)

	allBytes, err := hex.DecodeString(evt.Sig)
	if err != nil {
		panic(err)
	}

	var sigBytes [64]byte
	copy(sigBytes[:], allBytes)

	retEvt := &Np2pEvent{
		Pubkey:     pkey32,  //pkey.Bytes32(),
		Id:         evtId32, //evtId.Bytes32(),
		Created_at: float64(evt.Created_at),
		Kind:       evt.Kind,
		Tags:       tagsList,
		Content:    evt.Content,
		Sig:        &sigBytes,
	}

	return retEvt
}

func (evt *Np2pEventForREST) Verify() bool {
	libFormEvt := nostr.Event{ID: evt.Id, PubKey: evt.Pubkey, Kind: int(evt.Kind), Content: evt.Content, CreatedAt: nostr.Timestamp(evt.Created_at), Tags: np2p_util.ConvStringArrToTagArr(evt.Tags), Sig: evt.Sig}

	ok, _ := libFormEvt.CheckSignature()
	return ok
}
