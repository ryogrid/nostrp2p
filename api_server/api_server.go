package api_server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/holiman/uint256"
	"github.com/pavelkrolevets/uint512"
	"github.com/ryogrid/nostrp2p/core"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type NoArgReq struct {
}

type PostEventReq struct {
	Content string
}

type UpdateProfileReq struct {
	Name    string
	About   string
	Picture string
}

type GetProfileReq struct {
	ShortPkey uint64
}

type GetProfileResp struct {
	Name    string
	About   string
	Picture string
}

type GetEventsReq struct {
	Since int64
	Until int64
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
	Kinds   string   `json:"kinds"` // list of kind numbers (ex: "1,2,3")
	Since   int64    `json:"since"`
	Until   int64    `json:"until"`
	Limit   int64    `json:"limit"`
}

func (p *Np2pReqForREST) UnmarshalJSON(data []byte) error {
	type Np2pReqForREST2 struct {
		Ids     []string `json:"ids"`
		Tag     []string `json:"tag"` // "#<single-letter (a-zA-Z)>": <a list of tag values, for #e — a list of event ids, for #p — a list of pubkeys, etc.>
		Authors []string `json:"authors"`
		Kinds   string   `json:"kinds"`
		Since   int64    `json:"since"`
		Until   int64    `json:"until"`
		Limit   int64    `json:"limit"`
	}

	var req Np2pReqForREST2
	json.Unmarshal(data, &req)
	*p = *(*Np2pReqForREST)(&req)

	var tag map[string]interface{}
	json.Unmarshal(data, &tag)

	for k, v := range tag {
		if k[0] == '#' && len(k) == 2 {
			if v == nil {
				continue
			}
			p.Tag = []string{string(k[:2])}
			//for _, val := range v.([]string) {
			//	p.Tag = append(p.Tag, val)
			//}
			p.Tag = append(p.Tag, v.(string))
		}
	}

	return nil
}

func NewNp2pEventForREST(evt *schema.Np2pEvent) *Np2pEventForREST {
	//idBuf := make([]byte, 32)
	//binary.LittleEndian.PutUint64(idBuf, evt.Id)
	idStr := fmt.Sprintf("%x", evt.Id[:])
	sigStr := ""
	if evt.Sig != nil {
		sigStr = hex.EncodeToString(evt.Sig[:])
	}

	tagsArr := make([][]string, 0)
	//if evt.Kind == core.KIND_EVT_PROFILE {
	//	tagsArr = append(tagsArr, []string{"name", evt.Tags["name"][0].(string)})
	//	tagsArr = append(tagsArr, []string{"about", evt.Tags["about"][0].(string)})
	//	tagsArr = append(tagsArr, []string{"picture", evt.Tags["picture"][0].(string)})
	//}
	return &Np2pEventForREST{
		Id:         idStr, // remove leading zeros
		Pubkey:     fmt.Sprintf("%x", evt.Pubkey[:]),
		Created_at: evt.Created_at,
		Kind:       evt.Kind,
		Tags:       tagsArr,
		Content:    evt.Content,
		Sig:        sigStr,
	}
}

