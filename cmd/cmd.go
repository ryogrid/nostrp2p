package cmd

import (
	"fmt"
	"github.com/ryogrid/buzzoon/api_server"
	"github.com/ryogrid/buzzoon/buzz_util"
	"github.com/ryogrid/buzzoon/core"
	"github.com/spf13/cobra"
	"github.com/weaveworks/mesh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
)

var listenAddrPort = "127.0.0.1:20000"
var bootPeerAddrPort = ""
var publicKey = ""
var nickname = ""
var writable = true
var debug = false

var rootCmd = &cobra.Command{
	Use: "buzzoon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("buzzoon v0.0.1")
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Startup server.",
	Run: func(cmd *cobra.Command, args []string) {
		if !writable {
			buzz_util.DenyWriteMode = true
		}
		if debug {
			buzz_util.DebugMode = true
		}

		peers := &buzz_util.Stringset{}
		if bootPeerAddrPort != "" {
			peers.Set(bootPeerAddrPort)
		}

		logger := log.New(os.Stderr, nickname+"> ", log.LstdFlags)

		host, portStr, err := net.SplitHostPort(listenAddrPort)
		if err != nil {
			logger.Fatalf("mesh address: %s: %v", listenAddrPort, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Fatalf("mesh address: %s: %v", listenAddrPort, err)
		}

		// TODO: need to use big int (cmd.go)
		name, err := strconv.ParseUint(publicKey, 16, 64)
		if err != nil {
			logger.Fatalf("public key: %s: %v", listenAddrPort, err)
		}
		// TODO: need to print boot message with hex public key string

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

		peer := core.NewPeer(mesh.PeerName(name), &nickname, logger)
		gossip, err := router.NewGossip("buzzoon", peer)
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

		apiServ := api_server.NewApiServer(peer)
		go apiServ.LaunchAPIServer(host + ":" + strconv.Itoa(port+1))
		// TODO: need to implement classes handle message sending and receiving (cmd.go)

		// TODO: need to implemnt and create temporal post request receiver I/f manager (cmd.go)
		/*
			if name == 3 {
				time.Sleep(5 * time.Second)
				buzz_util.BuzzDbgPrintln("send hello buzzon")
				event := schema.BuzzEvent{
					Id:         0,
					Pubkey:     [32]byte{},
					Created_at: 0,
					Kind:       0,
					Tags:       nil,
					Content:    "hello buzzon",
					Sig:        [64]byte{},
				}
				events := []*schema.BuzzEvent{&event}
				//peer.MessageMan.SendMsgUnicast(1, &schema.BuzzPacket{events, nil, nil})
				//peer.MessageMan.SendMsgUnicast(2, &schema.BuzzPacket{events, nil, nil})
				peer.MessageMan.SendMsgBroadcast(&schema.BuzzPacket{events, nil, nil})
				peer.MessageMan.SendMsgBroadcast(&schema.BuzzPacket{events, nil, nil})
			}
		*/
		buzz_util.OSInterrupt()
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
		"Address and port of a server which already joined buzzoon network (optional)",
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
		"Your nickname on buzzoon (required)",
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

	rootCmd.AddCommand(serverCmd)
}
