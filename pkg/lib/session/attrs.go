package session

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
)

type Attrs struct {
	UserID string                          `json:"user_id"`
	Claims map[model.ClaimName]interface{} `json:"claims"`
}

func NewAttrs(userID string) *Attrs {
	return &Attrs{
		UserID: userID,
		Claims: map[model.ClaimName]interface{}{},
	}
}

func NewAttrsFromAuthenticationInfo(info authenticationinfo.T) *Attrs {
	attrs := NewAttrs(info.UserID)
	attrs.SetAMR(info.AMR)
	return attrs
}

func (a *Attrs) GetAMR() ([]string, bool) {
	amr, ok := a.Claims[model.ClaimAMR].([]string)
	return amr, ok
}

func (a *Attrs) SetAMR(value []string) {
	if len(value) > 0 {
		a.Claims[model.ClaimAMR] = value
	} else {
		delete(a.Claims, model.ClaimAMR)
	}
}
