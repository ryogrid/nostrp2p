package schema

type BuzzReq struct {
	Kind uint16
	Args map[string][]interface{}
}

func NewBuzzReq(kind uint16, args map[string][]interface{}) *BuzzReq {
	return &BuzzReq{
		Kind: kind,
		Args: args,
	}
}
