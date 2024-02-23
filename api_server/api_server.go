package api_server

import (
	"fmt"
	"github.com/ryogrid/buzzoon/buzz_util"
	"log"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ryogrid/buzzoon/core"
	"github.com/ryogrid/buzzoon/schema"
)

type PostEventReq struct {
	Content string
}

type PostEventResp struct {
	Status string
}

type ApiServer struct {
	buzzPeer *core.BuzzPeer
}

func NewApiServer(peer *core.BuzzPeer) *ApiServer {
	return &ApiServer{peer}
}

func (s *ApiServer) postEvent(w rest.ResponseWriter, req *rest.Request) {
	input := PostEventReq{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		rest.Error(w, "Content is required", 400)
		return
	}

	pubSlice := s.buzzPeer.Pubkey[:]
	var sigBytes [64]byte
	copy(sigBytes[:], pubSlice)
	tagsMap := make(map[string][]string)
	tagsMap["nickname"] = []string{*s.buzzPeer.Nickname}
	event := schema.BuzzEvent{
		Id:         buzz_util.GetRandUint64(),
		Pubkey:     s.buzzPeer.Pubkey,
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagsMap,
		Content:    input.Content,
		Sig:        sigBytes,
	}
	events := []*schema.BuzzEvent{&event}
	//for _, peerId := range s.buzzPeer.GetPeerList() {
	//	s.buzzPeer.MessageMan.SendMsgUnicast(peerId, schema.NewBuzzPacket(&events, nil, nil))
	//}
	s.buzzPeer.MessageMan.SendMsgBroadcast(schema.NewBuzzPacket(&events, nil))
	// store own issued event
	s.buzzPeer.MessageMan.DataMan.StoreEvent(&event)

	// display for myself
	s.buzzPeer.MessageMan.DataMan.DispPostAtStdout(&event)
	//fmt.Println(event.Tags["nickname"][0] + "> " + event.Content)

	w.WriteJson(&PostEventResp{
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
		AllowedMethods:                []string{"POST"},
		AllowedHeaders:                []string{"Accept", "content-type"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	router, err := rest.MakeRouter(
		&rest.Route{"POST", "/postEvent", s.postEvent},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Printf("Server started")
	log.Fatal(http.ListenAndServe(
		addrStr,
		api.MakeHandler(),
	))
}
