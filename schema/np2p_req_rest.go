package schema

import "encoding/json"

type Np2pReqForREST struct {
	Ids     []string `json:"ids"`
	Tag     []string `json:"tag"` // "#<single-letter (a-zA-Z)>": <a list of tag values, for #e — a list of event ids, for #p — a list of pubkeys, etc.>
	Authors []string `json:"authors"`
	Kinds   []int    `json:"kinds"` // list of kind numbers (ex: "1,2,3")
	Since   int64    `json:"since"`
	Until   int64    `json:"until"`
	Limit   int64    `json:"limit"`
}

func (p *Np2pReqForREST) UnmarshalJSON(data []byte) error {
	type Np2pReqForREST2 struct {
		Ids     []string `json:"ids"`
		Tag     []string `json:"tag"` // "#<single-letter (a-zA-Z)>": <a list of tag values, for #e — a list of event ids, for #p — a list of pubkeys, etc.>
		Authors []string `json:"authors"`
		Kinds   []int    `json:"kinds"`
		Since   int64    `json:"since"`
		Until   int64    `json:"until"`
		Limit   int64    `json:"limit"`
	}

	var req Np2pReqForREST2
	json.Unmarshal(data, &req)
	*p = *(*Np2pReqForREST)(&req)

	var tag map[string][]string
	json.Unmarshal(data, &tag)

	for k, v := range tag {
		if k[0] == '#' && len(k) == 2 {
			if v == nil {
				continue
			}
			//fmt.Println(v)
			//fmt.Println(reflect.TypeOf(v))
			p.Tag = []string{k}
			for _, val := range v {
				p.Tag = append(p.Tag, val)
			}
		}
	}

	return nil
}
