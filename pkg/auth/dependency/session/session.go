package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type IDPSession struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`

	CreatedAt time.Time   `json:"created_at"`
	Attrs     authn.Attrs `json:"attrs"`

	AccessInfo auth.AccessInfo `json:"access_info"`

	TokenHash string `json:"token_hash"`
}

var _ auth.AuthSession = &IDPSession{}

func (s *IDPSession) SessionID() string              { return s.ID }
func (s *IDPSession) SessionType() authn.SessionType { return auth.SessionTypeIdentityProvider }

func (s *IDPSession) AuthnAttrs() *authn.Attrs {
	return &s.Attrs
}

func (s *IDPSession) GetCreatedAt() time.Time         { return s.CreatedAt }
func (s *IDPSession) GetClientID() string             { return "" }
func (s *IDPSession) GetAccessInfo() *auth.AccessInfo { return &s.AccessInfo }

func (s *IDPSession) ToAPIModel() *model.Session {
	ua := model.ParseUserAgent(s.AccessInfo.LastAccess.UserAgent)
	ua.DeviceName = s.AccessInfo.LastAccess.Extra.DeviceName()
	return &model.Session{
		ID: s.ID,

		ACR:              s.Attrs.ACR,
		AMR:              s.Attrs.AMR,
		CreatedAt:        s.CreatedAt,
		LastAccessedAt:   s.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      s.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: s.AccessInfo.LastAccess.RemoteIP,
		UserAgent:        ua,
	}
}
