package cmd

import (
	"fmt"
	"github.com/ryogrid/buzzoon/buz_util"
	"github.com/spf13/cobra"
	"os"
)

var listenAddrPort = "127.0.0.1:20000"
var bootPeerAddrPort = ""
var publicKey = ""
var nickname = ""
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
		if debug {
			buz_util.DebugMode = true
		}

		peers := &buz_util.Stringset{}
		peers.Set("xxxxxxx:yyyyy")

		//peer, err := overlay.NewOverlayPeer(selfPeerId, &forwardAddress, int(listenPort+1000), peers, false)
		//if err != nil {
		//	log.Fatalln(err)
		//}

		// TODO: need to implement and create server instance (cmd.go)

		// TODO: need to implemnt and create temporal post request receiver I/f manager (cmd.go)

		//s := server.New(....)

		buz_util.OSInterrupt()
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
		&debug,
		"debug",
		"d",
		false,
		"If true, debug log is output to stderr (default: false)",
	)

	rootCmd.AddCommand(serverCmd)
}
