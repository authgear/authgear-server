package authenticationinfo

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type T struct {
	UserID          string    `json:"user_id,omitempty"`
	AMR             []string  `json:"amr,omitempty"`
	AuthenticatedAt time.Time `json:"authenticated_at,omitempty"`
}

type Entry struct {
	ID string `json:"id,omitempty"`
	T  T      `json:"t,omitempty"`
}

func NewEntry(t T) *Entry {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return &Entry{
		ID: id,
		T:  t,
	}
}
