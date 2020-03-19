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

	Attrs authn.Attrs `json:"attrs"`

	InitialAccess authn.AccessEvent `json:"initial_access"`
	LastAccess    authn.AccessEvent `json:"last_access"`

	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`

	TokenHash string `json:"token_hash"`
}

var _ auth.AuthSession = &IDPSession{}

func (s *IDPSession) SessionID() string              { return s.ID }
func (s *IDPSession) SessionType() authn.SessionType { return auth.SessionTypeIdentityProvider }

func (s *IDPSession) AuthnAttrs() *authn.Attrs {
	return &s.Attrs
}

func (s *IDPSession) ToAPIModel() *model.Session {
	ua := model.ParseUserAgent(s.LastAccess.UserAgent)
	ua.DeviceName = s.LastAccess.Extra.DeviceName()
	return &model.Session{
		ID: s.ID,

		IdentityID:        s.Attrs.PrincipalID,
		IdentityType:      string(s.Attrs.PrincipalType),
		IdentityUpdatedAt: s.Attrs.PrincipalUpdatedAt,

		AuthenticatorID:         s.Attrs.AuthenticatorID,
		AuthenticatorType:       string(s.Attrs.AuthenticatorType),
		AuthenticatorOOBChannel: string(s.Attrs.AuthenticatorOOBChannel),
		AuthenticatorUpdatedAt:  s.Attrs.AuthenticatorUpdatedAt,
		CreatedAt:               s.CreatedAt,
		LastAccessedAt:          s.AccessedAt,
		CreatedByIP:             s.InitialAccess.Remote.IP(),
		LastAccessedByIP:        s.LastAccess.Remote.IP(),
		UserAgent:               ua,
	}
}

type CreateReason string

const (
	CreateReasonSignup = "signup"
	CreateReasonLogin  = "login"
)
