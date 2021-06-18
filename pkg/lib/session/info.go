package session

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

func NewInfo(attrs *Attrs, isAnonymous bool, isVerified bool) *model.SessionInfo {
	amr, _ := attrs.GetAMR()
	return &model.SessionInfo{
		IsValid:       true,
		UserID:        attrs.UserID,
		UserAnonymous: isAnonymous,
		UserVerified:  isVerified,
		SessionAMR:    amr,
	}
}
