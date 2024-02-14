package api_server

import (
	"encoding/binary"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ryogrid/buzzoon/core"
	"github.com/ryogrid/buzzoon/schema"
	"log"
	"net/http"
	"time"
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

	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, uint64(s.buzzPeer.SelfId))
	var pubkeyBytes [32]byte
	copy(pubkeyBytes[:], buf)
	var sigBytes [64]byte
	copy(sigBytes[:], buf)
	tagsMap := make(map[string][]string)
	tagsMap["nickname"] = []string{*s.buzzPeer.Nickname}
	event := schema.BuzzEvent{
		Id:         0,
		Pubkey:     pubkeyBytes,
		Created_at: time.Now().Unix(),
		Kind:       1,
		Tags:       tagsMap,
		Content:    input.Content,
		Sig:        sigBytes,
	}
	events := []*schema.BuzzEvent{&event}
	s.buzzPeer.MessageMan.SendMsgBroadcast(&schema.BuzzPacket{events, nil, nil})

	w.WriteJson(&PostEventResp{
		"SUCCESS",
	})
}

func (s *ApiServer) LaunchAPIServer(addrStr string) {
	api := rest.NewApi()

	// the Middleware stack
	api.Use(rest.DefaultDevStack...)
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
