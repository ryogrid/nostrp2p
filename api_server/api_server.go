package api_server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ryogrid/nostrp2p/core"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
	"math"
	"net/http"
	"slices"
	"time"
)

type NoArgReq struct {
}

type Np2pEventForREST struct {
	Id         string     `json:"id"`         // string of ID (32bytes) in hex
	Pubkey     string     `json:"pubkey"`     // string of Pubkey(encoded 256bit uint (holiman/uint256)) in hex
	Created_at int64      `json:"created_at"` // unix timestamp in seconds
	Kind       uint16     `json:"kind"`       // integer between 0 and 65535
	Tags       [][]string `json:"tags"`       // Key: tag string, Value: string
	Content    string     `json:"content"`
	Sig        string     `json:"sig"` // string of Sig(64-bytes integr of the signature) in hex
}

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

func NewNp2pEventForREST(evt *schema.Np2pEvent) *Np2pEventForREST {
	//idStr := fmt.Sprintf("%x", evt.Id[:])
	idStr := hex.EncodeToString(evt.Id[:])
	pubkeyStr := hex.EncodeToString(evt.Pubkey[:])
	sigStr := ""
	if evt.Sig != nil {
		sigStr = hex.EncodeToString(evt.Sig[:])
	}

	tagsArr := make([][]string, 0)
	for k, v := range evt.Tags {
		tmpArr := make([]string, 0)
		tmpArr = append(tmpArr, k)
		for _, val := range v {
			tmpArr = append(tmpArr, val.(string))
		}
		tagsArr = append(tagsArr, tmpArr)
	}

	return &Np2pEventForREST{
		Id:         idStr,     // remove leading zeros
		Pubkey:     pubkeyStr, //fmt.Sprintf("%x", evt.Pubkey[:]),
		Created_at: evt.Created_at,
		Kind:       evt.Kind,
		Tags:       tagsArr,
		Content:    evt.Content,
		Sig:        sigStr,
	}
}

