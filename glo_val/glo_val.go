package glo_val

import (
	"github.com/ryogrid/buzzoon/buzz_const"
	"github.com/ryogrid/buzzoon/schema"
)

var SelfPubkey *[buzz_const.PubkeySize]byte // initialized at creation of BuzzPeer
var SelfPubkey64bit uint64                  // initialized at creation of BuzzPeer
var Nickname *string                        // initialized at server launch
var ProfileMyOwn *schema.BuzzProfile