func NewNp2pEventFromREST(evt *Np2pEventForREST) *schema.Np2pEvent {
	tagsMap := make(map[string][]interface{})
	//if evt.Kind == core.KIND_EVT_PROFILE {
	//	tagsMap["name"] = []interface{}{evt.Tags[0][1]}
	//	tagsMap["about"] = []interface{}{evt.Tags[1][1]}
	//	tagsMap["picture"] = []interface{}{evt.Tags[2][1]}
	//}

	pkey, err := uint256.FromHex("0x" + strings.TrimLeft(evt.Pubkey, "0"))
	if err != nil {
		panic(err)
	}
	evtId, err := uint256.FromHex("0x" + strings.TrimLeft(evt.Id, "0"))
	if err != nil {
		panic(err)
	}

	fmt.Println("evt.Sig: " + evt.Sig)

	lowerSig, err := uint512.FromHex("0x" + strings.TrimLeft(evt.Sig[:64], "0"))
	if err != nil {
		panic(err)
	}
	upperSig, err := uint512.FromHex("0x" + strings.TrimLeft(evt.Sig[64:128], "0"))
	if err != nil {
		panic(err)
	}
	lowerBytes := lowerSig.Bytes()
	upperBytes := upperSig.Bytes()
	fmt.Println(lowerBytes)
	fmt.Println(upperBytes)
	allBytes := make([]byte, 0)
	allBytes = append(allBytes, lowerBytes...)
	allBytes = append(allBytes, upperBytes...)

	var sigBytes [64]byte
	copy(sigBytes[:], allBytes)

	retEvt := &schema.Np2pEvent{
		Pubkey:     pkey.Bytes32(),
		Id:         evtId.Bytes32(),
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
	default:
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//w.WriteJson(&GeneralResp{
	//	"SUCCESS",
	//})
}

func (s *ApiServer) sendPost(w rest.ResponseWriter, input *Np2pEventForREST) {
	// TODO: need to implement post handling (ApiServer::sendPost)

	if input.Content == "" {
		rest.Error(w, "Content is required", 400)
		return
	}

	evt := NewNp2pEventFromREST(input)
	s.buzzPeer.MessageMan.BcastOwnPost(evt)
	// store for myself
	s.buzzPeer.MessageMan.DataMan.StoreEvent(evt)
	//// display for myself
	//s.buzzPeer.MessageMan.DispPostAtStdout(evt)

	w.WriteJson(&EventsResp{})
}

func (s *ApiServer) reqHandler(w rest.ResponseWriter, req *rest.Request) {
	input := Np2pReqForREST{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//if input.Kinds == nil || len(input.Kinds) == 0 {
	//	rest.Error(w, "Kinds is needed", http.StatusBadRequest)
	//	return
	//}
	kind, err := strconv.Atoi(strings.Split(input.Kinds, ",")[0])
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(kind)
	// TODO: need to check Created_at and Sig for authorizaton (ApiServer::reqHandler)
	//       accept only when ((currentTime - Created_at) < 10sec)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	//switch input.Kinds[0] {
	switch kind {
	case core.KIND_REQ_PROFILE:
		s.getProfile(w, &input)
	case core.KIND_REQ_SHARE_EVT_DATA:
		s.getEvents(w, &input)
	case core.KIND_REQ_POST:
		s.getEvents(w, &input)
	default:
		//rest.Error(w, "unknown kind", http.StatusBadRequest)
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
		w.WriteJson(&EventsResp{Events: []Np2pEventForREST{*NewNp2pEventForREST(nil)}})
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

	//events := s.buzzPeer.MessageMan.DataMan.GetLatestEvents(int64(input.Since), int64(input.Until))
	events := s.buzzPeer.MessageMan.DataMan.GetLatestEvents(0, math.MaxInt64)

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

	//nameIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "name" })
	//aboutIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "about" })
	//pictureIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "picture" })
	//if nameIdx == -1 || aboutIdx == -1 || pictureIdx == -1 || len(input.Tags[nameIdx]) < 2 || len(input.Tags[aboutIdx]) < 2 || len(input.Tags[pictureIdx]) < 2 {
	//	rest.Error(w, "since and until are required", http.StatusBadRequest)
	//	return
	//}

	//name := input.Tags[nameIdx][1]
	//about := input.Tags[aboutIdx][1]
	//picture := input.Tags[pictureIdx][1]
	//
	//prof := s.buzzPeer.MessageMan.BcastOwnProfile(&name, &about, &picture)

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
		//&rest.Route{"POST", "/updateProfile", s.updateProfile},
		//&rest.Route{"POST", "/getProfile", s.getProfile},
		//&rest.Route{"POST", "/gatherData", s.gatherData},
		//&rest.Route{"POST", "/getEvents", s.getEvents},
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
