package session

import (
	"github.com/authgear/authgear-server/pkg/lib/api/model"
)

func NewInfo(attrs *Attrs, isAnonymous bool, isVerified bool) *model.SessionInfo {
	acr, _ := attrs.GetACR()
	amr, _ := attrs.GetAMR()
	return &model.SessionInfo{
		IsValid:       true,
		UserID:        attrs.UserID,
		UserAnonymous: isAnonymous,
		UserVerified:  isVerified,
		SessionACR:    acr,
		SessionAMR:    amr,
	}
}
