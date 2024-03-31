package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ryogrid/nostrp2p/api_server"
	"github.com/ryogrid/nostrp2p/core"
	"github.com/spf13/cobra"
	"github.com/weaveworks/mesh"
)

var listenAddrPort = "127.0.0.1:20000"
var bootPeerAddrPort = ""
var publicKey = ""
var nickname = ""
var writable = true
var debug = false
var isEnabledSSL = false

var rootCmd = &cobra.Command{
	Use: "nostrp2p",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nostrp2p v0.0.1")
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Startup server.",
	Run: func(cmd *cobra.Command, args []string) {
		if !writable {
			glo_val.DenyWriteMode = true
		}
		if debug {
			np2p_util.DebugMode = true
		}

		peers := &np2p_util.Stringset{}
		if bootPeerAddrPort != "" {
			peers.Set(bootPeerAddrPort)
		}

		logger := log.New(os.Stderr, nickname+"> ", log.LstdFlags)

		host, portStr, err := net.SplitHostPort(listenAddrPort)
		if err != nil {
			logger.Fatalf("SplitHostPort error: %s: %v", listenAddrPort, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Fatalf("port sting coversion error: %s: %v", listenAddrPort, err)
		}

		fmt.Println("public key: ", publicKey)

		// use 6 bytes only
		name := np2p_util.Get6ByteUint64FromHexPubKeyStr(publicKey)

		pubKeyBytes, err := hex.DecodeString(publicKey)
		if err != nil {
			logger.Fatalf("public key: %s: %v", publicKey, err)
		}

		var tmpArr [32]byte
		copy(tmpArr[:], pubKeyBytes[:32])
		glo_val.SelfPubkey = &tmpArr
		glo_val.SelfPubkey64bit = np2p_util.GetUint64FromHexPubKeyStr(publicKey)
		fmt.Println(fmt.Sprintf("%x", *glo_val.SelfPubkey), fmt.Sprintf("%x", name))

		if isEnabledSSL {
			glo_val.IsEnabledSSL = true
			fmt.Println("REST I/F is offered over SSL")
		}
		// initializa rand generator
		np2p_util.InitializeRandGen(-1 * int64(name))

		router, err := mesh.NewRouter(mesh.Config{
			Host:               host,
			Port:               port,
			ProtocolMinVersion: mesh.ProtocolMaxVersion,
			Password:           nil,
			ConnLimit:          64,
			PeerDiscovery:      true,
			TrustedSubnets:     []*net.IPNet{},
		}, mesh.PeerName(name), nickname, mesh.NullOverlay{}, log.New(ioutil.Discard, "", 0))

		if err != nil {
			logger.Fatalf("Could not create router: %v", err)
		}

		//// initialized at server restart or update request
		//glo_val.Nickname = &nickname
		//glo_val.ProfileMyOwn = &schema.Np2pProfile{
		//	Pubkey64bit: name,
		//	Name:        nickname,
		//	About:       "brank yet",
		//	Picture:     "http://robohash.org/" + strconv.Itoa(int(name)) + ".png?size=200x200",
		//	UpdatedAt:   0,
		//}

		peer := core.NewPeer(mesh.PeerName(name), logger)

		// if log file exist, load it
		core.NewRecoveryManager(peer.MessageMan).Recover()
		time.Sleep(10 * time.Second)

		gossip, err := router.NewGossip("nostrp2p", peer)
		if err != nil {
			logger.Fatalf("Could not create gossip: %v", err)
		}

		peer.Register(gossip)

		go func() {
			logger.Printf("mesh router starting (%s)", listenAddrPort)
			router.Start()
		}()
		defer func() {
			logger.Printf("mesh router stopping")
			router.Stop()
		}()

		router.ConnectionMaker.InitiateConnections(peers.Slice(), true)
		peer.Router = router

		if !glo_val.DenyWriteMode {
			apiServ := api_server.NewApiServer(peer)
			go apiServ.LaunchAPIServer(host + ":" + strconv.Itoa(port+1))
		}

		np2p_util.OSInterrupt()
	},
}

// TODO: need to implement temporal post requesting to server (cmd.go)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()

	serverCmd.Flags().StringVarP(
		&listenAddrPort,
		"listen-addr-port",
		"l",
		"127.0.0.1:20000",
		"Address and port to bind to (optional)",
	)
	serverCmd.Flags().StringVarP(
		&bootPeerAddrPort,
		"boot-peer-addr-port",
		"b",
		"",
		"Address and port of a server which already joined nostrp2p network (optional)",
	)
	serverCmd.Flags().StringVarP(
		&publicKey,
		"public-key",
		"p",
		"",
		"Your public key (required)",
	)
	serverCmd.MarkFlagRequired("public-key")

	serverCmd.Flags().StringVarP(
		&nickname,
		"Your nickname on nostrp2p (required)",
		"n",
		"",
		"Port to forward",
	)
	serverCmd.MarkFlagRequired("nickname")

	serverCmd.Flags().BoolVarP(
		&writable,
		"writable",
		"w",
		true,
		"Whether handle write request (default: true)",
	)
	serverCmd.Flags().BoolVarP(
		&debug,
		"debug",
		"d",
		false,
		"If true, debug log is output to stderr (default: false)",
	)
	serverCmd.Flags().BoolVarP(
		&isEnabledSSL,
		"ssl",
		"s",
		false,
		"If true, REST I/F is offered over SSL (default: false)",
	)

	rootCmd.AddCommand(serverCmd)
}
