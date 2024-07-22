package idpsession

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/geoip"
)

type IDPSession struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`

	// CreatedAt is the timestamp that the user was initially authenticated at.
	CreatedAt time.Time `json:"created_at"`
	// Authenticated is the timestamp that the user was authenticated at.
	// It is equal to CreatedAt if the user has not reauthenticated at all.
	AuthenticatedAt time.Time     `json:"authenticated_at"`
	Attrs           session.Attrs `json:"attrs"`

	AccessInfo access.Info `json:"access_info"`

	TokenHash string `json:"token_hash"`
}

var _ session.ResolvedSession = &IDPSession{}
var _ session.ListableSession = &IDPSession{}

func (s *IDPSession) Session()         {}
func (s *IDPSession) ListableSession() {}

func (s *IDPSession) SessionID() string         { return s.ID }
func (s *IDPSession) SessionType() session.Type { return session.TypeIdentityProvider }

func (s *IDPSession) GetCreatedAt() time.Time                       { return s.CreatedAt }
func (s *IDPSession) GetAuthenticatedAt() time.Time                 { return s.AuthenticatedAt }
func (s *IDPSession) GetClientID() string                           { return "" }
func (s *IDPSession) GetAccessInfo() *access.Info                   { return &s.AccessInfo }
func (s *IDPSession) GetDeviceInfo() (map[string]interface{}, bool) { return nil, false }
func (s *IDPSession) GetUserID() string                             { return s.Attrs.UserID }
func (s *IDPSession) GetOIDCAMR() ([]string, bool)                  { return s.Attrs.GetAMR() }

func (s *IDPSession) ToAPIModel() *model.Session {
	ua := model.ParseUserAgent(s.AccessInfo.LastAccess.UserAgent)
	amr, _ := s.Attrs.GetAMR()
	apiModel := &model.Session{
		Meta: model.Meta{
			ID:        s.ID,
			CreatedAt: s.CreatedAt,
			// TODO(session): Session Updated At should be the time user actively updates it.
			UpdatedAt: s.AccessInfo.LastAccess.Timestamp,
		},
		Type: model.SessionTypeIDP,

		AMR: amr,

		LastAccessedAt:   s.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      s.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: s.AccessInfo.LastAccess.RemoteIP,

		DisplayName: ua.Format(),
		UserAgent:   ua.Raw,
	}

	ipInfo, ok := geoip.DefaultDatabase.IPString(s.AccessInfo.LastAccess.RemoteIP)
	if ok {
		apiModel.LastAccessedByIPCountryCode = ipInfo.CountryCode
		apiModel.LastAccessedByIPEnglishCountryName = ipInfo.EnglishCountryName
	}

	return apiModel
}

func (s *IDPSession) GetAuthenticationInfo() authenticationinfo.T {
	amr, _ := s.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:          s.GetUserID(),
		AuthenticatedAt: s.GetAuthenticatedAt(),
		AMR:             amr,
	}
}

func (s *IDPSession) GetAuthenticationInfoByThisSession() authenticationinfo.T {
	amr, _ := s.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:                     s.GetUserID(),
		AuthenticatedAt:            s.GetAuthenticatedAt(),
		AMR:                        amr,
		AuthenticatedBySessionType: string(s.SessionType()),
		AuthenticatedBySessionID:   s.SessionID(),
	}
}

func (s *IDPSession) SSOGroupIDPSessionID() string {
	return s.SessionID()
}

// IsSameSSOGroup returns true when the session argument
// - is the same idp session
// - is sso enabled offline grant that in the same sso group
func (s *IDPSession) IsSameSSOGroup(ss session.SessionBase) bool {
	if s.EqualSession(ss) {
		return true
	}
	if s.SSOGroupIDPSessionID() == "" {
		return false
	}
	return s.SSOGroupIDPSessionID() == ss.SSOGroupIDPSessionID()
}

func (s *IDPSession) EqualSession(ss session.SessionBase) bool {
	return s.SessionID() == ss.SessionID() && s.SessionType() == ss.SessionType()
}
