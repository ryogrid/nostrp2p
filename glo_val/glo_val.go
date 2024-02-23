package glo_val

import "github.com/ryogrid/buzzoon/schema"

var SelfPubkey *[32]byte   // initialized at creation of BuzzPeer
var SelfPubkey64bit uint64 // initialized at creation of BuzzPeer
var Nickname *string       // initialized at server launch
var ProfileMyOwn *schema.BuzzProfile
