package session

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

func NewInfo(s Session, isAnonymous bool, isVerified bool) *model.SessionInfo {
	info := s.GetAuthenticationInfo()
	return &model.SessionInfo{
		IsValid:         true,
		UserID:          info.UserID,
		UserAnonymous:   isAnonymous,
		UserVerified:    isVerified,
		SessionAMR:      info.AMR,
		AuthenticatedAt: info.AuthenticatedAt,
	}
}
