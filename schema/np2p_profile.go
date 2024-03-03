package schema

import "time"

type Np2pProfile struct {
	Pubkey64bit uint64
	Name        string
	About       string
	Picture     string
	UpdatedAt   int64 // unix timestamp in seconds
}

func NewNp2pProfile(pubKey64bit uint64, name string, about string, picture string) *Np2pProfile {
	return &Np2pProfile{
		Pubkey64bit: pubKey64bit,
		Name:        name,
		About:       about,
		Picture:     picture,
		UpdatedAt:   time.Now().Unix(),
	}
}
