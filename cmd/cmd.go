package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/ryogrid/nostrp2p/api_server"
	"github.com/ryogrid/nostrp2p/core"
	"github.com/ryogrid/nostrp2p/glo_val"
	"github.com/ryogrid/nostrp2p/np2p_util"
	"github.com/ryogrid/nostrp2p/transport"
	"github.com/spf13/cobra"
	"github.com/weaveworks/mesh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var listenAddrPort = "127.0.0.1:20000"
var bootPeerAddrPort = ""
var publicKey = ""
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

		logger := log.New(os.Stderr, "> ", log.LstdFlags)

		host, portStr, err := net.SplitHostPort(listenAddrPort)
		if err != nil {
			logger.Fatalf("SplitHostPort error: %s: %v", listenAddrPort, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Fatalf("port sting coversion error: %s: %v", listenAddrPort, err)
		}

		fmt.Println("public key: ", publicKey)

		var hexPubKeyStr string
		var pubKeyBytes []byte
		var err2 error
		if strings.HasPrefix(publicKey, "npub") {
			_, tmpPubKeyVal, err2_ := nip19.Decode(publicKey)
			if err2_ != nil {
				logger.Fatalf("public key: %s: %v", publicKey, err)
			}
			hexPubKeyStr = tmpPubKeyVal.(string)
			pubKeyBytes, err2 = hex.DecodeString(hexPubKeyStr)
		} else {
			hexPubKeyStr = publicKey
			pubKeyBytes, err2 = hex.DecodeString(publicKey)
			if err2 != nil {
				logger.Fatalf("public key: %s: %v", publicKey, err)
			}
		}

		var tmpArr [32]byte
		copy(tmpArr[:], pubKeyBytes[:32])
		glo_val.SelfPubkeyStr = hexPubKeyStr
		glo_val.SelfPubkey = &tmpArr
		glo_val.SelfPubkey64bit = np2p_util.GetUint64FromHexPubKeyStr(hexPubKeyStr)

		if isEnabledSSL {
			glo_val.IsEnabledSSL = true
			fmt.Println("REST I/F is offered over SSL")
		}

		// initialize rand generator
		np2p_util.InitializeRandGen(-1 * int64(glo_val.SelfPubkey64bit))

		// mesh library's peer ID is 6 bytes uint64 (MeshTransport only restriction)
		peerId := np2p_util.Get6ByteUint64FromHexPubKeyStr(hexPubKeyStr)
		peer := core.NewPeer(peerId, logger)

		fmt.Println(fmt.Sprintf("%x", *glo_val.SelfPubkey), fmt.Sprintf("%x", peerId))

		setupMeshTransport := func() *mesh.Router {
			router, err := mesh.NewRouter(mesh.Config{
				Host:               host,
				Port:               port,
				ProtocolMinVersion: mesh.ProtocolMaxVersion,
				Password:           nil,
				ConnLimit:          64,
				PeerDiscovery:      true,
				TrustedSubnets:     []*net.IPNet{},
			}, mesh.PeerName(peerId), "", mesh.NullOverlay{}, log.New(ioutil.Discard, "", 0))

			if err != nil {
				logger.Fatalf("Could not create router: %v", err)
			}

			tport := transport.NewMeshTransport(peer)
			gossip, err := router.NewGossip("nostrp2p", tport)
			tport.Register(gossip)
			peer.MessageMan.SetTransport(tport)
			if err != nil {
				logger.Fatalf("Could not create gossip: %v", err)
			}
			//peer.Register(gossip)

			go func() {
				logger.Printf("mesh router starting (%s)", listenAddrPort)
				router.Start()
			}()

			router.ConnectionMaker.InitiateConnections(peers.Slice(), true)
			tport.SetRouter(router)
			return router
		}
		router := setupMeshTransport()

		defer func() {
			logger.Printf("mesh router stopping")
			router.Stop()
		}()

		apiServ := api_server.NewApiServer(peer)
		go apiServ.LaunchAPIServer(host + ":" + strconv.Itoa(port+1))

		np2p_util.OSInterrupt()
	},
}

var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate new key pair.",
	Run: func(cmd *cobra.Command, args []string) {
		sk := nostr.GeneratePrivateKey()
		pk, _ := nostr.GetPublicKey(sk)
		nsec, _ := nip19.EncodePrivateKey(sk)
		npub, _ := nip19.EncodePublicKey(pk)

		fmt.Println("Secret Key:")
		fmt.Println(nsec)
		fmt.Println("Secret Key (In Hex Representation): ")
		fmt.Println(sk)
		fmt.Println("Public Key:")
		fmt.Println(npub)
		fmt.Println("Public key (In Hex Representation): ")
		fmt.Println(pk)
		fmt.Println()
		fmt.Println("Please keep the secret key secret.")
		fmt.Println("The key is used only at NostrP2P client as your identity.")
	},
}

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

	//serverCmd.Flags().StringVarP(
	//	&nickname,
	//	"Your nickname on nostrp2p (required)",
	//	"n",
	//	"",
	//	"Port to forward",
	//)
	//serverCmd.MarkFlagRequired("nickname")

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
	rootCmd.AddCommand(genkeyCmd)
}
