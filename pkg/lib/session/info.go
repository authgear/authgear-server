package session

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

func NewInfo(s Session, isAnonymous bool, isVerified bool) *model.SessionInfo {
	amr, _ := s.GetOIDCAMR()
	userID := s.GetUserID()
	return &model.SessionInfo{
		IsValid:         true,
		UserID:          userID,
		UserAnonymous:   isAnonymous,
		UserVerified:    isVerified,
		SessionAMR:      amr,
		AuthenticatedAt: s.GetAuthenticatedAt(),
	}
}
