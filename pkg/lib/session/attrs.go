package session

import "github.com/authgear/authgear-server/pkg/lib/authn"

type Attrs struct {
	UserID string                          `json:"user_id"`
	Claims map[authn.ClaimName]interface{} `json:"claims"`
}

func NewAttrs(userID string) *Attrs {
	return &Attrs{
		UserID: userID,
		Claims: map[authn.ClaimName]interface{}{},
	}
}

// NewBiometricAttrs is the same as NewAttrs.
// On Android, we cannot tell the exact biometric means used in the authentication.
// Therefore, we cannot reliably populate AMR and ACR.
//
// From RFC8176, the AMR values "swk" and "user" may apply.
//
// See https://developer.android.com/reference/androidx/biometric/BiometricPrompt#AUTHENTICATION_RESULT_TYPE_BIOMETRIC
func NewBiometricAttrs(userID string) *Attrs {
	return NewAttrs(userID)
}

// NewAnonymousAttrs is the same as NewAttrs.
//
// From RFC8176, the AMR value "swk" may apply.
func NewAnonymousAttrs(userID string) *Attrs {
	return NewAttrs(userID)
}

func (a *Attrs) GetACR() (string, bool) {
	acr, ok := a.Claims[authn.ClaimACR].(string)
	return acr, ok
}

func (a *Attrs) SetACR(value string) {
	if len(value) > 0 {
		a.Claims[authn.ClaimACR] = value
	} else {
		delete(a.Claims, authn.ClaimACR)
	}
}

func (a *Attrs) GetAMR() ([]string, bool) {
	amr, ok := a.Claims[authn.ClaimAMR].([]string)
	return amr, ok
}

func (a *Attrs) SetAMR(value []string) {
	if len(value) > 0 {
		a.Claims[authn.ClaimAMR] = value
	} else {
		delete(a.Claims, authn.ClaimAMR)
	}
}
