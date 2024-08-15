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
	// ShouldFireAuthenticatedEventWhenIssueOfflineGrant indicates we should fire authenticated event during code exchange
	// This value will be filled in during interaction / workflow / authentication flow
	ShouldFireAuthenticatedEventWhenIssueOfflineGrant bool `json:"should_fire_authenticated_event_when_issue_offline_grant,omitempty"`

	// AuthenticatedBySessionType and AuthenticatedBySessionID
	// means this authentication is done by an existing session.
	AuthenticatedBySessionType string
	AuthenticatedBySessionID   string
}

type Entry struct {
	ID             string `json:"id,omitempty"`
	T              T      `json:"t,omitempty"`
	OAuthSessionID string `json:"oauth_session_id,omitempty"`
	SAMLSessionID  string `json:"saml_session_id,omitempty"`
}

func NewEntry(t T, oauthSessionID string, samlSessionID string) *Entry {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return &Entry{
		ID:             id,
		T:              t,
		OAuthSessionID: oauthSessionID,
		SAMLSessionID:  samlSessionID,
	}
}
