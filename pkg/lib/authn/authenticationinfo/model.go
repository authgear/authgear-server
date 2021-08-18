package authenticationinfo

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type T struct {
	UserID string `json:"user_id,omitempty"`
	// AMR is authentication means used in the authentication.
	// On Android, we cannot tell the exact biometric means used in the authentication.
	// Therefore, we cannot reliably populate AMR.
	//
	// From RFC8176, the AMR values "swk" and "user" may apply.
	// See https://developer.android.com/reference/androidx/biometric/BiometricPrompt#AUTHENTICATION_RESULT_TYPE_BIOMETRIC
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
