package schema

import "github.com/ryogrid/buzzoon/buzz_util"

type BuzzReq struct {
	Id   uint64 // for avoiding duplicated receiving at broadcast
	Kind uint16
	Args map[string][]interface{}
}

func NewBuzzReq(kind uint16, args map[string][]interface{}) *BuzzReq {
	return &BuzzReq{
		Id:   buzz_util.GetRandUint64(),
		Kind: kind,
		Args: args,
	}
}
