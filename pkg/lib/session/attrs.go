package session

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
)

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

func NewAttrsFromAuthenticationInfo(info authenticationinfo.T) *Attrs {
	attrs := NewAttrs(info.UserID)
	attrs.SetAMR(info.AMR)
	return attrs
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
