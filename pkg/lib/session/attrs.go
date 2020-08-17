package session

import "github.com/authgear/authgear-server/pkg/lib/authn"

type Attrs struct {
	UserID string                          `json:"user_id"`
	Claims map[authn.ClaimName]interface{} `json:"claims"`
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
