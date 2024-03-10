package api_server

import (
	"encoding/binary"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ryogrid/nostrp2p/core"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/schema"
	"log"
	"math"
	"net/http"
	"slices"
	"strconv"
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

type Np2pEventAndReq struct {
	Id         string     // string of ID (32bytes) in hex
	Pubkey     string     // string of Pubkey(encoded 256bit uint (holiman/uint256)) in hex
	Created_at int64      // unix timestamp in seconds
	Kind       uint16     // integer between 0 and 65535
	Tags       [][]string // Key: tag string, Value: string
	Content    string
	Sig        string // string of Sig(64-bytes integr of the signature) in hex
}

func NewNp2pEventAndReq(evt *schema.Np2pEvent) *Np2pEventAndReq {
	idBuf := make([]byte, 32)
	binary.LittleEndian.PutUint64(idBuf, evt.Id)
	idStr := fmt.Sprintf("%x", np2p_util.Gen256bitHash(idBuf))
	sigStr := idStr + idStr

	tagsArr := make([][]string, 0)
	if evt.Kind == core.KIND_EVT_PROFILE {
		tagsArr = append(tagsArr, []string{"name", evt.Tags["name"][0].(string)})
		tagsArr = append(tagsArr, []string{"about", evt.Tags["about"][0].(string)})
		tagsArr = append(tagsArr, []string{"picture", evt.Tags["picture"][0].(string)})
	}
	return &Np2pEventAndReq{
		Id:         idStr, // remove leading zeros
		Pubkey:     fmt.Sprintf("%x", evt.Pubkey[:]),
		Created_at: evt.Created_at,
		Kind:       evt.Kind,
		Tags:       tagsArr,
		Content:    evt.Content,
		Sig:        sigStr,
	}
}

type EventsResp struct {
	Events []Np2pEventAndReq
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

func (s *ApiServer) sendEventHandler(w rest.ResponseWriter, req *rest.Request) {
	input := Np2pEventAndReq{}
	err := req.DecodeJsonPayload(&input)

	if np2p_util.DenyWriteMode {
		rest.Error(w, "Write is denied", http.StatusNotAcceptable)
		return
	}

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: need to check Sig (ApiServer::sendEventHandler)

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

func (s *ApiServer) sendPost(w rest.ResponseWriter, input *Np2pEventAndReq) {
	// TODO: need to implement post handling (ApiServer::sendPost)

	if input.Content == "" {
		rest.Error(w, "Content is required", 400)
		return
	}

	evt := s.buzzPeer.MessageMan.BcastOwnPost(input.Content)
	// store for myself
	s.buzzPeer.MessageMan.DataMan.StoreEvent(evt)
	// display for myself
	s.buzzPeer.MessageMan.DispPostAtStdout(evt)

	w.WriteJson(&EventsResp{})
}

func (s *ApiServer) getProfile(w rest.ResponseWriter, input *Np2pEventAndReq) {
	// TODO: need to implement profile request handling (ApiServer::getProfile)

	prof := s.buzzPeer.MessageMan.DataMan.GetProfileLocal(math.MaxUint64)
	//prof := s.buzzPeer.MessageMan.DataMan.GetProfileLocal(input.ShortPkey)
	// TODO: when profile is not found, request latest profile (ApiServer::getProfile)

	if prof != nil {
		// TODO: need to set approprivate event data (ApiServer::getProfile)
		w.WriteJson(&EventsResp{Events: []Np2pEventAndReq{*NewNp2pEventAndReq(nil)}})
	} else {
		// profile data will be included on response of "getEvents"
		w.WriteJson(&EventsResp{Events: []Np2pEventAndReq{}})
	}
}

func (s *ApiServer) reqHandler(w rest.ResponseWriter, req *rest.Request) {
	input := Np2pEventAndReq{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: need to check Created_at and Sig for authorizaton (ApiServer::reqHandler)
	//       accept only when ((currentTime - Created_at) < 10sec)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	switch input.Content {
	case "getProfile":
		s.getProfile(w, &input)
	case "getEvents":
		s.getEvents(w, &input)
	default:
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *ApiServer) getEvents(w rest.ResponseWriter, input *Np2pEventAndReq) {
	if input.Tags == nil {
		rest.Error(w, "Tags is null", http.StatusBadRequest)
		return
	}

	sinceIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "since" })
	untilIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "until" })
	if sinceIdx == -1 || untilIdx == -1 || len(input.Tags[sinceIdx]) < 2 || len(input.Tags[untilIdx]) < 2 {
		rest.Error(w, "since and until are required", http.StatusBadRequest)
		return
	}

	since, err1 := strconv.Atoi(input.Tags[sinceIdx][1])
	until, err2 := strconv.Atoi(input.Tags[untilIdx][1])
	if err1 != nil || err2 != nil {
		rest.Error(w, "since and until must be integer", http.StatusBadRequest)
		return
	}

	events := s.buzzPeer.MessageMan.DataMan.GetLatestEvents(int64(since), int64(until))

	retEvents := make([]Np2pEventAndReq, 0)
	for _, evt := range *events {
		retEvents = append(retEvents, *NewNp2pEventAndReq(evt))
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

func (s *ApiServer) updateProfile(w rest.ResponseWriter, input *Np2pEventAndReq) {
	// TODO: need to implement profile update handling (ApiServer::updateProfile)
	//if input.Name == "" {
	//	rest.Error(w, "Name is required", 400)
	//	return
	//}
	//
	//prof := s.buzzPeer.MessageMan.BcastOwnProfile(&input.Name, &input.About, &input.Picture)
	//// update local profile
	//glo_val.ProfileMyOwn = prof
	//
	//w.WriteJson(&GeneralResp{
	//	"SUCCESS",
	//})

	if input.Tags == nil {
		rest.Error(w, "Tags is null", http.StatusBadRequest)
		return
	}

	nameIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "name" })
	aboutIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "about" })
	pictureIdx := slices.IndexFunc(input.Tags, func(ss []string) bool { return ss[0] == "picture" })
	if nameIdx == -1 || aboutIdx == -1 || pictureIdx == -1 || len(input.Tags[nameIdx]) < 2 || len(input.Tags[aboutIdx]) < 2 || len(input.Tags[pictureIdx]) < 2 {
		rest.Error(w, "since and until are required", http.StatusBadRequest)
		return
	}

	name := input.Tags[nameIdx][1]
	about := input.Tags[aboutIdx][1]
	picture := input.Tags[pictureIdx][1]

	prof := s.buzzPeer.MessageMan.BcastOwnProfile(&name, &about, &picture)
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
		&rest.Route{"POST", "/sendEvent", s.sendEventHandler},
		//&rest.Route{"POST", "/updateProfile", s.updateProfile},
		//&rest.Route{"POST", "/getProfile", s.getProfile},
		//&rest.Route{"POST", "/gatherData", s.gatherData},
		//&rest.Route{"POST", "/getEvents", s.getEvents},
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