func NewNp2pEventFromREST(evt *Np2pEventForREST) *schema.Np2pEvent {
	tagsMap := make(map[string][]interface{})
	for _, tag := range evt.Tags {
		vals := make([]interface{}, 0)
		for _, val := range tag[1:] {
			vals = append(vals, val)
		}
		tagsMap[tag[0]] = vals
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

	retEvt := &schema.Np2pEvent{
		Pubkey:     pkey32,  //pkey.Bytes32(),
		Id:         evtId32, //evtId.Bytes32(),
		Created_at: evt.Created_at,
		Kind:       evt.Kind,
		Tags:       tagsMap,
		Content:    evt.Content,
		Sig:        &sigBytes,
	}

	return retEvt
}

type EventsResp struct {
	Events []Np2pEventForREST `json:"results"`
}
type GeneralResp struct {
	Status string
}

type ApiServer struct {
	buzzPeer *core.Np2pPeer
}

func NewApiServer(peer *core.Np2pPeer) *ApiServer {
	return &ApiServer{peer}
}

func (s *ApiServer) publishHandler(w rest.ResponseWriter, req *rest.Request) {
	input := Np2pEventForREST{}
	err := req.DecodeJsonPayload(&input)

	if glo_val.DenyWriteMode {
		rest.Error(w, "Write is denied", http.StatusNotAcceptable)
		return
	}

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: need to check Sig (ApiServer::publishHandler)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	switch input.Kind {
	case core.KIND_EVT_POST:
		s.sendPost(w, &input)
	case core.KIND_EVT_PROFILE:
		s.updateProfile(w, &input)
	case core.KIND_EVT_REACTION:
		s.sendReaction(w, &input)
	default:
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *ApiServer) sendReaction(w rest.ResponseWriter, input *Np2pEventForREST) {
	evt := NewNp2pEventFromREST(input)
	err := s.buzzPeer.MessageMan.UnicastEventData(evt.Tags["p"][0].(string), evt)
	if err != nil {
		fmt.Println(evt.Tags["p"][0].(string))
		fmt.Println(err)
	} else {
		// when data is sent successfully (= target server is online and received the data)
		// reaction event is stored for myself
		s.buzzPeer.MessageMan.DataMan.StoreEvent(evt)
	}

	w.WriteJson(&EventsResp{})
}

func (s *ApiServer) sendPost(w rest.ResponseWriter, input *Np2pEventForREST) {
	if input.Content == "" {
		rest.Error(w, "Content is required", 400)
		return
	}

	evt := NewNp2pEventFromREST(input)
	s.buzzPeer.MessageMan.BcastOwnPost(evt)
	// store for myself
	s.buzzPeer.MessageMan.DataMan.StoreEvent(evt)

	w.WriteJson(&EventsResp{})
}

func (s *ApiServer) reqHandler(w rest.ResponseWriter, req *rest.Request) {
	input := Np2pReqForREST{}
	err := req.DecodeJsonPayload(&input)

	//fmt.Println("reqHandler")
	//fmt.Println(input.Tag)
	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if input.Kinds == nil || len(input.Kinds) == 0 {
		//rest.Error(w, "Kinds is needed", http.StatusBadRequest)
		//return

		// for supporting Nostr clients
		w.WriteJson(&EventsResp{
			Events: []Np2pEventForREST{},
		})
		return
	}

	// TODO: need to check Created_at and Sig for authorizaton (ApiServer::reqHandler)
	//       accept only when ((currentTime - Created_at) < 10sec)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	// TODO: need to implement each kind and other fliter condition request handling (ApiServer::reqHandler)
	if slices.Contains(input.Kinds, core.KIND_REQ_SHARE_EVT_DATA) || slices.Contains(input.Kinds, core.KIND_REQ_POST) {
		s.getEvents(w, &input)
	} else if slices.Contains(input.Kinds, core.KIND_REQ_PROFILE) {
		s.getProfile(w, &input)
	} else {
		w.WriteJson(&EventsResp{
			Events: []Np2pEventForREST{},
		})
		return
	}
}

func (s *ApiServer) getProfile(w rest.ResponseWriter, input *Np2pReqForREST) {
	// TODO: need to implement profile request handling (ApiServer::getProfile)

	prof := s.buzzPeer.MessageMan.DataMan.GetProfileLocal(math.MaxUint64)
	//prof := s.buzzPeer.MessageMan.DataMan.GetProfileLocal(input.ShortPkey)
	// TODO: when profile is not found, request latest profile (ApiServer::getProfile)

	if prof != nil {
		// TODO: need to set approprivate event data (ApiServer::getProfile)
		w.WriteJson(&EventsResp{Events: []Np2pEventForREST{}})
	} else {
		// profile data will be included on response of "getEvents"
		w.WriteJson(&EventsResp{Events: []Np2pEventForREST{}})
	}
}

func (s *ApiServer) getEvents(w rest.ResponseWriter, input *Np2pReqForREST) {
	//if input.Since == 0 || input.Until == 0 {
	//	rest.Error(w, "value of since and untile is invalid", http.StatusBadRequest)
	//	return
	//}

	// for supporting Nostr clients
	isPeriodSpecified := true
	if input.Since == 0 {
		dt := time.Now()
		curUnix := dt.Unix()
		input.Since = curUnix - 60*60*24*7 // 1week
		isPeriodSpecified = false
	}
	if input.Until == 0 {
		input.Until = math.MaxInt64
	}

	events := s.buzzPeer.MessageMan.DataMan.GetLatestEvents(input.Since, input.Until)

	// for supporting Nostr clients
	// limit 50
	if !isPeriodSpecified && len(*events) > 50 {
		*events = (*events)[len(*events)-50:]
	}

	retEvents := make([]Np2pEventForREST, 0)

	for _, evt := range *events {
		retEvents = append(retEvents, *NewNp2pEventForREST(evt))
	}

	w.WriteJson(&EventsResp{
		Events: retEvents,
	})
}

// TODO: TEMPORAL IMPL
func (s *ApiServer) gatherData(w rest.ResponseWriter, req *rest.Request) {
	input := NoArgReq{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.buzzPeer.MessageMan.BcastShareEvtDataReq()

	w.WriteJson(&GeneralResp{
		"SUCCESS",
	})
}

func (s *ApiServer) updateProfile(w rest.ResponseWriter, input *Np2pEventForREST) {
	if input.Tags == nil {
		rest.Error(w, "Tags is null", http.StatusBadRequest)
		return
	}

	evt := NewNp2pEventFromREST(input)
	prof := s.buzzPeer.MessageMan.BcastOwnProfile(evt)
	// update local profile
	glo_val.ProfileMyOwn = prof

	w.WriteJson(&GeneralResp{
		"SUCCESS",
	})
}

func (s *ApiServer) LaunchAPIServer(addrStr string) {
	api := rest.NewApi()

	// the Middleware stack
	//api.Use(rest.DefaultDevStack...)
	api.Use(
		//&rest.AccessLogApacheMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
		},
		&rest.JsonIndentMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	)
	api.Use(&rest.JsonpMiddleware{
		CallbackNameKey: "cb",
	})
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return true
		},
		AllowedMethods:                []string{"POST", "OPTIONS"},
		AllowedHeaders:                []string{"Accept", "content-type", "Access-Control-Request-Headers", "Access-Control-Request-Method", "Origin", "Referer", "User-Agent"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	router, err := rest.MakeRouter(
		&rest.Route{"POST", "/publish", s.publishHandler},
		&rest.Route{"POST", "/req", s.reqHandler},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Printf("Server started")
	if glo_val.IsEnabledSSL {
		log.Fatal(http.ListenAndServeTLS(
			addrStr,
			"cert.pem",
			"privkey.pem",
			api.MakeHandler(),
		))
	} else {
		log.Fatal(http.ListenAndServe(
			addrStr,
			api.MakeHandler(),
		))
	}
}
