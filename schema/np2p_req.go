package schema

import "github.com/ryogrid/nostrp2p/np2p_util"

type Np2pReq struct {
	Id   uint64 // for avoiding duplicated receiving at broadcast
	Kind uint16
	Args map[string][]interface{}
}

func NewNp2pReq(kind uint16, args map[string][]interface{}) *Np2pReq {
	return &Np2pReq{
		Id:   np2p_util.GetRandUint64(),
		Kind: kind,
		Args: args,
	}
}
