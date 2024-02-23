package schema

import "time"

type BuzzProfile struct {
	Pubkey64bit uint64
	Name        string
	About       string
	Picture     string
	UpdatedAt   int64 // unix timestamp in seconds
}

func NewBuzzProfile(pubKey64bit uint64, name string, about string, picture string) *BuzzProfile {
	return &BuzzProfile{
		Pubkey64bit: pubKey64bit,
		Name:        name,
		About:       about,
		Picture:     picture,
		UpdatedAt:   time.Now().Unix(),
	}
}
