package glo_val

import (
	"github.com/ryogrid/nostrp2p/np2p_const"
	"github.com/ryogrid/nostrp2p/schema"
)

var SelfPubkey *[np2p_const.PubkeySize]byte // initialized at creation of Np2pPeer
var SelfPubkey64bit uint64                  // initialized at creation of Np2pPeer
// var Nickname *string                        // initialized at server launch
// var ProfileMyOwn *schema.Np2pProfile
var CurrentProfileEvt *schema.Np2pEvent
var CurrentFollowListEvt *schema.Np2pEvent

var IsEnabledSSL bool = false

var DenyWriteMode = false

// 'DebugMode' global variable is defined on np2p_util.go ....
