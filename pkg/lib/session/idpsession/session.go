package idpsession

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

type IDPSession struct {
	ID     string                 `json:"id"`
	Labels map[string]interface{} `json:"labels"`
	AppID  string                 `json:"app_id"`

	CreatedAt time.Time     `json:"created_at"`
	Attrs     session.Attrs `json:"attrs"`

	AccessInfo access.Info `json:"access_info"`

	TokenHash string `json:"token_hash"`
}

func (s *IDPSession) SessionID() string            { return s.ID }
func (s *IDPSession) SessionType() session.Type    { return session.TypeIdentityProvider }
func (s *IDPSession) SessionAttrs() *session.Attrs { return &s.Attrs }

func (s *IDPSession) GetCreatedAt() time.Time     { return s.CreatedAt }
func (s *IDPSession) GetClientID() string         { return "" }
func (s *IDPSession) GetAccessInfo() *access.Info { return &s.AccessInfo }

func (s *IDPSession) ToAPIModel() *model.Session {
	ua := model.ParseUserAgent(s.AccessInfo.LastAccess.UserAgent)
	acr, _ := s.Attrs.GetACR()
	amr, _ := s.Attrs.GetAMR()
	return &model.Session{
		Meta: model.Meta{
			ID:        s.ID,
			CreatedAt: s.CreatedAt,
			// TODO(session): Session Updated At should be the time user actively updates it.
			UpdatedAt: s.AccessInfo.LastAccess.Timestamp,
		},
		Type: model.SessionTypeIDP,

		ACR: acr,
		AMR: amr,

		LastAccessedAt:   s.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      s.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: s.AccessInfo.LastAccess.RemoteIP,
		UserAgent:        ua,
	}
}
